package gomplate

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The conversions CEL's own are too strict for. `string(x)` has a fixed set of
// overloads and refuses a map or a null; `int("3")` parses but nothing picks
// int-versus-double from the value; and there is no date formatter at all. Each
// is declared globally and as a member, so `text(x)` and `x.text()` both read
// naturally depending on where in an expression they land.
var _ = Describe("cast helpers", func() {
	evaluate := func(expression string, environment map[string]any) (any, error) {
		return RunExpression(environment, Template{Expression: expression})
	}

	DescribeTable("text stringifies anything",
		func(expression string, expected string) {
			Expect(evaluate(expression, map[string]any{
				"value": map[string]any{"name": "folio", "count": 2},
			})).To(Equal(expected))
		},
		Entry("a string is itself", `"folio".text()`, "folio"),
		Entry("an integer", `(2).text()`, "2"),
		Entry("a double keeps its fraction", `(6.9).text()`, "6.9"),
		Entry("a bool", `true.text()`, "true"),
		Entry("the global form", `text("folio")`, "folio"),
		// The case `string()` refuses outright, and the reason this exists: a
		// register records what a scraper reported, whatever shape it arrived in.
		Entry("a map", `value.name.text()`, "folio"),
	)

	DescribeTable("int parses and truncates",
		func(expression string, expected int64) {
			Expect(evaluate(expression, nil)).To(Equal(expected))
		},
		Entry("a numeric string", `"3".int()`, int64(3)),
		Entry("a string with surrounding space", `" 17 ".int()`, int64(17)),
		Entry("a double truncates", `(6.9).int()`, int64(6)),
		Entry("an integer is itself", `(3).int()`, int64(3)),
		Entry("the global form", `int("3")`, int64(3)),
	)

	DescribeTable("float parses",
		func(expression string, expected float64) {
			Expect(evaluate(expression, nil)).To(Equal(expected))
		},
		Entry("a fractional string", `"6.9".float()`, 6.9),
		Entry("a whole string", `"7".float()`, float64(7)),
		Entry("an integer widens", `(7).float()`, float64(7)),
	)

	DescribeTable("date formats a timestamp as YYYY-MM-DD",
		func(expression string, expected string) {
			Expect(evaluate(expression, map[string]any{
				"observed_at": "2026-03-24T18:45:00Z",
			})).To(Equal(expected))
		},
		Entry("from a timestamp", `timestamp("2026-03-24T00:00:00Z").date()`, "2026-03-24"),
		// The string receiver is what lets a projection read `source.first_observed.date()`
		// instead of wrapping every field in `timestamp()` first.
		Entry("from an RFC3339 string", `observed_at.date()`, "2026-03-24"),
		Entry("the global form", `date("2026-03-24T18:45:00Z")`, "2026-03-24"),
		// Not the local day: a register that says a finding was detected on the
		// 24th must mean the same day to every reader of the document.
		Entry("normalises to UTC", `timestamp("2026-03-24T23:30:00-05:00").date()`, "2026-03-25"),
	)

	DescribeTable("reports a value it cannot convert rather than returning a zero",
		func(expression string, message string) {
			_, err := evaluate(expression, nil)
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("a word is not an integer", `"several".int()`, `cannot parse "several"`),
		Entry("a word is not a float", `"several".float()`, `cannot parse "several"`),
		// Go parses "NaN", and JSON cannot write it. Both conversions refuse it
		// rather than handing back something that cannot be serialised.
		Entry("NaN is not an integer", `"NaN".int()`, "NaN"),
		Entry("NaN is not a float", `"NaN".float()`, "NaN"),
		// Beyond int64 the Go conversion is undefined and lands on the minimum
		// int64, so a byte count past 2^63 would read as a large negative number.
		Entry("beyond int64 is refused", `"18446744073709551616".int()`, "cannot represent"),
		Entry("a phrase is not a date", `"last tuesday".date()`, "RFC3339"),
	)

	// The value .int() refuses is exactly the one .float() is for.
	It("keeps a number too large for an int as a double", func() {
		Expect(evaluate(`"18446744073709551616".float()`, nil)).To(Equal(math.Pow(2, 64)))
	})
})
