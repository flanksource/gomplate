package grammar

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("antlr rule bodies translate to JS regex", func() {
	// Each case is a real fragment or rule body from CEL.g4 (or a minimal
	// reduction of one), paired with the regex a Monarch tokenizer needs.
	// Alternations of single characters and ranges collapse into one character
	// class -- both to keep the emitted regex readable and because `~(...)`
	// negation is only expressible as a class.
	cases := []struct {
		name string
		body string
		want string
	}{
		{"literal", `'=='`, `==`},
		{"literal needing escape", `'.'`, `\.`},
		{"single-char alternation collapses", `'e' | 'E'`, `[eE]`},
		{"char range", `'0'..'9'`, `[0-9]`},
		{"range alternation collapses", `'A'..'Z' | 'a'..'z'`, `[A-Za-z]`},
		{"multi-char alternation cannot collapse", `'0x' | 'ab'`, `(?:0x|ab)`},
		{"one or more", `'0'..'9'+`, `[0-9]+`},
		{"optional group", `( '+' | '-' )?`, `[+\-]?`},
		{"negated char set", `~('\\'|'"'|'\n'|'\r')`, `[^\\"\n\r]`},
		{"any char non-greedy", `.*?`, `[\s\S]*?`},
		{"sequence", `'0x' HEXDIGIT+`, `0x[0-9a-fA-F]+`},
		{"redundant group is dropped", `('.' '0'..'9'+)`, `\.[0-9]+`},
		{"group is kept when a quantifier binds to it", `('.' '0'..'9'+)?`, `(?:\.[0-9]+)?`},
		{"quantifier binds directly to a class", `HEXDIGIT+`, `[0-9a-fA-F]+`},
	}

	for _, tc := range cases {
		ginkgo.It(tc.name, func() {
			got, err := TranslateRuleBody(tc.body, celFragmentsForTest())
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(tc.want))
		})
	}

	ginkgo.It("expands fragment references transitively", func() {
		// ESC_BYTE_SEQ : BACKSLASH ( 'x' | 'X' ) HEXDIGIT HEXDIGIT
		got, err := TranslateRuleBody(`ESC_BYTE_SEQ`, celFragmentsForTest())
		Expect(err).ToNot(HaveOccurred())
		Expect(got).To(Equal(`\\[xX][0-9a-fA-F][0-9a-fA-F]`))
	})

	ginkgo.It("fails loudly on an unknown reference rather than emitting a broken regex", func() {
		_, err := TranslateRuleBody(`NOT_A_RULE+`, map[string]string{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("NOT_A_RULE"))
	})

	ginkgo.It("fails loudly on a cyclic reference instead of recursing forever", func() {
		_, err := TranslateRuleBody(`A`, map[string]string{"A": `'x' B`, "B": `'y' A`})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("cycle"))
	})

	// ANTLR resolves an ambiguous alternation by longest match; JS regex takes
	// the first alternative that matches. Translating order-for-order silently
	// produces a tokenizer that stops early -- and an anchored test would not
	// catch it, because backtracking hides the bug when the whole input must
	// match. Monarch matches a prefix, so the order has to be corrected.
	ginkgo.It("orders alternatives so the longest fixed prefix is tried first", func() {
		got, err := TranslateRuleBody(`'"' 'a'* '"' | '"""' 'a'* '"""'`, map[string]string{})
		Expect(err).ToNot(HaveOccurred())
		Expect(got).To(Equal(`(?:"""a*"""|"a*")`))
	})

	ginkgo.It("prefers the hex branch over the bare-digit branch", func() {
		got, err := TranslateRuleBody(`'0'..'9'+ | '0x' '0'..'9'+`, map[string]string{})
		Expect(err).ToNot(HaveOccurred())
		Expect(got).To(Equal(`(?:0x[0-9]+|[0-9]+)`))
	})

	ginkgo.It("counts a leading character class as one character of prefix", func() {
		// `RAW '"""'` must outrank `RAW '"'` just as `'"""'` outranks `'"'`.
		got, err := TranslateRuleBody(`RAW '"' 'a'* '"' | RAW '"""' 'a'* '"""'`,
			map[string]string{"RAW": `'r' | 'R'`})
		Expect(err).ToNot(HaveOccurred())
		Expect(got).To(Equal(`(?:[rR]"""a*"""|[rR]"a*")`))
	})
})

// celFragmentsForTest is the subset of CEL.g4 fragments the cases above lean on.
func celFragmentsForTest() map[string]string {
	return map[string]string{
		"BACKSLASH":    `'\\'`,
		"HEXDIGIT":     `('0'..'9'|'a'..'f'|'A'..'F')`,
		"ESC_BYTE_SEQ": `BACKSLASH ( 'x' | 'X' ) HEXDIGIT HEXDIGIT`,
	}
}
