package gomplate

import (
	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
	"github.com/flanksource/commons/decimal"
	"regexp"
)

var plainDecimal = regexp.MustCompile(`^[+-]?[0-9]+(\.[0-9]+)?$`)

func decimalSumOption() cel.EnvOption {
	return cel.Function("sum", cel.MemberOverload("collection_sum_decimal_strings", []*cel.Type{cel.ListType(cel.StringType)}, cel.StringType, cel.UnaryBinding(sumDecimalStrings)))
}
func sumDecimalStrings(value ref.Val) ref.Val {
	numbers := make([]decimal.Decimal, 0, int(value.(traits.Lister).Size().(types.Int)))
	for iterator := value.(traits.Lister).Iterator(); iterator.HasNext() == types.True; {
		item := iterator.Next()
		text, ok := item.(types.String)
		if !ok || !plainDecimal.MatchString(string(text)) {
			return types.NewErr("sum requires decimal strings, got %v", item)
		}
		number, err := decimal.Parse(string(text))
		if err != nil {
			return types.NewErr("sum parsing %q: %v", text, err)
		}
		numbers = append(numbers, number)
	}
	if len(numbers) == 0 {
		return types.String("0")
	}
	total, err := decimal.Sum(numbers...)
	if err != nil {
		return types.NewErr("decimal sum: %v", err)
	}
	return types.String(total.String())
}
