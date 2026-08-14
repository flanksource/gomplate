package grammar

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("text/template vocabulary read from the standard library", func() {
	var g *GoTemplateGrammar

	ginkgo.BeforeEach(func() {
		var err error
		g, err = ParseGoTemplate()
		Expect(err).ToNot(HaveOccurred())
	})

	ginkgo.It("recovers every action keyword from parse.key", func() {
		Expect(g.Keywords).To(ConsistOf(
			"block", "break", "continue", "define", "else", "end",
			"if", "range", "nil", "template", "with",
		))
	})

	ginkgo.It("treats `.` as a token rather than a keyword", func() {
		// parse.key carries "." as itemDot, but highlighting it as a keyword
		// would colour every field access.
		Expect(g.Keywords).ToNot(ContainElement("."))
	})

	ginkgo.It("recovers the builtin functions text/template supplies", func() {
		Expect(g.Builtins).To(ConsistOf(
			"and", "call", "html", "index", "slice", "js", "len", "not", "or",
			"print", "printf", "println", "urlquery",
			"eq", "ge", "gt", "le", "lt", "ne",
		))
	})

	ginkgo.It("recovers the delimiters and the trim marker", func() {
		Expect(g.LeftDelim).To(Equal("{{"))
		Expect(g.RightDelim).To(Equal("}}"))
		Expect(g.LeftComment).To(Equal("/*"))
		Expect(g.RightComment).To(Equal("*/"))
		Expect(g.TrimMarker).To(Equal("-"))
	})
})
