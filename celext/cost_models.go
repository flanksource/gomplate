package celext

import (
	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/cost"
)

// costModels is embedded in every vendored library: its cost models stay registered, as
// upstream, unless the library is built with its WithoutCostModels option.
type costModels struct{ disabled bool }

func (c costModels) costModel(models ...cost.OverloadModel) cel.EnvOption {
	if c.disabled {
		return func(env *cel.Env) (*cel.Env, error) { return env, nil }
	}
	return cel.CostModel(models...)
}

// StringsWithoutCostModels disables overload cost models for Strings.
func StringsWithoutCostModels() StringsOption {
	return func(lib *stringLib) *stringLib { lib.costs.disabled = true; return lib }
}

// EncodersWithoutCostModels disables overload cost models for Encoders.
func EncodersWithoutCostModels() EncodersOption {
	return func(lib *encoderLib) *encoderLib { lib.costs.disabled = true; return lib }
}

// ListsWithoutCostModels disables overload cost models for Lists.
func ListsWithoutCostModels() ListsOption {
	return func(lib *listsLib) *listsLib { lib.costs.disabled = true; return lib }
}

// MathWithoutCostModels disables overload cost models for Math.
func MathWithoutCostModels() MathOption {
	return func(lib *mathLib) *mathLib { lib.costs.disabled = true; return lib }
}

// SetsWithoutCostModels disables overload cost models for Sets.
func SetsWithoutCostModels() SetsOption {
	return func(lib *setsLib) *setsLib { lib.costs.disabled = true; return lib }
}
