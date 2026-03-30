package gomplate

import (
	gocontext "context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/recolabs/gnata"
)

var jsonataExpressionCache = cache.New(time.Hour, time.Hour)

// gnataBuiltinFunctions is the set of function names natively provided by gnata/JSONata.
// These take precedence over custom gomplate functions; any custom function whose name
// matches an entry here is silently skipped during registration.
var gnataBuiltinFunctions = map[string]struct{}{
	"string": {}, "length": {}, "substring": {}, "substringBefore": {},
	"substringAfter": {}, "trim": {}, "pad": {}, "contains": {},
	"split": {}, "join": {}, "base64encode": {}, "base64decode": {},
	"encodeUrl": {}, "encodeUrlComponent": {}, "decodeUrl": {}, "decodeUrlComponent": {},
	"formatNumber": {}, "formatBase": {}, "formatInteger": {}, "parseInteger": {},
	"number": {}, "abs": {}, "floor": {}, "ceil": {}, "round": {}, "power": {},
	"sqrt": {}, "random": {}, "sum": {}, "max": {}, "min": {}, "average": {},
	"count": {}, "append": {}, "reverse": {}, "shuffle": {}, "distinct": {},
	"flatten": {}, "zip": {},
	"keys": {}, "values": {}, "spread": {}, "merge": {}, "error": {},
	"lookup":  {},
	"boolean": {}, "not": {}, "exists": {}, "assert": {}, "type": {},
	"now": {}, "millis": {}, "fromMillis": {}, "toMillis": {},
	"uppercase": {}, "lowercase": {}, "match": {}, "replace": {},
	"eval": {}, "sort": {}, "sift": {}, "each": {}, "map": {}, "filter": {},
	"single": {}, "reduce": {},
}

// customJSONataEnv is the shared, reusable gnata environment that combines
// JSONata's native built-ins with gomplate-specific custom functions.
// Built-ins always take precedence: any custom function whose name conflicts
// with a gnata built-in is skipped during construction (see getJSONataCustomFuncs).
// The type is inferred from gnata.NewCustomEnv which returns an internal evaluator.Environment.
var customJSONataEnv = gnata.NewCustomEnv(map[string]gnata.CustomFunc{}) //nolint:gochecknoglobals

func init() {
	customJSONataEnv = gnata.NewCustomEnv(getJSONataCustomFuncs())
}

// getJSONataCustomFuncs returns a map of gomplate-specific functions for use in
// JSONata expressions. Functions whose names match gnata built-ins are excluded
// so that the native implementation always takes precedence when there is a name
// collision.
func getJSONataCustomFuncs() map[string]gnata.CustomFunc {
	allFuncs := jsonataCustomFuncs()

	result := make(map[string]gnata.CustomFunc, len(allFuncs))
	for name, fn := range allFuncs {
		if _, isBuiltin := gnataBuiltinFunctions[name]; !isBuiltin {
			result[name] = fn
		}
	}
	return result
}

// RunJSONata evaluates a JSONata expression against the supplied environment.
// The environment keys become the top-level fields of the input document
// (identical to how CEL variables are exposed).
// Results are normalised to standard Go types via gnata.NormalizeValue.
func RunJSONata(environment map[string]any, expression string) (any, error) {
	cacheKey := expression
	if compiled, ok := jsonataExpressionCache.Get(cacheKey); ok {
		expr := compiled.(*gnata.Expression)
		result, err := expr.EvalWithCustomFuncs(gocontext.Background(), environment, customJSONataEnv)
		if err != nil {
			return nil, fmt.Errorf("jsonata: %w", err)
		}
		return gnata.NormalizeValue(result), nil
	}

	expr, err := gnata.Compile(expression)
	if err != nil {
		return nil, fmt.Errorf("jsonata: compile error: %w", err)
	}

	jsonataExpressionCache.SetDefault(cacheKey, expr)

	result, err := expr.EvalWithCustomFuncs(gocontext.Background(), environment, customJSONataEnv)
	if err != nil {
		return nil, fmt.Errorf("jsonata: %w", err)
	}
	return gnata.NormalizeValue(result), nil
}

// jsonataToString converts a JSONata result to its string representation,
// matching the behaviour of RunTemplateContext for other template types.
func jsonataToString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case json.Number:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}
