package gomplate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// What an expression returns when it builds a value rather than selecting one.
//
// A map or list constructed inside CEL is held in cel-go's own representation,
// whose Value() is a map[ref.Val]ref.Val -- it marshals to {"Adapter":{}} and is
// unusable to a caller that wants to write the result somewhere. Every consumer
// that returns structured data hits this; the ones returning strings and bools
// never do, which is why it went unnoticed.
var _ = Describe("expression results", func() {
	evaluate := func(expression string, environment map[string]any) any {
		result, err := RunExpression(environment, Template{Expression: expression})
		Expect(err).NotTo(HaveOccurred())
		return result
	}

	It("returns a constructed map as a plain Go map", func() {
		Expect(evaluate(`{"kind": "finding", "severity": "high"}`, nil)).
			To(Equal(map[string]any{"kind": "finding", "severity": "high"}))
	})

	It("returns a constructed list of maps as plain Go values", func() {
		Expect(evaluate(`[{"kind": "finding"}, {"kind": "inventory"}]`, nil)).
			To(Equal([]any{
				map[string]any{"kind": "finding"},
				map[string]any{"kind": "inventory"},
			}))
	})

	It("converts the whole way down, not just the outermost level", func() {
		Expect(evaluate(`{"finding": {"advisory": {"id": "GHSA-p63j-vcc4-9vmv"}}}`, nil)).
			To(Equal(map[string]any{
				"finding": map[string]any{"advisory": map[string]any{"id": "GHSA-p63j-vcc4-9vmv"}},
			}))
	})

	// A count written back as `2.0` would make a schema declaring the field an
	// integer describe something the document does not look like. A genuinely
	// fractional value is left alone.
	It("keeps whole numbers whole and fractions fractional", func() {
		Expect(evaluate(`{"count": 2, "score": 6.9}`, nil)).
			To(Equal(map[string]any{"count": int64(2), "score": 6.9}))
	})

	It("reads a value out of the environment into a constructed map", func() {
		Expect(evaluate(`{"name": source.name, "day": source.first_observed.date()}`, map[string]any{
			"source": map[string]any{"name": "folio", "first_observed": "2026-03-24T00:00:00Z"},
		})).To(Equal(map[string]any{"name": "folio", "day": "2026-03-24"}))
	})

	// out.Value() on a CEL null is structpb.NullValue -- a protobuf enum whose
	// value is 0, so a caller checking for nil sees a number and a caller writing
	// the result records a zero where the expression said nothing.
	It("returns a null as a nil", func() {
		Expect(evaluate(`null`, nil)).To(BeNil())
	})

	It("returns a null inside a constructed map as a nil", func() {
		Expect(evaluate(`{"owner": null, "name": "folio"}`, nil)).
			To(Equal(map[string]any{"owner": nil, "name": "folio"}))
	})

	DescribeTable("leaves a scalar exactly as it was",
		func(expression string, expected any) {
			Expect(evaluate(expression, nil)).To(Equal(expected))
		},
		Entry("a string", `"folio"`, "folio"),
		Entry("an integer", `2`, int64(2)),
		Entry("a double", `6.9`, 6.9),
		Entry("a bool", `true`, true),
	)

	// Selecting a map straight out of the environment never went through cel-go's
	// own representation, so it must come back untouched -- including the integer
	// widths a caller may be asserting on.
	It("returns an environment map unchanged", func() {
		Expect(evaluate(`tags`, map[string]any{
			"tags": map[string]any{"cluster": "production", "replicas": int64(3)},
		})).To(Equal(map[string]any{"cluster": "production", "replicas": int64(3)}))
	})
})
