// Code generated by gencel. DO NOT EDIT.

package strings

import "log"
import "github.com/google/cel-go/common/types/ref"
import "github.com/google/cel-go/cel"

func transferSlice[K any](arg ref.Val) []K {
	list, ok := arg.Value().([]ref.Val)
	if !ok {
		log.Printf("Not a list %T\n", arg.Value())
		return nil
	}

	var out = make([]K, len(list))
	for i, val := range list {
		out[i] = val.Value().(K)
	}

	return out
}

var CelEnvOption = []cel.EnvOption{
	durationStringGen,
	durationNanosecondsGen,
	durationSecondsGen,
	durationHoursGen,
	durationDaysGen,
	durationWeeksGen,
	durationMinutesGen,
}
