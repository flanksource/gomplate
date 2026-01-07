package gomplate

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	gotemplate "text/template"
	"time"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	_ "github.com/flanksource/gomplate/v3/js"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/patrickmn/go-cache"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/registry"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/samber/oops"
	"google.golang.org/protobuf/types/known/structpb"
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

// ListAllFuncs returns the sorted list of built-in template function names.
func ListAllFuncs() []string {
	names := make(map[string]struct{}, len(funcMap))
	for name, fn := range funcMap {
		if isNamespaceFactory(fn) {
			addNamespaceFuncs(names, name, fn)
			continue
		}
		names[name] = struct{}{}
	}

	out := make([]string, 0, len(names))
	for name := range names {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func addNamespaceFuncs(names map[string]struct{}, namespace string, fn any) {
	val := reflect.ValueOf(fn)
	if val.Kind() != reflect.Func {
		return
	}
	typ := val.Type()
	if typ.NumIn() != 0 || typ.NumOut() != 1 {
		return
	}
	if typ.Out(0).Kind() != reflect.Interface {
		return
	}

	var res reflect.Value
	func() {
		defer func() {
			if recover() != nil {
				res = reflect.Value{}
			}
		}()
		res = val.Call(nil)[0]
	}()
	if !res.IsValid() {
		return
	}
	if res.Kind() == reflect.Interface {
		if res.IsNil() {
			return
		}
		res = res.Elem()
	}

	resType := res.Type()
	for i := 0; i < resType.NumMethod(); i++ {
		method := resType.Method(i)
		if method.PkgPath != "" {
			continue
		}
		names[namespace+"."+method.Name] = struct{}{}
	}
}

func isNamespaceFactory(fn any) bool {
	val := reflect.ValueOf(fn)
	if val.Kind() != reflect.Func {
		return false
	}
	typ := val.Type()
	return typ.NumIn() == 0 && typ.NumOut() == 1 && typ.Out(0).Kind() == reflect.Interface
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

func (t Template) String() string {
	if t.Template != "" {
		return "gotemplate: " + t.Template
	}
	if t.Expression != "" {
		return "cel: " + t.Expression
	}
	if t.Javascript != "" {
		return "js: " + t.Javascript
	}
	if t.JSONPath != "" {
		return "jsonpath: " + t.JSONPath
	}
	return ""
}

func (t Template) ShortString() string {
	if t.Template != "" {
		return "gotemplate: " + short(t.Template)
	}
	if t.Expression != "" {
		return "cel: " + short(t.Expression)
	}
	if t.Javascript != "" {
		return "js: " + short(t.Javascript)
	}
	if t.JSONPath != "" {
		return "jsonpath: " + short(t.JSONPath)
	}
	return ""
}

func short(v string) string {
	v = strings.TrimSpace(v)
	if len(v) == 0 {
		return ""
	}
	lines := strings.Split(v, "\n")
	if len(lines) == 1 {
		return lines[0]
	}
	return fmt.Sprintf("%s .. %d more lines", lines[0], len(lines)-1)
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
	return RunExpressionContext(newContext(), _environment, template)
}

func RunExpressionContext(ctx commonsContext.Context, _environment map[string]any, template Template) (any, error) {
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
				if ctx.Logger != nil {
					ctx.Logger.V(7).Infof("%s using cached cel program", template.ShortString())
				}
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
			return "", oops.With("template", template.Expression).Errorf("issues: %s", issues.String())
		}

		prg, err = env.Program(ast, cel.Globals(data))
		if err != nil {
			return "", err
		}

		celExpressionCache.SetDefault(template.CacheKey(_environment), &prg)
	}

	out, _, err := prg.Eval(data)
	if err != nil {
		return nil, oops.With("template", template.Expression).Wrap(err)
	}
	if ctx.Logger != nil && out.Value() != template.Expression {
		ctx.Logger.V(6).Infof("templated %s => %v", template.ShortString(), out)
	}
	return out.Value(), nil

}

func newContext() commonsContext.Context {
	return commonsContext.NewContext(context.TODO(),
		commonsContext.WithLogger(logger.GetLogger("gomplate")))
}

func RunTemplateBool(environment map[string]any, template Template) (bool, error) {
	output, err := RunTemplateContext(newContext(), environment, template)
	if err != nil {
		return false, err
	}

	result, err := strconv.ParseBool(output)
	if err != nil {
		return false, fmt.Errorf("failed to parse template output (%s) as bool: %w", output, err)
	}

	return result, nil
}

func RunTemplate(environment map[string]any, template Template) (string, error) {
	return RunTemplateContext(newContext(), environment, template)
}

func RunTemplateContext(ctx commonsContext.Context, environment map[string]any, template Template) (string, error) {
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
		return goTemplate(ctx, template, environment)
	}

	// cel-go
	if template.Expression != "" {
		out, err := RunExpressionContext(ctx, environment, template)
		if err != nil {
			return "", err
		}
		if _, ok := out.(structpb.NullValue); ok || out == nil {
			return "", nil
		}
		return fmt.Sprintf("%v", out), nil
	}

	return "", nil
}

func goTemplate(ctx commonsContext.Context, template Template, environment map[string]any) (string, error) {
	var tpl *gotemplate.Template

	if template.IsCacheable() {
		cached, ok := goTemplateCache.Get(template.CacheKey(nil))
		if ok {
			if cachedTpl, ok := cached.(*gotemplate.Template); ok {
				if ctx.Logger != nil {
					ctx.Logger.V(7).Infof("%s using cached template", template.ShortString())
				}
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
			return "", oops.With("template", template.Template).Wrap(err)
		}

		goTemplateCache.SetDefault(template.CacheKey(nil), tpl)
	}

	data, err := Serialize(environment)
	if err != nil {
		return "", oops.
			// With("environment", environment)
			Wrapf(err, "error serializing env")
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", oops.
			With("template", template.Template).
			// With("environment", environment).
			Wrap(err)
	}

	out := strings.TrimSpace(buf.String())
	if ctx.Logger != nil && out != template.Template {
		ctx.Logger.V(6).Infof("templated %s ==> %s", template.ShortString(), out)
	}
	return out, nil
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
	header, content := extractHeaderAndContent(template.Template)
	if header == "" {
		return template, nil
	}

	template.Template = content

	fields := strings.Fields(header)
	for _, field := range fields {
		split := strings.SplitN(field, "=", 2)
		if len(split) != 2 {
			return template, fmt.Errorf("invalid header: %s", field)
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

const templateHeaderPrefix = "# gotemplate: "

func extractHeaderAndContent(template string) (string, string) {
	scanner := bufio.NewScanner(strings.NewReader(template))

	// Loop through headers.
	// There could be multiple, we look for the gotemplate header.
	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" {
			// Special case for yaml where the header might not start from the first line.
			continue
		}

		// end of headers.
		isHeader := strings.HasPrefix(line, "#")
		if !isHeader {
			break
		}

		if strings.HasPrefix(line, templateHeaderPrefix) {
			header := strings.TrimPrefix(line, templateHeaderPrefix)
			return header, strings.Replace(template, fmt.Sprintf("%s\n", line), "", 1)
		}
	}

	return "", template
}
