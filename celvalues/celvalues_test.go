package celvalues_test

import (
	"fmt"
	"math"
	"testing"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"github.com/flanksource/commons/decimal"
	"github.com/flanksource/gomplate/v4/celvalues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCELValues(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Typed CEL values")
}

func evaluate(expression string, bindings map[string]any) (any, error) {
	options := make([]cel.EnvOption, 0, 1+len(bindings))
	options = append(options, celvalues.Library())
	for name := range bindings {
		options = append(options, cel.Variable(name, cel.DynType))
	}
	env, err := cel.NewEnv(options...)
	if err != nil {
		return nil, err
	}
	tree, issues := env.Compile(expression)
	if issues.Err() != nil {
		return nil, issues.Err()
	}
	program, err := env.Program(tree)
	if err != nil {
		return nil, err
	}
	value, _, err := program.Eval(bindings)
	if err != nil {
		return nil, err
	}
	return value.Value(), nil
}

var _ = Describe("Date values", func() {
	It("compares calendar days in both operand positions and renders OIPA dates", func() {
		date, err := celvalues.NewDate("2026-11-10")
		Expect(err).NotTo(HaveOccurred())
		for _, operand := range []string{`"2026-11-10"`, `timestamp("2026-11-10T13:45:00Z")`, "same"} {
			for _, operator := range []string{"==", "!=", "<", "<=", ">", ">="} {
				want := operator == "==" || operator == "<=" || operator == ">="
				for _, expression := range []string{fmt.Sprintf("date %s %s", operator, operand), fmt.Sprintf("%s %s date", operand, operator)} {
					value, err := evaluate(expression, map[string]any{"date": date, "same": date})
					Expect(err).NotTo(HaveOccurred(), expression)
					Expect(value).To(Equal(want), expression)
				}
			}
		}
		value, err := evaluate(`string(date)`, map[string]any{"date": date})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal("2026-11-10"))
	})

	It("orders leap days and uses the UTC calendar date of timestamp operands", func() {
		leap, err := celvalues.NewDate("2024-02-29")
		Expect(err).NotTo(HaveOccurred())
		value, err := evaluate(`date > "2024-02-28" && "2024-03-01" > date && date == timestamp("2024-03-01T01:00:00+02:00")`, map[string]any{"date": leap})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(true))
	})

	It("rejects invalid dates and invalid comparison operands", func() {
		for _, raw := range []string{"", "2023-02-29", "10/11/2026", "2026-11-10T00:00:00Z"} {
			_, err := celvalues.NewDate(raw)
			Expect(err).To(HaveOccurred(), raw)
		}
		date, err := celvalues.NewDate("2026-11-10")
		Expect(err).NotTo(HaveOccurred())
		for _, expression := range []string{`date < "tomorrow"`, `"tomorrow" == date`, `date > 0`} {
			_, err := evaluate(expression, map[string]any{"date": date})
			Expect(err).To(HaveOccurred(), expression)
		}
	})
})

var _ = Describe("Money values", func() {
	It("compares exact amounts with all numeric types in either position", func() {
		money, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "125.00", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		for _, operand := range []string{"125", "125u", "125.0", "same"} {
			for _, operator := range []string{"==", "!=", "<", "<=", ">", ">="} {
				want := operator == "==" || operator == "<=" || operator == ">="
				for _, expression := range []string{fmt.Sprintf("money %s %s", operator, operand), fmt.Sprintf("%s %s money", operand, operator)} {
					value, err := evaluate(expression, map[string]any{"money": money, "same": money})
					Expect(err).NotTo(HaveOccurred(), expression)
					Expect(value).To(Equal(want), expression)
				}
			}
		}
		value, err := evaluate(`100 < money && money > 100 && string(money) == "USD 125.00"`, map[string]any{"money": money})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(true))
	})

	It("retains fractional precision, large amounts, signs and currency decimal places", func() {
		for _, example := range []struct {
			amount, currency, want string
			places                 int
		}{
			{"9007199254740993.0000000001", "USD", "USD 9007199254740993.00", 2},
			{"-0.125", "USD", "USD -0.125", 2},
			{"0", "JPY", "JPY 0", 0},
			{"1.2", "KWD", "KWD 1.200", 3},
		} {
			money, err := decimal.NewMoney(decimal.MoneyOptions{Amount: example.amount, Currency: example.currency, DecimalPlaces: example.places})
			Expect(err).NotTo(HaveOccurred())
			value, err := evaluate(`string(money)`, map[string]any{"money": money})
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(example.want))
		}
		money, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "9007199254740993.001", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		value, err := evaluate(`money > 9007199254740993`, map[string]any{"money": money})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(true))
	})

	It("rejects cross-currency comparisons for every operator", func() {
		usd, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "100", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		eur, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "100", Currency: "EUR", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		for _, operator := range []string{"==", "!=", "<", "<=", ">", ">="} {
			for _, expression := range []string{"usd " + operator + " eur", "eur " + operator + " usd"} {
				_, err := evaluate(expression, map[string]any{"usd": usd, "eur": eur})
				Expect(err).To(MatchError(ContainSubstring("currencies")), expression)
			}
		}
	})

	It("rejects invalid amounts, currencies and nonfinite numeric operands", func() {
		for _, options := range []decimal.MoneyOptions{
			{Amount: "NaN", Currency: "USD", DecimalPlaces: 2},
			{Amount: "1.0", Currency: "", DecimalPlaces: 2},
			{Amount: "1", Currency: "USD", DecimalPlaces: -1},
		} {
			_, err := decimal.NewMoney(options)
			Expect(err).To(HaveOccurred())
		}
		money, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "1", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		_, err = evaluate(`money > operand`, map[string]any{"money": money, "operand": types.Double(math.Inf(1))})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("typed value integration", func() {
	It("orders different values correctly for every comparison in either position", func() {
		earlier, err := celvalues.NewDate("2026-11-09")
		Expect(err).NotTo(HaveOccurred())
		later, err := celvalues.NewDate("2026-11-10")
		Expect(err).NotTo(HaveOccurred())
		lower, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "-0.125", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		higher, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "0.125", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		for _, pair := range []map[string]any{{"lower": earlier, "higher": later}, {"lower": lower, "higher": higher}} {
			for _, example := range []struct {
				operator string
				forward  bool
			}{
				{"==", false}, {"!=", true}, {"<", true}, {"<=", true}, {">", false}, {">=", false},
			} {
				value, err := evaluate("lower "+example.operator+" higher", pair)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal(example.forward))
				value, err = evaluate("higher "+example.operator+" lower", pair)
				Expect(err).NotTo(HaveOccurred())
				want := !example.forward
				if example.operator == "==" || example.operator == "!=" {
					want = example.forward
				}
				Expect(value).To(Equal(want))
			}
		}
	})

	It("preserves nested native values and null equality without changing ordinary operators", func() {
		date, err := celvalues.NewDate("2026-11-10")
		Expect(err).NotTo(HaveOccurred())
		money, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "100", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		bindings := map[string]any{"row": map[string]any{"date": date, "money": money, "absent": nil}}
		value, err := evaluate(`row.date != null && null != row.money && row.absent == null && [row.money].map(x, string(x))[0] == "USD 100.00" && 1 < 2 && "b" > "a"`, bindings)
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(true))
	})
})
