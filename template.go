package gomplate

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	gotemplate "text/template"
	"time"

	_ "github.com/flanksource/gomplate/v3/js"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/registry"
	_ "github.com/robertkrimen/otto/underscore"
)

var funcMap gotemplate.FuncMap

var (
	// keep the cache period low as lots of anonymous functions can pile up the cache.
	goTemplateCache    = cache.New(time.Hour, time.Hour)
	celExpressionCache = cache.New(time.Hour, time.Hour)
)

func init() {
	funcMap = CreateFuncs(context.Background())
}

type Template struct {
	Template   string `yaml:"template,omitempty" json:"template,omitempty"` // Go template
	JSONPath   string `yaml:"jsonPath,omitempty" json:"jsonPath,omitempty"`
	Expression string `yaml:"expr,omitempty" json:"expr,omitempty"` // A cel-go expression
	Javascript string `yaml:"javascript,omitempty" json:"javascript,omitempty"`
	RightDelim string `yaml:"-" json:"-"`
	LeftDelim  string `yaml:"-" json:"-"`

	// Pass in additional cel-env options like functions
	// that aren't simple enough to be included in Functions
	CelEnvs []cel.EnvOption `yaml:"-" json:"-"`

	// A map of functions that are accessible to cel expressions
	// and go templates.
	// NOTE: For cel expressions, the functions must be of type func() any.
	// If any other function type is used, an error will be returned.
	// Opt to CelEnvs for those cases.
	Functions map[string]any `yaml:"-" json:"-"`
}

func (t Template) CacheKey(env map[string]any) string {
	envVars := make([]string, 0, len(env)+1)
	for k := range env {
		envVars = append(envVars, k)
	}
	sort.Slice(envVars, func(i, j int) bool { return envVars[i] < envVars[j] })

	return strings.Join(envVars, "-") +
		t.RightDelim +
		t.LeftDelim +
		t.Expression +
		t.Javascript +
		t.JSONPath +
		t.Template
}

func (t Template) IsCacheable() bool {
	// Note: If custom functions are provided then we don't cache the template
	// because it's not possible to uniquely identify a function to be used as a cache key.
	// Pointers don't work well because different functions, that are behaviourly different,
	// but syntatically identical, will have the same pointer value.
	//
	// Reference: https://pkg.go.dev/reflect#Value.Pointer
	// 	> If v's Kind is Func, the returned pointer is an underlying code pointer,
	//  > but not necessarily enough to identify a single function uniquely.
	// 	> The only guarantee is that the result is zero if and only if v is a nil func Value.
	return len(t.CelEnvs) == 0 && len(t.Functions) == 0
}

func (t Template) IsEmpty() bool {
	return t.Template == "" && t.JSONPath == "" && t.Expression == "" && t.Javascript == ""
}

func RunExpression(_environment map[string]any, template Template) (any, error) {
	data, err := Serialize(_environment)
	if err != nil {
		return "", err
	}

	envOptions := GetCelEnv(data)
	for name, fn := range template.Functions {
		_name := name
		_fn := fn
		envOptions = append(envOptions, cel.Function(_name, cel.Overload(
			_name,
			nil,
			cel.AnyType,
			cel.FunctionBinding(func(values ...ref.Val) ref.Val {
				ogFunc, ok := _fn.(func() any)
				if !ok {
					return types.WrapErr(fmt.Errorf("%s is expected to be of type func() any", _name))
				}

				out := ogFunc()
				return types.DefaultTypeAdapter.NativeToValue(out)
			}),
		)))
	}

	envOptions = append(envOptions, template.CelEnvs...)

	var prg cel.Program
	if template.IsCacheable() {
		cached, ok := celExpressionCache.Get(template.CacheKey(_environment))
		if ok {
			if cachedPrg, ok := cached.(*cel.Program); ok {
				prg = *cachedPrg
			}
		}
	}

	if prg == nil {
		env, err := cel.NewEnv(envOptions...)
		if err != nil {
			return "", err
		}

		ast, issues := env.Compile(strings.ReplaceAll(template.Expression, "\n", " "))
		if issues != nil && issues.Err() != nil {
			return "", issues.Err()
		}

		prg, err = env.Program(ast, cel.Globals(data))
		if err != nil {
			return "", err
		}

		celExpressionCache.SetDefault(template.CacheKey(_environment), &prg)
	}

	out, _, err := prg.Eval(data)
	if err != nil {
		return nil, errors.Wrapf(err, "error evaluating expression %s: %s", template.Expression, err)
	}
	return out.Value(), nil

}

