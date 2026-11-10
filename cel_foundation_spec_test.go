package gomplate

import (
	"time"

	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CEL evaluation foundation", func() {
	It("preserves integers beyond JSON precision in constructed aggregates", func() {
		value, err := RunExpression(nil, Template{Expression: `{"rows":[{"amount":9007199254740993,"fraction":1.0}]}`})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(map[string]any{"rows": []any{map[string]any{"amount": int64(9007199254740993), "fraction": float64(1)}}}))
	})

	It("does not mutate values reused across evaluations", func() {
		created := time.Date(2026, time.November, 4, 11, 8, 19, 0, time.UTC)
		step := map[string]any{"created": &created, "rows": []any{map[string]any{"due": "2026-09-01"}}}
		environment := map[string]any{"step": step}
		for range 2 {
			value, err := RunExpression(environment, Template{Expression: `step.rows[0].due == "2026-09-01"`})
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(BeTrue())
			Expect(step["created"]).To(BeIdenticalTo(&created))
		}
	})

	It("preserves CEL values nested in caller-provided maps and lists", func() {
		native := types.NewRefValList(types.DefaultTypeAdapter, []ref.Val{types.Int(9007199254740993)})
		environment := map[string]any{"rows": []any{map[string]any{"native": native}}}
		value, err := RunExpression(environment, Template{Expression: `rows[0].native[0]`})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(int64(9007199254740993)))
		Expect(environment["rows"].([]any)[0].(map[string]any)["native"]).To(BeIdenticalTo(native))
	})
})
