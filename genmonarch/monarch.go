package genmonarch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Language is a Monaco Monarch language definition. It marshals to the shape
// monaco.languages.setMonarchTokensProvider expects.
type Language struct {
	ID           string    `json:"-"`
	DefaultToken string    `json:"defaultToken"`
	TokenPostfix string    `json:"tokenPostfix"`
	Start        string    `json:"start,omitempty"`
	Brackets     []Bracket `json:"brackets,omitempty"`
	// Attributes are the named word lists a rule refers to as `@name`. Order
	// within a list is irrelevant; the lists are sorted for a stable diff.
	Attributes map[string][]string `json:"-"`
	// Tokenizer holds the states. `root` is the entry state.
	Tokenizer *States `json:"tokenizer"`
}

// Bracket is a bracket pair Monaco should match and colour.
type Bracket struct {
	Open  string `json:"open"`
	Close string `json:"close"`
	Token string `json:"token"`
}

// MarshalJSON flattens Attributes alongside the fixed fields, which is how
// Monarch expects word lists to appear.
func (l Language) MarshalJSON() ([]byte, error) {
	type alias Language // avoid recursing into this method
	base, err := json.Marshal(alias(l))
	if err != nil {
		return nil, err
	}

	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(base, &fields); err != nil {
		return nil, err
	}
	for name, words := range l.Attributes {
		sorted := append([]string(nil), words...)
		sort.Strings(sorted)
		raw, err := json.Marshal(dedupe(sorted))
		if err != nil {
			return nil, err
		}
		if _, clash := fields[name]; clash {
			return nil, fmt.Errorf("attribute %q collides with a reserved Monarch field", name)
		}
		fields[name] = raw
	}
	return marshalOrdered(fields, orderedKeys(fields))
}

// States is an ordered set of tokenizer states. Both the state order and the
// rule order inside a state are significant: Monarch takes the first rule that
// matches, so a shorter pattern declared first silently shadows a longer one.
type States struct {
	names  []string
	states map[string][]Rule
}

// NewStates returns an empty, ordered state set.
func NewStates() *States {
	return &States{states: map[string][]Rule{}}
}

// Add appends a state. Adding the same name twice appends to it.
func (s *States) Add(name string, rules ...Rule) *States {
	if _, seen := s.states[name]; !seen {
		s.names = append(s.names, name)
	}
	s.states[name] = append(s.states[name], rules...)
	return s
}

// Names lists the states in declaration order.
func (s *States) Names() []string { return append([]string(nil), s.names...) }

// Rules returns the rules of a state.
func (s *States) Rules(name string) []Rule { return s.states[name] }

func (s *States) MarshalJSON() ([]byte, error) {
	fields := map[string]json.RawMessage{}
	for name, rules := range s.states {
		raw, err := json.Marshal(rules)
		if err != nil {
			return nil, fmt.Errorf("state %s: %w", name, err)
		}
		fields[name] = raw
	}
	return marshalOrdered(fields, s.names)
}

// Rule is one tokenizer rule: a pattern with an action, or an include of
// another state.
type Rule struct {
	// Regex is JS regex source, without delimiters.
	Regex string
	// Action fires when Regex matches.
	Action Action
	// Include names a state to splice in, e.g. `@whitespace`. When set, the
	// rest of the rule is ignored.
	Include string
}

func (r Rule) MarshalJSON() ([]byte, error) {
	if r.Include != "" {
		return json.Marshal(map[string]string{"include": r.Include})
	}
	action, err := r.Action.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return marshalTuple(r.Regex, action)
}

// Match builds a rule that emits a single token.
func Match(regex, token string) Rule {
	return Rule{Regex: regex, Action: Action{Token: token}}
}

// MatchGroups builds a rule that emits one token per capture group.
func MatchGroups(regex string, tokens ...string) Rule {
	return Rule{Regex: regex, Action: Action{Tokens: tokens}}
}

// Push builds a rule that emits a token and enters another state.
func Push(regex, token, next string) Rule {
	return Rule{Regex: regex, Action: Action{Token: token, Next: next}}
}

// Pop builds a rule that emits a token and returns to the previous state.
func Pop(regex, token string) Rule {
	return Rule{Regex: regex, Action: Action{Token: token, Next: "@pop"}}
}

// Include splices another state's rules in at this position.
func Include(state string) Rule { return Rule{Include: state} }

// Action is what a rule does when its pattern matches.
type Action struct {
	// Token is the single token class to emit.
	Token string
	// Tokens is one token class per capture group; mutually exclusive with Token.
	Tokens []string
	// Next is the state to enter: a state name, `@pop`, or `@push`.
	Next string
	// Cases selects between actions by testing the match against word lists.
	Cases *Cases
}

