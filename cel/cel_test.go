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

func TestEnv(t *testing.T) {
	var envopt []cel.EnvOption
	envopt = append(envopt, funcs.ReplaceGen)

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
