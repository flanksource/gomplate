package cel

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/flanksource/gomplate/v3/funcs"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
)

func panIf(err error) {
	if err != nil {
		panic(err)
	}
}

func TestEnv(t *testing.T) {
	var envopt []cel.EnvOption
	envopt = append(envopt, funcs.ReplaceregexpGen)

	env, err := cel.NewEnv(envopt...)
	panIf(err)

	expr := `Replace("flank", "rank", "flanksource")`
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		panIf(err)
	}

	prg, err := env.Program(ast)
	panIf(err)

	out, _, err := prg.Eval(map[string]any{})
	panIf(err)

	if out.Value() != "ranksource" {
		t.Fatalf("Expected ranksource got %s\n", out)
	}
}

func Split(name string) (string, string) {
	return name, name + "suffix"
}

var ReplaceGen = cel.Function("Suffixer",
	cel.Overload("Replace_interface{}_interface{}_interface{}",
		[]*cel.Type{
			cel.StringType,
		},
		cel.StringType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			name := args[0].Value().(string)
			a, b := Split(name)
			return types.DefaultTypeAdapter.NativeToValue([]any{a, b})
		}),
	),
)

func TestMultipleReturns(t *testing.T) {
	var envopt []cel.EnvOption
	envopt = append(envopt, ReplaceGen)

	env, err := cel.NewEnv(envopt...)
	panIf(err)

	expr := `Suffixer("flanksource")`
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		panIf(err)
	}

	prg, err := env.Program(ast)
	panIf(err)

	out, _, err := prg.Eval(map[string]any{})
	panIf(err)

	res := out.Value().([]any)
	if len(res) != 2 {
		t.Fatalf("Expected ranksource got %s\n", out)
	}
}

type Person struct {
	Name string
}

func (t *Person) Speak() string {
	return fmt.Sprintf("HI %s", t.Name)
}

func TestEnvNative(t *testing.T) {
	var p Person

	env, err := cel.NewEnv(ext.NativeTypes(reflect.TypeOf(p)))
	panIf(err)

	expr := `(cel.Person{Name: "ranksource"}).Name`
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		panIf(issues.Err())
	}

	prg, err := env.Program(ast)
	panIf(err)

	out, _, err := prg.Eval(map[string]any{})
	panIf(err)

	if out.Value() != "ranksource" {
		t.Fatalf("Expected ranksource got %s\n", out)
	}
}
