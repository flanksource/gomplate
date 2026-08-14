// Package grammar extracts lexical vocabulary from the grammars of the parsers
// gomplate actually runs, so the editor tokenizers stay in step with them
// instead of being transcribed by hand.
package grammar

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// node is a parsed ANTLR rule body. Every node renders to JS regex source.
type node interface{ isNode() }

type (
	// altNode is `a | b | c`.
	altNode struct{ alts []node }
	// seqNode is juxtaposition: `a b c`.
	seqNode struct{ items []node }
	// repeatNode is a suffixed element: `a+`, `a*?`, `a?`.
	repeatNode struct {
		item node
		op   string
	}
	// litNode is a quoted literal: `'0x'`.
	litNode struct{ text string }
	// rangeNode is `'a'..'z'`.
	rangeNode struct{ lo, hi rune }
	// notNode is `~(...)`.
	notNode struct{ item node }
	// refNode is a reference to another lexer rule or fragment.
	refNode struct{ name string }
	// anyNode is `.`.
	anyNode struct{}
)

func (*altNode) isNode()    {}
func (*seqNode) isNode()    {}
func (*repeatNode) isNode() {}
func (*litNode) isNode()    {}
func (*rangeNode) isNode()  {}
func (*notNode) isNode()    {}
func (*refNode) isNode()    {}
func (*anyNode) isNode()    {}

// TranslateRuleBody converts one ANTLR lexer rule body into JS regex source,
// inlining any referenced rule from rules (transitively).
func TranslateRuleBody(body string, rules map[string]string) (string, error) {
	p := &parser{src: []rune(body)}
	root, err := p.parseAlt()
	if err != nil {
		return "", err
	}
	if p.peek() != 0 {
		return "", fmt.Errorf("unexpected %q at offset %d in %q", p.peek(), p.pos, body)
	}
	return render(root, rules, nil)
}

type parser struct {
	src []rune
	pos int
}

func (p *parser) peek() rune {
	p.skipSpace()
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) skipSpace() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t' || p.src[p.pos] == '\n' || p.src[p.pos] == '\r') {
		p.pos++
	}
}

func (p *parser) parseAlt() (node, error) {
	var alts []node
	for {
		s, err := p.parseSeq()
		if err != nil {
			return nil, err
		}
		alts = append(alts, s)
		if p.peek() != '|' {
			break
		}
		p.pos++
	}
	if len(alts) == 1 {
		return alts[0], nil
	}
	return &altNode{alts: alts}, nil
}

func (p *parser) parseSeq() (node, error) {
	var items []node
	for {
		c := p.peek()
		if c == 0 || c == '|' || c == ')' {
			break
		}
		e, err := p.parseElem()
		if err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("empty alternative at offset %d", p.pos)
	}
	if len(items) == 1 {
		return items[0], nil
	}
	return &seqNode{items: items}, nil
}

func (p *parser) parseElem() (node, error) {
	atom, err := p.parseAtom()
	if err != nil {
		return nil, err
	}
	switch p.peek() {
	case '+', '*', '?':
		op := string(p.src[p.pos])
		p.pos++
		if p.pos < len(p.src) && p.src[p.pos] == '?' && op != "?" {
			op += "?"
			p.pos++
		}
		return &repeatNode{item: atom, op: op}, nil
	}
	return atom, nil
}

func (p *parser) parseAtom() (node, error) {
	switch c := p.peek(); {
	case c == '\'':
		lit, err := p.parseLiteral()
		if err != nil {
			return nil, err
		}
		if !p.hasPrefix("..") {
			return &litNode{text: lit}, nil
		}
		p.pos += 2
		hi, err := p.parseLiteral()
		if err != nil {
			return nil, err
		}
		lo, _ := utf8.DecodeRuneInString(lit)
		hiR, _ := utf8.DecodeRuneInString(hi)
		return &rangeNode{lo: lo, hi: hiR}, nil
	case c == '(':
		p.pos++
		inner, err := p.parseAlt()
		if err != nil {
			return nil, err
		}
		if p.peek() != ')' {
			return nil, fmt.Errorf("unclosed group at offset %d", p.pos)
		}
		p.pos++
		return inner, nil
	case c == '~':
		p.pos++
		inner, err := p.parseAtom()
		if err != nil {
			return nil, err
		}
		return &notNode{item: inner}, nil
	case c == '.':
		p.pos++
		return &anyNode{}, nil
	case isIdentStart(c):
		start := p.pos
		for p.pos < len(p.src) && isIdentPart(p.src[p.pos]) {
			p.pos++
		}
		return &refNode{name: string(p.src[start:p.pos])}, nil
	default:
		return nil, fmt.Errorf("unexpected %q at offset %d", c, p.pos)
	}
}

func (p *parser) hasPrefix(s string) bool {
	p.skipSpace()
	return strings.HasPrefix(string(p.src[p.pos:]), s)
}

// parseLiteral consumes a single-quoted ANTLR literal and unescapes it.
func (p *parser) parseLiteral() (string, error) {
	p.skipSpace()
	if p.pos >= len(p.src) || p.src[p.pos] != '\'' {
		return "", fmt.Errorf("expected a quoted literal at offset %d", p.pos)
	}
	p.pos++
	var b strings.Builder
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch c {
		case '\'':
			p.pos++
			return b.String(), nil
		case '\\':
			p.pos++
			if p.pos >= len(p.src) {
				return "", fmt.Errorf("dangling escape at offset %d", p.pos)
			}
			r, err := unescape(p.src, &p.pos)
			if err != nil {
				return "", err
			}
			b.WriteRune(r)
		default:
			b.WriteRune(c)
			p.pos++
		}
	}
	return "", fmt.Errorf("unterminated literal at offset %d", p.pos)
}

// unescape decodes the escape following a backslash, advancing *pos past it.
func unescape(src []rune, pos *int) (rune, error) {
	c := src[*pos]
	*pos++
	switch c {
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	case 'f':
		return '\f', nil
	case 'b':
		return '\b', nil
	case '\\', '\'', '"':
		return c, nil
	case 'u':
		// ANTLR writes \uXXXX or \u{XXXX}.
		digits := ""
		if *pos < len(src) && src[*pos] == '{' {
			*pos++
			for *pos < len(src) && src[*pos] != '}' {
				digits += string(src[*pos])
				*pos++
			}
			*pos++ // closing brace
		} else {
			for i := 0; i < 4 && *pos < len(src); i++ {
				digits += string(src[*pos])
				*pos++
			}
		}
		var r rune
		if _, err := fmt.Sscanf(digits, "%x", &r); err != nil {
			return 0, fmt.Errorf("bad unicode escape \\u%s: %w", digits, err)
		}
		return r, nil
	default:
		return c, nil
	}
}

func isIdentStart(c rune) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentPart(c rune) bool { return isIdentStart(c) || (c >= '0' && c <= '9') }
