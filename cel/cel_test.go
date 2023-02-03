package cel

import (
	"testing"

	"github.com/flanksource/gomplate/v3/funcs"
	"github.com/google/cel-go/cel"
)

func panIf(err error) {
	if err != nil {
		panic(err)
	}
}

func TestCelNamespace(t *testing.T) {
	var envopt []cel.EnvOption
	envopt = append(envopt, funcs.CelEnvOption...)

	env, err := cel.NewEnv(envopt...)
	panIf(err)

	expr := `regexp.Replace("flank", "rank", "flanksource")`
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

func TestCelMultipleReturns(t *testing.T) {
	env, err := cel.NewEnv(funcs.CelEnvOption...)
	panIf(err)

	expr := `base64.Encode("flanksource")`
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

	if res[0] != "Zmxhbmtzb3VyY2U=" {
		t.Fatalf("Expected Zmxhbmtzb3VyY2U= got %s\n", res[0])
	}
}

func TestCelVariadic(t *testing.T) {
	env, err := cel.NewEnv(funcs.CelEnvOption...)
	panIf(err)

	expr := `math.Add([1,2,3,4,5])`
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		panIf(err)
	}

	prg, err := env.Program(ast)
	panIf(err)

	out, _, err := prg.Eval(map[string]any{})
	panIf(err)

	res := out.Value().(int64)
	if res != 15 {
		t.Fatalf("Expected 15 got %d\n", res)
	}
}
