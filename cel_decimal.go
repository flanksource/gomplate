package gomplate

import (
	"math/big"
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

var plainDecimal = regexp.MustCompile(`^[+-]?[0-9]+(\.[0-9]+)?$`)

const celDecimalStringLengthLimit = 4096

func decimalSumOption() cel.EnvOption {
	return cel.Function("sum", cel.MemberOverload("collection_sum_decimal_strings", []*cel.Type{cel.ListType(cel.StringType)}, cel.StringType, cel.UnaryBinding(sumDecimalStrings)))
}

// sumDecimalStrings adds bounded plain decimal strings without rounding.
func sumDecimalStrings(value ref.Val) ref.Val {
	total, scale := new(big.Rat), 0
	for iterator := value.(traits.Lister).Iterator(); iterator.HasNext() == types.True; {
		item := iterator.Next()
		text, ok := item.(types.String)
		if ok && len(text) > celDecimalStringLengthLimit {
			return types.NewErr("sum decimal string exceeds limit of %d bytes", celDecimalStringLengthLimit)
		}
		if !ok || !plainDecimal.MatchString(string(text)) {
			return types.NewErr("sum requires decimal strings, got %v", item)
		}
		number, ok := new(big.Rat).SetString(string(text))
		if !ok {
			return types.NewErr("sum requires decimal strings, got %q", text)
		}
		if point := strings.IndexByte(string(text), '.'); point >= 0 {
			scale = max(scale, len(text)-point-1)
		}
		total.Add(total, number)
	}
	result := total.FloatString(scale)
	if scale > 0 {
		result = strings.TrimRight(strings.TrimRight(result, "0"), ".")
	}
	return types.String(result)
}