func (a Action) MarshalJSON() ([]byte, error) {
	switch {
	case a.Cases != nil:
		return json.Marshal(map[string]*Cases{"cases": a.Cases})
	case a.Tokens != nil && a.Next != "":
		return nil, fmt.Errorf("a grouped action cannot also switch state")
	case a.Tokens != nil:
		return json.Marshal(a.Tokens)
	case a.Next != "":
		return json.Marshal(map[string]string{"token": a.Token, "next": a.Next})
	default:
		return json.Marshal(a.Token)
	}
}

// Cases is an ordered set of guarded actions. Monarch evaluates the guards in
// order, so `@default` belongs last and a more specific guard must precede a
// broader one.
type Cases struct {
	guards  []string
	actions map[string]Action
}

// NewCases returns an empty, ordered case set.
func NewCases() *Cases { return &Cases{actions: map[string]Action{}} }

// When appends a guard, e.g. `@keywords` or `$1@namespaces`.
func (c *Cases) When(guard string, action Action) *Cases {
	if _, seen := c.actions[guard]; !seen {
		c.guards = append(c.guards, guard)
	}
	c.actions[guard] = action
	return c
}

// Token appends a guard emitting a single token class.
func (c *Cases) Token(guard, token string) *Cases {
	return c.When(guard, Action{Token: token})
}

// Groups appends a guard emitting one token class per capture group.
func (c *Cases) Groups(guard string, tokens ...string) *Cases {
	return c.When(guard, Action{Tokens: tokens})
}

// Default appends the fallback, which must be added last.
func (c *Cases) Default(action Action) *Cases { return c.When("@default", action) }

func (c *Cases) MarshalJSON() ([]byte, error) {
	fields := map[string]json.RawMessage{}
	for guard, action := range c.actions {
		raw, err := action.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("case %s: %w", guard, err)
		}
		fields[guard] = raw
	}
	return marshalOrdered(fields, c.guards)
}

// Configuration is a Monaco language configuration: the editor behaviours that
// are not tokenization.
type Configuration struct {
	Comments         *Comments   `json:"comments,omitempty"`
	Brackets         [][2]string `json:"brackets,omitempty"`
	AutoClosingPairs []Pair      `json:"autoClosingPairs,omitempty"`
	SurroundingPairs []Pair      `json:"surroundingPairs,omitempty"`
	// WordPattern decides what counts as one word for completion. Dotted names
	// such as `k8s.isHealthy` must match as a single word or completing them
	// inserts a duplicated namespace.
	WordPattern string `json:"wordPattern,omitempty"`
}

// Comments declares how comments are written.
type Comments struct {
	LineComment  string    `json:"lineComment,omitempty"`
	BlockComment [2]string `json:"blockComment,omitempty"`
}

// Pair is a pair of strings the editor closes or surrounds a selection with.
type Pair struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

// marshalOrdered writes a JSON object with its keys in the given order, so a
// regenerated file diffs cleanly and Monarch's order-sensitive constructs keep
// their meaning.
func marshalOrdered(fields map[string]json.RawMessage, order []string) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	for _, key := range order {
		raw, ok := fields[key]
		if !ok {
			continue
		}
		if !first {
			buf.WriteByte(',')
		}
		first = false
		encoded, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(encoded)
		buf.WriteByte(':')
		buf.Write(raw)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func marshalTuple(first string, rest ...json.RawMessage) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')
	encoded, err := json.Marshal(first)
	if err != nil {
		return nil, err
	}
	buf.Write(encoded)
	for _, raw := range rest {
		buf.WriteByte(',')
		buf.Write(raw)
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

// orderedKeys keeps the fixed Monarch fields first and the generated word lists
// after them, each group sorted, so the emitted file is stable.
func orderedKeys(fields map[string]json.RawMessage) []string {
	fixed := []string{"defaultToken", "tokenPostfix", "start", "brackets"}
	var rest []string
	for key := range fields {
		if !contains(fixed, key) && key != "tokenizer" {
			rest = append(rest, key)
		}
	}
	sort.Strings(rest)
	out := make([]string, 0, len(fields))
	for _, key := range fixed {
		if _, ok := fields[key]; ok {
			out = append(out, key)
		}
	}
	return append(append(out, rest...), "tokenizer")
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func dedupe(sorted []string) []string {
	out := sorted[:0]
	for i, s := range sorted {
		if i == 0 || s != sorted[i-1] {
			out = append(out, s)
		}
	}
	return out
}

// escapeRegexLiteral escapes a string for literal use inside a JS regex. Used
// for the configurable template delimiters, which are data, not patterns.
func escapeRegexLiteral(s string) string {
	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune(`\.+*?()|[]{}^$/`, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
