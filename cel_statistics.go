package gomplate

import (
	"math"
	"slices"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
)

func statisticsCelOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Function("percentile", cel.MemberOverload("collection_percentile", []*cel.Type{cel.ListType(cel.DynType), cel.DynType}, cel.DoubleType, cel.BinaryBinding(percentileOf))),
		cel.Function("stats", cel.MemberOverload("collection_stats", []*cel.Type{cel.ListType(cel.DynType)}, cel.MapType(cel.StringType, cel.DynType), cel.UnaryBinding(statsOf))),
	}
}

// orderedSamples returns the list's finite int, uint or double samples in ascending order.
func orderedSamples(name string, value ref.Val) ([]float64, ref.Val) {
	list := value.(traits.Lister)
	samples := make([]float64, 0, int(list.Size().(types.Int)))
	for iterator := list.Iterator(); iterator.HasNext() == types.True; {
		var sample float64
		switch item := iterator.Next().(type) {
		case types.Int:
			sample = float64(item)
		case types.Uint:
			sample = float64(item)
		case types.Double:
			sample = float64(item)
		default:
			return nil, types.NewErr("%s requires numeric samples, got %s", name, item.Type())
		}
		if math.IsNaN(sample) || math.IsInf(sample, 0) {
			return nil, types.NewErr("%s requires finite samples, got %v", name, sample)
		}
		samples = append(samples, sample)
	}
	if len(samples) == 0 {
		return nil, types.NewErr("%s requires at least one sample", name)
	}
	slices.Sort(samples)
	return samples, nil
}

// nearestRank is sorted[max(ceil(q*n/100)-1, 0)] for q in (0, 100]. q*n is
// multiplied before dividing so whole-number ranks stay exact (7/100*100 is
// 7.000000000000001, which would round up to the next sample).
func nearestRank(ordered []float64, q float64) types.Double {
	return types.Double(ordered[max(int(math.Ceil(q*float64(len(ordered))/100))-1, 0)])
}

func percentileOf(value, rank ref.Val) ref.Val {
	var q float64
	switch number := rank.(type) {
	case types.Int:
		q = float64(number)
	case types.Uint:
		q = float64(number)
	case types.Double:
		q = float64(number)
	default:
		return types.NewErr("percentile requires a numeric q, got %s", rank.Type())
	}
	if !(q > 0 && q <= 100) {
		return types.NewErr("percentile requires q in (0, 100], got %v", q)
	}
	ordered, err := orderedSamples("percentile", value)
	if err != nil {
		return err
	}
	return nearestRank(ordered, q)
}

// meanOf averages finite samples; when their sum overflows it averages them
// scaled by the largest magnitude so the mean of finite samples stays finite.
func meanOf(ordered []float64) float64 {
	count := float64(len(ordered))
	total := 0.0
	for _, sample := range ordered {
		total += sample
	}
	if !math.IsInf(total, 0) {
		return total / count
	}
	scale := max(-ordered[0], ordered[len(ordered)-1])
	total = 0
	for _, sample := range ordered {
		total += sample / scale
	}
	return total / count * scale
}

func statsOf(value ref.Val) ref.Val {
	ordered, err := orderedSamples("stats", value)
	if err != nil {
		return err
	}
	return types.NewRefValMap(types.DefaultTypeAdapter, map[ref.Val]ref.Val{
		types.String("min"): types.Double(ordered[0]), types.String("avg"): types.Double(meanOf(ordered)),
		types.String("max"): types.Double(ordered[len(ordered)-1]),
		types.String("p50"): nearestRank(ordered, 50), types.String("p90"): nearestRank(ordered, 90),
		types.String("p95"): nearestRank(ordered, 95), types.String("p99"): nearestRank(ordered, 99),
		types.String("samples"): types.Int(len(ordered)),
	})
}
