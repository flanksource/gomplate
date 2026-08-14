package grammar

import (
	"fmt"
	"sort"
	"strings"
)

// render turns a parsed rule body into JS regex source. Referenced rules are
// inlined from rules; stack carries the inlining chain so a cyclic grammar
// fails loudly instead of recursing forever.
func render(n node, rules map[string]string, stack []string) (string, error) {
	// An alternation of single characters and ranges is far more useful as one
	// character class: it reads better and it is the only form `~` can negate.
	// A lone literal stays as-is -- `'='` is `=`, not `[=]`.
	switch n.(type) {
	case *altNode, *rangeNode:
		if set, ok := charSet(n, rules, stack); ok {
			return "[" + set + "]", nil
		}
	}

	switch t := n.(type) {
	case *litNode:
		return escapeLiteral(t.text), nil

	case *rangeNode:
		return "[" + escapeClass(t.lo) + "-" + escapeClass(t.hi) + "]", nil

	case *anyNode:
		return `[\s\S]`, nil

	case *altNode:
		alts := longestPrefixFirst(t.alts, rules, stack)
		parts := make([]string, 0, len(alts))
		for _, a := range alts {
			s, err := render(a, rules, stack)
			if err != nil {
				return "", err
			}
			parts = append(parts, s)
		}
		return "(?:" + strings.Join(parts, "|") + ")", nil

	case *seqNode:
		var b strings.Builder
		for _, it := range t.items {
			s, err := render(it, rules, stack)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
		}
		return b.String(), nil

	case *repeatNode:
		inner, err := render(t.item, rules, stack)
		if err != nil {
			return "", err
		}
		if needsGroupBeforeQuantifier(inner) {
			inner = "(?:" + inner + ")"
		}
		return inner + t.op, nil

	case *notNode:
		set, ok := charSet(t.item, rules, stack)
		if !ok {
			return "", fmt.Errorf("`~` can only negate a set of single characters")
		}
		return "[^" + set + "]", nil

	case *refNode:
		body, ok := rules[t.name]
		if !ok {
			return "", fmt.Errorf("unknown rule reference %q", t.name)
		}
		for _, seen := range stack {
			if seen == t.name {
				return "", fmt.Errorf("cycle in rule references: %s -> %s", strings.Join(stack, " -> "), t.name)
			}
		}
		sub, err := parseBody(body)
		if err != nil {
			return "", fmt.Errorf("rule %s: %w", t.name, err)
		}
		return render(sub, rules, append(stack, t.name))

	default:
		return "", fmt.Errorf("unsupported node %T", n)
	}
}

// longestPrefixFirst reorders alternatives so the one with the longest
// fixed-width leading segment is tried first.
//
// ANTLR picks the longest match among ambiguous alternatives; JS regex picks the
// first that matches. Left in declaration order, CEL's STRING rule would tokenize
// `"""x"""` as an empty string followed by junk, because `'"' ... '"'` is
// declared before `'"""' ... '"""'`.
func longestPrefixFirst(alts []node, rules map[string]string, stack []string) []node {
	out := make([]node, len(alts))
	copy(out, alts)
	sort.SliceStable(out, func(i, j int) bool {
		return fixedPrefixLen(out[i], rules, stack) > fixedPrefixLen(out[j], rules, stack)
	})
	return out
}

// fixedPrefixLen counts the characters an alternative is guaranteed to consume
// before its first variable-width element. A character class counts as one.
func fixedPrefixLen(n node, rules map[string]string, stack []string) int {
	switch t := n.(type) {
	case *litNode:
		return len([]rune(t.text))
	case *rangeNode, *anyNode, *notNode:
		return 1
	case *altNode:
		if _, ok := charSet(t, rules, stack); ok {
			return 1
		}
		return 0 // a genuine branch contributes no guaranteed prefix
	case *seqNode:
		total := 0
		for _, it := range t.items {
			n := fixedPrefixLen(it, rules, stack)
			total += n
			if n == 0 {
				break // variable width from here on
			}
		}
		return total
	case *refNode:
		body, ok := rules[t.name]
		if !ok {
			return 0
		}
		for _, seen := range stack {
			if seen == t.name {
				return 0
			}
		}
		sub, err := parseBody(body)
		if err != nil {
			return 0
		}
		return fixedPrefixLen(sub, rules, append(stack, t.name))
	default:
		return 0 // repeatNode and anything else is variable width
	}
}

