package nilsafe

import (
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

func TestNilSafe(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		vars    map[string]any
		want    ref.Val
		wantErr string
	}{
		{
			name: "missing variable returns null",
			expr: "x",
			vars: map[string]any{},
			want: types.NullValue,
		},
		{
			name: "null field access returns null",
			expr: "a.field",
			vars: map[string]any{"a": nil},
			want: types.NullValue,
		},
		{
			name: "chained null access returns null",
			expr: "a.b.c.d",
			vars: map[string]any{"a": map[string]any{"b": nil}},
			want: types.NullValue,
		},
		{
			name: "missing map key returns null",
			expr: `m["missing"]`,
			vars: map[string]any{"m": map[string]any{"exists": 1}},
			want: types.NullValue,
		},
		{
			name: "out of bounds list index returns null",
			expr: "items[99]",
			vars: map[string]any{"items": []int{1, 2, 3}},
			want: types.NullValue,
		},
		{
			name: "method on null returns null",
			expr: "a.size()",
			vars: map[string]any{"a": nil},
			want: types.NullValue,
		},
		{
			name: "null arithmetic returns null",
			expr: "a + 1",
			vars: map[string]any{"a": nil},
			want: types.NullValue,
		},
		{
			name: "null comparison returns null",
			expr: "a > 0",
			vars: map[string]any{"a": nil},
			want: types.NullValue,
		},
		{
			name:    "division by zero preserved",
			expr:    "1 / 0",
			vars:    map[string]any{},
			wantErr: "division by zero",
		},
		{
			name:    "type conversion error preserved",
			expr:    `int("bad")`,
			vars:    map[string]any{},
			wantErr: "type conversion error",
		},
		{
			name:    "regex error preserved",
			expr:    `"x".matches("[")`,
			vars:    map[string]any{},
			wantErr: "error parsing regexp",
		},
		{
			name: "null equals null is true",
			expr: "a == null",
			vars: map[string]any{"a": nil},
			want: types.True,
		},
		{
			name: "null not equals value is true",
			expr: "a != 1",
			vars: map[string]any{"a": nil},
			want: types.True,
		},
		{
			name: "present variable works normally",
			expr: "x + 1",
			vars: map[string]any{"x": int64(41)},
			want: types.Int(42),
		},
		{
			name: "mixed present and missing",
			expr: "x",
			vars: map[string]any{"y": int64(1)},
			want: types.NullValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				Library(),
				cel.Variable("x", cel.DynType),
				cel.Variable("y", cel.DynType),
				cel.Variable("a", cel.DynType),
				cel.Variable("m", cel.DynType),
				cel.Variable("items", cel.DynType),
			)
			if err != nil {
				t.Fatalf("cel.NewEnv() failed: %v", err)
			}
			ast, iss := env.Compile(tt.expr)
			if iss.Err() != nil {
				t.Fatalf("env.Compile(%q) failed: %v", tt.expr, iss.Err())
			}
			prg, err := env.Program(ast)
			if err != nil {
				t.Fatalf("env.Program() failed: %v", err)
			}
			out, _, err := prg.Eval(tt.vars)
			if tt.wantErr != "" {
				if err == nil && !types.IsError(out) {
					t.Fatalf("expected error containing %q, got result: %v", tt.wantErr, out)
				}
				errMsg := ""
				if err != nil {
					errMsg = err.Error()
				} else {
					errMsg = out.(*types.Err).Error()
				}
				if !strings.Contains(errMsg, tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %s", tt.wantErr, errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.Equal(tt.want) != types.True {
				t.Errorf("got %v (%T), want %v (%T)", out, out, tt.want, tt.want)
			}
		})
	}
}

