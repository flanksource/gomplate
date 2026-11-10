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

func bigMoney(amount, currency string, places int) bigdecimal.Money {
	value, err := bigdecimal.NewMoney(bigdecimal.MoneyOptions{Amount: amount, Currency: currency, DecimalPlaces: places})
	Expect(err).NotTo(HaveOccurred())
	return value
}

var _ = Describe("Money arithmetic", func() {
	bindings := func() map[string]any {
		return map[string]any{
			"total": bigMoney("100", "USD", 2),
			"fee":   bigMoney("12.35", "USD", 2),
			"eur":   bigMoney("1", "EUR", 2),
			"small": func() decimal.Money {
				value, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "1", Currency: "USD", DecimalPlaces: 2})
				Expect(err).NotTo(HaveOccurred())
				return value
			}(),
			"row": map[string]any{"total": bigMoney("47500", "ZAR", 2)},
		}
	}

	DescribeTable("keeps results exact Money in either operand position",
		func(expression, want string) {
			value, err := evaluate("string("+expression+")", bindings())
			Expect(err).NotTo(HaveOccurred(), expression)
			Expect(value).To(Equal(want), expression)
		},
		Entry("money plus money", "total + fee", "USD 112.35"),
		Entry("money minus money", "total - fee", "USD 87.65"),
		Entry("money plus int", "total + 15000", "USD 15100.00"),
		Entry("int plus money", "15000 + total", "USD 15100.00"),
		Entry("double minus money", "0.1 - fee", "USD -12.25"),
		Entry("money minus double", "total - 0.1 - 0.2", "USD 99.70"),
		Entry("money times int", "fee * 3", "USD 37.05"),
		Entry("uint times money", "3u * fee", "USD 37.05"),
		Entry("money divided by int", "total / 8", "USD 12.50"),
		Entry("negation", "-fee", "USD -12.35"),
		Entry("nested record field", "row.total + 15000 - 12.35", "ZAR 62487.65"),
		Entry("chained money products", "row.total - row.total / 4 * 3", "ZAR 11875.00"),
	)

	It("compares arithmetic results exactly", func() {
		value, err := evaluate(`total - 0.1 - 0.2 == 99.7 && total + fee * 3 == 137.05 && -fee < 0`, bindings())
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(BeTrue())
	})

	It("leaves ordinary numeric arithmetic to CEL", func() {
		value, err := evaluate(`1 + 2 == 3 && 1.5 * 2.0 == 3.0 && -(2) == -2 && "a" + "b" == "ab"`, bindings())
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(BeTrue())
	})

	DescribeTable("rejects invalid Money arithmetic",
		func(expression, message string) {
			_, err := evaluate(expression, bindings())
			Expect(err).To(MatchError(ContainSubstring(message)), expression)
		},
		Entry("different currencies", "total + eur", "different currencies: USD and EUR"),
		Entry("division by zero", "total / 0", "non-finite"),
		Entry("money times money", "total * fee", "cannot multiply Money by Money"),
		Entry("number divided by money", "1 / total", "cannot divide a number by Money"),
		Entry("string operand", `total + "1"`, "expects money or a number, got string"),
		Entry("decimal.Money", "small + 1", "commons/bigdecimal.Money"),
	)

	It("works through gomplate templates", func() {
		value, err := gomplate.RunExpression(bindings(), gomplate.Template{Expression: `row.total + 15000 - 12.35 == 62487.65 && string(-fee) == "USD -12.35"`, Strict: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(BeTrue())
	})
})
