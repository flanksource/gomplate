package celvalues_test

import (
	"github.com/flanksource/commons/bigdecimal"
	"github.com/flanksource/commons/decimal"
	"github.com/flanksource/gomplate/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type currencyRecord struct {
	Premium bigdecimal.Money `json:"premium"`
}

var _ = Describe("Commons Money integration", func() {
	It("compares big Money with every numeric type and operator in both positions", func() {
		money, err := bigdecimal.NewMoney(bigdecimal.MoneyOptions{Amount: "125.00", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		for _, operand := range []string{"125", "125u", "125.0", "same"} {
			for _, operator := range []string{"==", "!=", "<", "<=", ">", ">="} {
				want := operator == "==" || operator == "<=" || operator == ">="
				for _, expression := range []string{"money " + operator + " " + operand, operand + " " + operator + " money"} {
					value, err := evaluate(expression, map[string]any{"money": money, "same": money})
					Expect(err).NotTo(HaveOccurred(), expression)
					Expect(value).To(Equal(want), expression)
				}
			}
		}
		other, err := bigdecimal.NewMoney(bigdecimal.MoneyOptions{Amount: "125", Currency: "EUR"})
		Expect(err).NotTo(HaveOccurred())
		_, err = evaluate("money == other", map[string]any{"money": money, "other": other})
		Expect(err).To(MatchError(ContainSubstring("currencies")))
	})
	It("preserves higher precision in nested and registered native inputs", func() {
		money, err := bigdecimal.NewMoney(bigdecimal.MoneyOptions{Amount: "9007199254740993.0000000001", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		Expect(gomplate.RegisterType(currencyRecord{})).To(Succeed())
		for _, host := range []any{map[string]any{"premium": money}, currencyRecord{Premium: money}} {
			value, err := gomplate.RunExpression(map[string]any{"row": host}, gomplate.Template{Expression: `row.premium > 9007199254740993 && 9007199254740993 < row.premium && string(row.premium) == "USD 9007199254740993.0000000001"`, Strict: true})
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(BeTrue())
			result, err := gomplate.RunExpression(map[string]any{"row": host}, gomplate.Template{Expression: `[row.premium]`, Strict: true})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]any{money}))
		}
	})
	It("adapts Money returned by template functions and rejects mixed precision", func() {
		small, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "125", Currency: "USD"})
		Expect(err).NotTo(HaveOccurred())
		value, err := gomplate.RunExpression(nil, gomplate.Template{Expression: `lookup() == 125`, Strict: true, Functions: map[string]any{"lookup": func() any { return small }}})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(BeTrue())
		large, err := bigdecimal.NewMoney(bigdecimal.MoneyOptions{Amount: "125", Currency: "USD"})
		Expect(err).NotTo(HaveOccurred())
		_, err = gomplate.RunExpression(map[string]any{"small": small, "large": large}, gomplate.Template{Expression: `small == large`, Strict: true})
		Expect(err).To(HaveOccurred())
	})
})