func RunTemplate(environment map[string]any, template Template) (string, error) {
	// javascript
	if template.Javascript != "" {
		vm := otto.New()
		for k, v := range environment {
			if err := vm.Set(k, v); err != nil {
				return "", fmt.Errorf("error setting %s: %w", k, err)
			}
		}

		out, err := vm.Run(template.Javascript)
		if err != nil {
			return "", fmt.Errorf("failed to run javascript: %w", err)
		}

		if s, err := out.ToString(); err != nil {
			return "", fmt.Errorf("failed to cast output to string: %w", err)
		} else {
			return s, nil
		}
	}

	// gotemplate
	if template.Template != "" {
		return goTemplate(template, environment)
	}

	// cel-go
	if template.Expression != "" {
		out, err := RunExpression(environment, template)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", out), nil
	}

	return "", nil
}

func goTemplate(template Template, environment map[string]any) (string, error) {
	var tpl *gotemplate.Template

	if template.IsCacheable() {
		cached, ok := goTemplateCache.Get(template.CacheKey(nil))
		if ok {
			if cachedTpl, ok := cached.(*gotemplate.Template); ok {
				tpl = cachedTpl
			}
		}
	}

	if tpl == nil {
		template, err := parseAndStripTemplateHeader(template)
		if err != nil {
			return "", err
		}

		tpl = gotemplate.New("")
		if template.LeftDelim != "" {
			tpl = tpl.Delims(template.LeftDelim, template.RightDelim)
		}

		funcs := make(map[string]any)
		for k, v := range funcMap {
			funcs[k] = v
		}
		for k, v := range template.Functions {
			funcs[k] = v
		}

		tpl, err = tpl.Funcs(funcs).Parse(template.Template)
		if err != nil {
			return "", err
		}

		goTemplateCache.SetDefault(template.CacheKey(nil), tpl)
	}

	data, err := Serialize(environment)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template %s: %v", strings.Split(template.Template, "\n")[0], err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// LoadSharedLibrary loads a shared library for Otto
func LoadSharedLibrary(source string) error {
	source = strings.TrimSpace(source)
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("failed to read shared library %s: %s", source, err)
	}

	fmt.Printf("Loaded %s: \n%s\n", source, string(data))
	registry.Register(func() string { return string(data) })
	return nil
}

func parseAndStripTemplateHeader(template Template) (Template, error) {
	fm, content := extractFrontmatterAndContent(template.Template)
	if fm == "" {
		return template, nil
	}

	template.Template = content
	scanner := bufio.NewScanner(strings.NewReader(fm))
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			return template, fmt.Errorf("invalid header: %s", line)
		}

		switch split[0] {
		case "right-delim":
			template.RightDelim = split[1]
		case "left-delim":
			template.LeftDelim = split[1]
		}
	}

	return template, nil
}

func extractFrontmatterAndContent(input string) (string, string) {
	const delimiter = "---"

	input = strings.TrimSpace(input)

	if !strings.HasPrefix(input, delimiter) {
		return "", input
	}

	endIndex := strings.Index(input[len(delimiter):], delimiter)
	if endIndex == -1 {
		return "", input
	}

	frontmatterEndIndex := endIndex + 2*len(delimiter)
	frontmatter := strings.TrimSpace(input[len(delimiter) : endIndex+len(delimiter)])
	content := strings.TrimSpace(input[frontmatterEndIndex:])

	return frontmatter, content
}
