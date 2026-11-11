package gomplate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These pairs back the "Equivalent functions" table in the canary-checker CEL docs: each entry
// evaluates both spellings on an input that shows where they agree or differ.
var _ = Describe("Equivalent CEL functions", func() {
	DescribeTable("evaluate to the documented results",
		func(expression string, expected any) {
			Expect(RunExpression(nil, Template{Expression: expression})).To(Equal(expected))
		},
		Entry("lowerAscii leaves non-ASCII letters", `"ÆX".lowerAscii()`, "Æx"),
		Entry("toLower folds Unicode letters", `"ÆX".toLower()`, "æx"),
		Entry("upperAscii leaves non-ASCII letters", `"æx".upperAscii()`, "æX"),
		Entry("toUpper folds Unicode letters", `"æx".toUpper()`, "ÆX"),
		Entry("replace with a count", `"aaa".replace("a", "b", 2)`, "bba"),
		Entry("replace without a count", `"aaa".replace("a", "b")`, "bbb"),
		Entry("replaceAll", `"aaa".replaceAll("a", "b")`, "bbb"),
		Entry("trim removes whitespace", "\" \\tx\\n \".trim()", "x"),
		Entry("trimSpace removes whitespace", "\" \\tx\\n \".trimSpace()", "x"),
		Entry("strings.quote escapes for CEL", `strings.quote("a\"b")`, `"a\"b"`),
		Entry("quote", `"a\"b".quote()`, `"a\"b"`),
		Entry("squote", `"it".squote()`, `'it'`),
		Entry("split with a limit", `"a,b,c".split(",", 2)`, []string{"a", "b,c"}),
		Entry("distinct keeps first occurrences in order", `[2, 1, 2, 1].distinct()`, []any{int64(2), int64(1)}),
		Entry("uniq", `[2, 1, 2, 1].uniq()`, []any{int64(2), int64(1)}),
		Entry("sort", `[3, 1, 2].sort()`, []any{int64(1), int64(2), int64(3)}),
		Entry("sortBy macro", `[{"n": 2}, {"n": 1}].sortBy(e, e.n).map(e, e.n)`, []any{int64(1), int64(2)}),
		Entry("lists.range counts from zero", `lists.range(3)`, []int{0, 1, 2}),
		Entry("math.Seq takes a list of bounds, both inclusive", `math.Seq([0, 2])`, []int64{0, 1, 2}),
		Entry("math.abs keeps the int type", `math.abs(-2)`, int64(2)),
		Entry("math.Abs returns a double", `math.Abs(-2)`, float64(2)),
		Entry("math.greatest", `math.greatest(1, 5, 3)`, int64(5)),
		Entry("math.least", `math.least([4, 2])`, int64(2)),
		Entry("json.encode", `json.encode({"a": 1})`, `{"a":1}`),
		Entry("toJSON", `{"a": 1}.toJSON()`, `{"a":1}`),
		Entry("json.parse yields integral numbers as int", `json.parse('{"a":1}').a`, int64(1)),
		Entry("yaml.encode", `yaml.encode({"a": 1})`, "a: 1\n"),
		Entry("toYAML", `toYAML({"a": 1})`, "a: 1\n"),
		Entry("yaml.parse", `yaml.parse("a: 1").a`, int64(1)),
		Entry("YAML", `YAML("a: 1").a`, int64(1)),
		Entry("JSON parses whole numbers as ints", `'{"a":1}'.JSON().a`, int64(1)),
		Entry("list indexOf", `[1, 2, 3].indexOf(2)`, int64(1)),
		Entry("string indexOf", `"hello".indexOf("l")`, int64(2)),
		Entry("string lastIndexOf", `"hello".lastIndexOf("l")`, int64(3)),
	)
})
