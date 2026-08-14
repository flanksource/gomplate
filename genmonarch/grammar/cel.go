package grammar

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// celGrammarSource is cel-go's own ANTLR grammar, vendored so the tokenizer is
// generated from the same lexical rules the parser enforces. cel_test.go asserts
// it stays byte-identical to the copy in the resolved cel-go module.
//
//go:embed CEL.g4
var celGrammarSource string

// CELGrammarSource returns the vendored grammar text.
func CELGrammarSource() string { return celGrammarSource }

// CELGrammar is the lexical vocabulary of CEL, extracted from CEL.g4.
type CELGrammar struct {
	// Operators are the punctuation tokens, ordered longest-first so a
	// tokenizer trying them in sequence never lets `<` shadow `<=`.
	Operators []string
	// Keywords are the literal word tokens: in, true, false, null.
	Keywords []string
	// Patterns maps a composite token rule (STRING, NUM_INT, IDENTIFIER, ...)
	// to JS regex source.
	Patterns map[string]string
}

// compositeRules are the token rules a tokenizer needs as regexes rather than
// as literal strings. Everything else in the lexer section is a single literal
// and lands in Operators or Keywords.
var compositeRules = []string{
	"WHITESPACE", "COMMENT",
	"NUM_FLOAT", "NUM_INT", "NUM_UINT",
	"STRING", "BYTES",
	"IDENTIFIER", "ESC_IDENTIFIER",
}

// ParseCEL extracts the lexical vocabulary from the vendored CEL.g4.
func ParseCEL() (*CELGrammar, error) {
	rules, order, fragments, err := parseANTLRRules(celGrammarSource)
	if err != nil {
		return nil, err
	}

	g := &CELGrammar{Patterns: map[string]string{}}
	for _, name := range order {
		// Fragments exist only to be inlined -- BACKSLASH is not an operator.
		if fragments[name] {
			continue
		}
		body := rules[name]
		if lit, ok := soleLiteral(body); ok {
			if isWord(lit) {
				g.Keywords = append(g.Keywords, lit)
			} else {
				g.Operators = append(g.Operators, lit)
			}
		}
	}

	for _, name := range compositeRules {
		body, ok := rules[name]
		if !ok {
			return nil, fmt.Errorf("CEL.g4 no longer defines the %s rule", name)
		}
		pattern, err := TranslateRuleBody(body, rules)
		if err != nil {
			return nil, fmt.Errorf("rule %s: %w", name, err)
		}
		g.Patterns[name] = pattern
	}

	// Longest first, then lexicographic so the output is stable across runs.
	sort.SliceStable(g.Operators, func(i, j int) bool {
		if len(g.Operators[i]) != len(g.Operators[j]) {
			return len(g.Operators[i]) > len(g.Operators[j])
		}
		return g.Operators[i] < g.Operators[j]
	})
	sort.Strings(g.Keywords)
	return g, nil
}

// parseANTLRRules splits an ANTLR grammar into `NAME -> body`, keeping only the
// lexer rules (an initial capital) and fragments. order preserves the order of
// declaration, which is the lexer's own precedence; fragments records which of
// them are inline-only.
func parseANTLRRules(src string) (rules map[string]string, order []string, fragments map[string]bool, err error) {
	rules = map[string]string{}
	fragments = map[string]bool{}

	for _, chunk := range splitTopLevel(stripComments(src), ';') {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		colon := indexTopLevel(chunk, ':')
		if colon < 0 {
			continue // options {...}, grammar header, etc.
		}
		name := strings.TrimSpace(chunk[:colon])
		isFragment := strings.HasPrefix(name, "fragment")
		if isFragment {
			name = strings.TrimSpace(strings.TrimPrefix(name, "fragment"))
		}
		if name == "" || !unicode.IsUpper(rune(name[0])) || strings.ContainsAny(name, " \t\n") {
			continue // parser rule, or not a rule at all
		}
		body := stripAction(strings.TrimSpace(chunk[colon+1:]))
		if body == "" {
			continue
		}
		rules[name] = body
		order = append(order, name)
		fragments[name] = isFragment
	}

	if len(rules) == 0 {
		return nil, nil, nil, fmt.Errorf("no lexer rules found; is CEL.g4 intact?")
	}
	return rules, order, fragments, nil
}

// stripAction removes a trailing ANTLR action such as `-> channel(HIDDEN)`.
func stripAction(body string) string {
	if i := indexTopLevel(body, '-'); i >= 0 && strings.HasPrefix(body[i:], "->") {
		return strings.TrimSpace(body[:i])
	}
	return body
}

// soleLiteral reports whether body is exactly one quoted literal, returning it.
func soleLiteral(body string) (string, bool) {
	p := &parser{src: []rune(body)}
	if p.peek() != '\'' {
		return "", false
	}
	lit, err := p.parseLiteral()
	if err != nil || p.peek() != 0 {
		return "", false
	}
	return lit, true
}

func isWord(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && r != '_' {
			return false
		}
	}
	return s != ""
}

// stripComments removes `//` and `/* */` comments that are not inside a quoted
// literal -- the COMMENT rule itself contains a literal `'//'`.
func stripComments(src string) string {
	var b strings.Builder
	runes := []rune(src)
	for i := 0; i < len(runes); i++ {
		switch {
		case runes[i] == '\'':
			j := scanLiteral(runes, i)
			b.WriteString(string(runes[i:j]))
			i = j - 1
		case runes[i] == '/' && i+1 < len(runes) && runes[i+1] == '/':
			for i < len(runes) && runes[i] != '\n' {
				i++
			}
			b.WriteRune('\n')
		case runes[i] == '/' && i+1 < len(runes) && runes[i+1] == '*':
			i += 2
			for i+1 < len(runes) && (runes[i] != '*' || runes[i+1] != '/') {
				i++
			}
			i++
		default:
			b.WriteRune(runes[i])
		}
	}
	return b.String()
}

// scanLiteral returns the index just past the quoted literal starting at i.
func scanLiteral(runes []rune, i int) int {
	j := i + 1
	for j < len(runes) {
		if runes[j] == '\\' {
			j += 2
			continue
		}
		if runes[j] == '\'' {
			return j + 1
		}
		j++
	}
	return len(runes)
}

// splitTopLevel splits on sep, ignoring separators inside quoted literals.
func splitTopLevel(src string, sep rune) []string {
	var out []string
	runes := []rune(src)
	start := 0
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\'' {
			i = scanLiteral(runes, i) - 1
			continue
		}
		if runes[i] == sep {
			out = append(out, string(runes[start:i]))
			start = i + 1
		}
	}
	return append(out, string(runes[start:]))
}

// indexTopLevel finds target outside any quoted literal, or -1.
func indexTopLevel(src string, target rune) int {
	runes := []rune(src)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\'' {
			i = scanLiteral(runes, i) - 1
			continue
		}
		if runes[i] == target {
			return i
		}
	}
	return -1
}
