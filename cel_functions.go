package gomplate

import (
	"fmt"
	"sort"
	"strings"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"github.com/flanksource/gomplate/v4/celvalues"
)

// CelFunction declares a CEL function whose implementation is supplied on each evaluation, so
// the compiled environment and program can be cached while Impl closes over per-call state
// (a request context, a database handle, ...).
type CelFunction struct {
	Name      string
	Overloads []CelOverload
	// Impl serves every overload; overloadID names the one that matched the call.
	Impl cel.LateBoundFunction
}

// CelOverload is one signature of a CelFunction. Member declares a receiver-style overload,
// whose receiver is Args[0].
type CelOverload struct {
	ID     string
	Member bool
	Args   []*cel.Type
	Result *cel.Type
}

// celFunctions returns the template's CelFunctions followed by its zero-argument Functions,
// which are declared the same way so they no longer stop the program from being cached.
func (t Template) celFunctions() []CelFunction {
	functions := make([]CelFunction, 0, len(t.CelFunctions)+len(t.Functions))
	functions = append(functions, t.CelFunctions...)
	names := make([]string, 0, len(t.Functions))
	for name := range t.Functions {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		functions = append(functions, zeroArgCelFunction(name, t.Functions[name]))
	}
	return functions
}

func zeroArgCelFunction(name string, function any) CelFunction {
	return CelFunction{
		Name:      name,
		Overloads: []CelOverload{{ID: name, Result: cel.AnyType}},
		Impl: func(string, ...ref.Val) ref.Val {
			fn, ok := function.(func() any)
			if !ok {
				return types.WrapErr(fmt.Errorf("%s is expected to be of type func() any", name))
			}
			return celvalues.NewAdapter(types.DefaultTypeAdapter).NativeToValue(fn())
		},
	}
}

func celFunctionDeclarations(functions []CelFunction) []cel.EnvOption {
	options := make([]cel.EnvOption, 0, len(functions))
	for _, function := range functions {
		overloads := make([]cel.FunctionOpt, 0, len(function.Overloads))
		for _, overload := range function.Overloads {
			declare := cel.Overload
			if overload.Member {
				declare = cel.MemberOverload
			}
			overloads = append(overloads, declare(overload.ID, overload.Args, overload.Result, cel.LateFunctionBinding()))
		}
		options = append(options, cel.Function(function.Name, overloads...))
	}
	return options
}

// celFunctionImpls validates the declarations and returns the implementations to bind for
// one evaluation.
func celFunctionImpls(functions []CelFunction) (map[string]cel.LateBoundFunction, error) {
	if len(functions) == 0 {
		return nil, nil
	}
	impls := make(map[string]cel.LateBoundFunction, len(functions))
	for _, function := range functions {
		if function.Impl == nil {
			return nil, fmt.Errorf("cel function %s has no implementation", function.Name)
		}
		if len(function.Overloads) == 0 {
			return nil, fmt.Errorf("cel function %s declares no overloads", function.Name)
		}
		for _, overload := range function.Overloads {
			if overload.ID == "" || overload.Result == nil {
				return nil, fmt.Errorf("cel function %s has an overload without an id or result type", function.Name)
			}
		}
		if _, duplicate := impls[function.Name]; duplicate {
			return nil, fmt.Errorf("cel function %s is declared more than once", function.Name)
		}
		impls[function.Name] = function.Impl
	}
	return impls, nil
}

// celFunctionSignature identifies the declarations, never the implementations, so templates
// that declare the same functions share an environment and a program.
func celFunctionSignature(functions []CelFunction) string {
	var signature strings.Builder
	for _, function := range functions {
		signature.WriteString(function.Name)
		for _, overload := range function.Overloads {
			fmt.Fprintf(&signature, "|%s:%t:", overload.ID, overload.Member)
			for _, arg := range overload.Args {
				signature.WriteString(arg.String() + ",")
			}
			signature.WriteString("->" + overload.Result.String())
		}
		signature.WriteString(";")
	}
	return signature.String()
}
