package celvalues_test

import (
	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"github.com/flanksource/commons/decimal"
	"github.com/flanksource/gomplate/v4"
	"github.com/flanksource/gomplate/v4/celvalues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gomplate typed scalar integration", func() {
	It("retains typed values through native input, constructed aggregates and custom function round trips", func() {
		date, err := celvalues.NewDate("2026-11-10")
		Expect(err).NotTo(HaveOccurred())
		money, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "9007199254740993.001", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		row := map[string]any{"date": date, "money": money, "nothing": nil}
		result, err := gomplate.RunExpression(map[string]any{"row": row}, gomplate.Template{
			Expression: `{"day": row.date, "amounts": [row.money], "nothing": row.nothing}`,
			Strict:     true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(map[string]any{"day": date, "amounts": []any{money}, "nothing": nil}))
		value, err := gomplate.RunExpression(map[string]any{"saved": result}, gomplate.Template{
			Expression: `roundTrip(saved).day == "2026-11-10" && roundTrip(saved).amounts[0] > 9007199254740993`,
			Strict:     true,
			CelEnvs: []cel.EnvOption{cel.Function("roundTrip", cel.Overload("round_trip", []*cel.Type{cel.DynType}, cel.DynType,
				cel.UnaryBinding(func(value ref.Val) ref.Val {
					return celvalues.NewAdapter(types.DefaultTypeAdapter).NativeToValue(gomplate.CELValue(value))
				})))},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(true))
		Expect(row).To(Equal(map[string]any{"date": date, "money": money, "nothing": nil}))
	})

	It("retains existing nilsafe and plain scalar comparison behavior", func() {
		for _, expression := range []string{
			`row.absent == 0`,
			`row.absent == ""`,
			`row.absent < 2`,
			`1 < 2 && 2u >= 1u && "a" < "b" && [1,2] == [1,2]`,
		} {
			value, err := gomplate.RunExpression(map[string]any{"row": map[string]any{}}, gomplate.Template{Expression: expression})
			Expect(err).NotTo(HaveOccurred(), expression)
			Expect(value).To(Equal(true), expression)
		}
	})

	It("evaluates comparison operands once in strict mode", func() {
		money, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "125", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		for _, expression := range []string{`lookup() > 100`, `100 < lookup()`, `lookup() == 125`, `125 == lookup()`} {
			calls := 0
			value, err := gomplate.RunExpression(nil, gomplate.Template{
				Expression: expression,
				Strict:     true,
				CelEnvs: []cel.EnvOption{cel.Function("lookup", cel.Overload("lookup_money", nil, cel.DynType,
					cel.FunctionBinding(func(...ref.Val) ref.Val {
						calls++
						return celvalues.NewAdapter(types.DefaultTypeAdapter).NativeToValue(money)
					})))},
			})
			Expect(err).NotTo(HaveOccurred(), expression)
			Expect(value).To(Equal(true), expression)
			Expect(calls).To(Equal(1), expression)
		}
	})
})