func TestNilSafe_HasMacro(t *testing.T) {
	env, err := cel.NewEnv(
		Library(),
		cel.Variable("obj", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		t.Fatalf("cel.NewEnv() failed: %v", err)
	}
	ast, iss := env.Compile(`has(obj.field)`)
	if iss.Err() != nil {
		t.Fatalf("Compile failed: %v", iss.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		t.Fatalf("Program failed: %v", err)
	}

	out, _, err := prg.Eval(map[string]any{
		"obj": map[string]any{"field": "value"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != types.True {
		t.Errorf("has(obj.field) = %v, want true", out)
	}

	out, _, err = prg.Eval(map[string]any{
		"obj": map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != types.False {
		t.Errorf("has(obj.field) on missing = %v, want false", out)
	}
}

func TestNilSafe_ZeroValues(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		vars    map[string]any
		want    ref.Val
		wantErr string
	}{
		{name: "null == 0 is true", expr: "a == 0", vars: map[string]any{"a": nil}, want: types.True},
		{name: "null == false is true", expr: "a == false", vars: map[string]any{"a": nil}, want: types.True},
		{name: "null == empty string is true", expr: `a == ""`, vars: map[string]any{"a": nil}, want: types.True},
		{name: "null != 1 is true", expr: "a != 1", vars: map[string]any{"a": nil}, want: types.True},
		{name: "null != 0 is false", expr: "a != 0", vars: map[string]any{"a": nil}, want: types.False},
		{name: "null + 1", expr: "a + 1", vars: map[string]any{"a": nil}, want: types.Int(1)},
		{name: "null - 1", expr: "a - 1", vars: map[string]any{"a": nil}, want: types.Int(-1)},
		{name: "null * 5", expr: "a * 5", vars: map[string]any{"a": nil}, want: types.IntZero},
		{name: "null > 0 is false", expr: "a > 0", vars: map[string]any{"a": nil}, want: types.False},
		{name: "null < 1 is true", expr: "a < 1", vars: map[string]any{"a": nil}, want: types.True},
		{name: "null >= 0 is true", expr: "a >= 0", vars: map[string]any{"a": nil}, want: types.True},
		{name: "null <= 0 is true", expr: "a <= 0", vars: map[string]any{"a": nil}, want: types.True},
		{name: "method on null still returns null", expr: "a.size()", vars: map[string]any{"a": nil}, want: types.NullValue},
		{name: "present variable works normally", expr: "x + 1", vars: map[string]any{"x": int64(41)}, want: types.Int(42)},
		{name: "both present comparison", expr: "x > 0", vars: map[string]any{"x": int64(5)}, want: types.True},
		{name: "division by zero preserved", expr: "1 / 0", vars: map[string]any{}, wantErr: "division by zero"},
		{name: "null == null is true", expr: "a == b", vars: map[string]any{"a": nil, "b": nil}, want: types.True},
		{name: "null != null is false", expr: "a != b", vars: map[string]any{"a": nil, "b": nil}, want: types.False},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := cel.NewEnv(
				Library(WithZeroValues()),
				cel.Variable("x", cel.DynType),
				cel.Variable("a", cel.DynType),
				cel.Variable("b", cel.DynType),
			)
			if err != nil {
				t.Fatalf("cel.NewEnv() failed: %v", err)
			}
			ast, iss := env.Compile(tt.expr)
			if iss.Err() != nil {
				t.Fatalf("env.Compile(%q) failed: %v", tt.expr, iss.Err())
			}
			prg, err := env.Program(ast)
			if err != nil {
				t.Fatalf("env.Program() failed: %v", err)
			}
			out, _, err := prg.Eval(tt.vars)
			if tt.wantErr != "" {
				if err == nil && !types.IsError(out) {
					t.Fatalf("expected error containing %q, got result: %v", tt.wantErr, out)
				}
				errMsg := ""
				if err != nil {
					errMsg = err.Error()
				} else {
					errMsg = out.(*types.Err).Error()
				}
				if !strings.Contains(errMsg, tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %s", tt.wantErr, errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.Equal(tt.want) != types.True {
				t.Errorf("got %v (%T), want %v (%T)", out, out, tt.want, tt.want)
			}
		})
	}
}

func TestWithoutLibrary_ErrorsReturned(t *testing.T) {
	env, err := cel.NewEnv(
		cel.Variable("x", cel.DynType),
	)
	if err != nil {
		t.Fatalf("cel.NewEnv() failed: %v", err)
	}
	ast, iss := env.Compile("x")
	if iss.Err() != nil {
		t.Fatalf("Compile failed: %v", iss.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		t.Fatalf("Program failed: %v", err)
	}
	_, _, err = prg.Eval(map[string]any{})
	if err == nil {
		t.Errorf("expected error for missing variable without library, got nil")
	}
}
