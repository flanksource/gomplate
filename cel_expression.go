package gomplate

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gomplate/v4/celvalues"
	"github.com/patrickmn/go-cache"
	"github.com/samber/oops"
)

var celExpressionCache = cache.New(time.Hour, time.Hour)

func RunExpression(environment map[string]any, template Template) (any, error) {
	return RunExpressionContext(newContext(), environment, template)
}

func RunExpressionContext(ctx commonsContext.Context, environment map[string]any, template Template) (any, error) {
	tracker := celTrackerFromContext(ctx)
	if tracker != nil {
		if err := tracker.begin(); err != nil {
			return nil, err
		}
		defer tracker.abort()
	}

	functions := template.celFunctions()
	impls, err := celFunctionImpls(functions)
	if err != nil {
		return nil, oops.With("template", template.Expression).Wrap(err)
	}
	nativeTypes := currentNativeTypes()
	data := celInput(environment, nativeTypes)
	cacheKey := template.celCacheKey(environment, nativeTypes.generation, functions)

	var program cel.Program
	var ast *cel.Ast
	if tracker == nil && template.IsCacheable() {
		cached, found := celExpressionCache.Get(cacheKey)
		if found {
			if cachedProgram, ok := cached.(*cel.Program); ok {
				program = *cachedProgram
			}
		}
	}

	if program == nil {
		program, ast, err = compileCELProgram(data, template, nativeTypes, tracker != nil)
		if err != nil {
			return "", err
		}
		if tracker == nil && template.IsCacheable() {
			celExpressionCache.Set(cacheKey, &program, template.CacheTime)
		}
	}

	var activation any = data
	if impls != nil {
		if activation, err = cel.FunctionVars(data, impls); err != nil {
			return nil, oops.With("template", template.Expression).Wrap(err)
		}
	}
	out, details, err := program.Eval(activation)
	if tracker != nil {
		tracker.complete(ast, details, out)
	}
	if err != nil {
		return nil, oops.With("template", template.Expression).Wrap(err)
	}
	if ctx.Logger != nil && out.Value() != template.Expression && properties.On(false, "gomplate.log") {
		ctx.Logger.V(4).Infof("templated %s => %v", template.ShortString(), out)
	}
	return CELValue(out), nil
}

// celInput presents the environment to CEL without copying it: each value is wrapped in its
// CEL form, and the maps, lists and structs beneath it are converted only as an expression
// reads them. Registered native types are left for the environment's native type provider.
func celInput(environment map[string]any, nativeTypes *nativeTypeSnapshot) map[string]any {
	var data map[string]any
	adapter := celvalues.NewAdapter(types.DefaultTypeAdapter)
	for key, value := range environment {
		if nativeTypes.preserves(value) {
			continue
		}
		if data == nil {
			data = maps.Clone(environment)
		}
		data[key] = adapter.NativeToValue(value)
	}
	if data == nil {
		return environment // every value is a registered native type, or there are none
	}
	return data
}

func compileCELProgram(data map[string]any, template Template, nativeTypes *nativeTypeSnapshot, trackState bool) (cel.Program, *cel.Ast, error) {
	env, err := extendedCelEnv(data, template, nativeTypes)
	if err != nil {
		return nil, nil, err
	}
	expression := strings.ReplaceAll(template.Expression, "\n", " ")
	if trackState {
		expression = template.Expression
	}
	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, nil, oops.With("template", template.Expression).Errorf("issues: %s", issues.String())
	}

	// OptOptimize folds constants and precompiles regexes while the program is
	// built. That makes this function -- the cache=miss path -- allocate a few
	// percent more, which is the trade being bought: a program is compiled once
	// and then evaluated from celExpressionCache for an hour, so the work moves
	// off the hot path. Do not "fix" a compile-path allocation regression here by
	// dropping it.
	evalOptions := []cel.EvalOption{cel.OptOptimize}
	if trackState {
		evalOptions = append(evalOptions, cel.OptTrackState)
	}
	program, err := env.Program(ast, cel.EvalOptions(evalOptions...))
	if err != nil {
		return nil, nil, err
	}
	return program, ast, nil
}

var (
	extendedCelEnvCache = cache.New(time.Hour, time.Hour)
	// extendedCelEnvBuilds counts the environments built for the cache, for its specs.
	extendedCelEnvBuilds atomic.Int64
)

// extendedCelEnv layers the per-call declarations over the base environment. Declarations
// alone decide the result, so it is cached by them; CelEnvs are opaque and never cached.
func extendedCelEnv(data map[string]any, template Template, nativeTypes *nativeTypeSnapshot) (*cel.Env, error) {
	functions := template.celFunctions()
	base, err := baseCelEnv()
	if template.Strict {
		base, err = strictBaseCelEnv()
	}
	if err != nil {
		return nil, err
	}
	options := celEnvOptions(data, template, nativeTypes, functions)
	if len(template.CelEnvs) > 0 {
		return base.Extend(options...)
	}

	key := celEnvCacheKey(data, template.Strict, nativeTypes.generation, functions)
	if cached, found := extendedCelEnvCache.Get(key); found {
		return cached.(*cel.Env), nil
	}
	env, err := base.Extend(options...)
	if err != nil {
		return nil, err
	}
	extendedCelEnvBuilds.Add(1)
	extendedCelEnvCache.SetDefault(key, env)
	return env, nil
}

func celEnvCacheKey(data map[string]any, strict bool, nativeTypeGeneration uint64, functions []CelFunction) string {
	names := slices.Sorted(maps.Keys(data))
	return fmt.Sprintf("%t:%d:%s:%s", strict, nativeTypeGeneration, strings.Join(names, ","), celFunctionSignature(functions))
}

func celEnvOptions(data map[string]any, template Template, nativeTypes *nativeTypeSnapshot, functions []CelFunction) []cel.EnvOption {
	envOptions := make([]cel.EnvOption, 0, len(data)+len(functions)+len(template.CelEnvs)+1)
	if nativeTypes.envOption != nil {
		envOptions = append(envOptions, nativeTypes.envOption)
	}
	for key := range data {
		envOptions = append(envOptions, cel.Variable(key, cel.DynType))
	}
	envOptions = append(envOptions, celFunctionDeclarations(functions)...)
	envOptions = append(envOptions, template.CelEnvs...)
	return envOptions
}

// celCacheKey keys the program cache. The auto-derived key includes the function declarations,
// so two templates that differ only in them never share a program; an explicit CacheKey is the
// caller's assertion and is used as given.
func (t Template) celCacheKey(environment map[string]any, nativeTypeGeneration uint64, functions []CelFunction) string {
	profile := ""
	if t.Strict {
		profile = "strict:"
	}
	if nativeTypeGeneration != 0 {
		profile += strconv.FormatUint(nativeTypeGeneration, 10) + ":"
	}
	if t.CacheKey != "" || len(functions) == 0 {
		return profile + t.cacheKey(environment)
	}
	return profile + t.cacheKey(environment) + "\x00" + celFunctionSignature(functions)
}