func parseBody(body string) (node, error) {
	p := &parser{src: []rune(body)}
	n, err := p.parseAlt()
	if err != nil {
		return nil, err
	}
	if p.peek() != 0 {
		return nil, fmt.Errorf("unexpected %q at offset %d", p.peek(), p.pos)
	}
	return n, nil
}

// charSet renders n as the body of a character class, reporting false when n is
// anything other than single characters and ranges. References are resolved so
// that `LETTER | DIGIT | '_'` collapses to `A-Za-z0-9_` rather than staying an
// alternation of three separate classes.
func charSet(n node, rules map[string]string, stack []string) (string, bool) {
	switch t := n.(type) {
	case *litNode:
		if len([]rune(t.text)) != 1 {
			return "", false
		}
		return escapeClass([]rune(t.text)[0]), true
	case *rangeNode:
		return escapeClass(t.lo) + "-" + escapeClass(t.hi), true
	case *altNode:
		var b strings.Builder
		for _, a := range t.alts {
			s, ok := charSet(a, rules, stack)
			if !ok {
				return "", false
			}
			b.WriteString(s)
		}
		return b.String(), true
	case *refNode:
		body, ok := rules[t.name]
		if !ok {
			return "", false
		}
		for _, seen := range stack {
			if seen == t.name {
				return "", false
			}
		}
		sub, err := parseBody(body)
		if err != nil {
			return "", false
		}
		return charSet(sub, rules, append(stack, t.name))
	default:
		return "", false
	}
}

// needsGroupBeforeQuantifier reports whether src must be wrapped before a
// quantifier binds to it. A single char, an escape, a character class or an
// existing group already binds as a unit.
func needsGroupBeforeQuantifier(src string) bool {
	switch {
	case len(src) == 1:
		return false
	case strings.HasPrefix(src, "(?:") && strings.HasSuffix(src, ")") && balanced(src):
		return false
	case strings.HasPrefix(src, "[") && strings.HasSuffix(src, "]") && classIsWhole(src):
		return false
	case len(src) == 2 && src[0] == '\\':
		return false
	default:
		return true
	}
}

// balanced reports whether the outermost `(` of src closes at its final `)`.
func balanced(src string) bool {
	depth := 0
	for i, r := range src {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 && i != len(src)-1 {
				return false
			}
		}
	}
	return depth == 0
}

// classIsWhole reports whether src is a single character class, i.e. its
// opening `[` closes only at the final `]`.
func classIsWhole(src string) bool {
	for i := 1; i < len(src)-1; i++ {
		if src[i] == '\\' {
			i++
			continue
		}
		if src[i] == ']' {
			return false
		}
	}
	return true
}

// regexMeta are the characters that must be escaped outside a character class.
// `/` is deliberately absent: it is only special inside a JS regex *literal*,
// and Monarch patterns are strings handed to `new RegExp`.
const regexMeta = `\.+*?()|[]{}^$`

func escapeLiteral(s string) string {
	var b strings.Builder
	for _, r := range s {
		b.WriteString(escapeRune(r, regexMeta))
	}
	return b.String()
}

// escapeClass escapes a rune for use inside a character class, where the only
// special characters are `\`, `]`, `^` and `-`.
func escapeClass(r rune) string { return escapeRune(r, `\]^-`) }

func escapeRune(r rune, meta string) string {
	switch r {
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\t':
		return `\t`
	case '\f':
		return `\f`
	case '\b':
		return `\b`
	}
	if r < 0x20 || r == 0x7f {
		return fmt.Sprintf(`\u%04X`, r)
	}
	if strings.ContainsRune(meta, r) {
		return `\` + string(r)
	}
	return string(r)
}
