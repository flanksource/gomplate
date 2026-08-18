package funcs

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// The conversions CEL's own are too strict for.
//
// `string()` and `int()` are exact: each has a fixed overload set, so `string()`
// refuses anything it was not told about and neither reads a value whose type is
// only known at runtime. That is right for a language and wrong for templating
// over scraped data, where a property arrives as whatever the scraper made of it
// and the template's job is to say what it means. `date` has no equivalent at
// all -- CEL can parse a timestamp but not render one.
//
// Each is declared twice, globally and as a member, following celFirst in
// coll.go: `text(x)` reads better as an argument and `x.text()` reads better in
// a chain, and which one an author reaches for is not worth a rule.
//
// `int` is the exception: the global belongs to the standard library already, so
// only the member form is added and the two compose as one function.

// castDateLayout is the calendar day, the unit a register or a report means when
// it says a date. Deliberately not RFC3339: a document stating a finding was
// detected on the 24th must mean the same day to whoever reads it, which is why
// every conversion below normalises to UTC first.
const castDateLayout = "2006-01-02"

var celText = cel.Function("text",
	cel.Overload("text_dyn", []*cel.Type{cel.DynType}, cel.StringType,
		cel.UnaryBinding(celTextImpl)),
	cel.MemberOverload("dyn_text", []*cel.Type{cel.DynType}, cel.StringType,
		cel.UnaryBinding(celTextImpl)),
)

// celInt has no global overload: the standard library's `int()` already converts
// every typed case, and a second global taking dyn would be ambiguous against
// each of them. The member form is what the standard library has no answer for.
var celInt = cel.Function("int",
	cel.MemberOverload("dyn_int", []*cel.Type{cel.DynType}, cel.IntType,
		cel.UnaryBinding(celIntImpl)),
)

var celFloat = cel.Function("float",
	cel.Overload("float_dyn", []*cel.Type{cel.DynType}, cel.DoubleType,
		cel.UnaryBinding(celFloatImpl)),
	cel.MemberOverload("dyn_float", []*cel.Type{cel.DynType}, cel.DoubleType,
		cel.UnaryBinding(celFloatImpl)),
)

var celDate = cel.Function("date",
	cel.Overload("date_dyn", []*cel.Type{cel.DynType}, cel.StringType,
		cel.UnaryBinding(celDateImpl)),
	cel.MemberOverload("dyn_date", []*cel.Type{cel.DynType}, cel.StringType,
		cel.UnaryBinding(celDateImpl)),
)

func celTextImpl(value ref.Val) ref.Val {
	if isNullVal(value) {
		return types.NewErr("text() does not accept null")
	}
	return types.String(fmt.Sprint(value.Value()))
}

// celIntImpl parses and truncates, because the two things a caller means by
// "as an integer" are a numeric string and a number that is not one yet. A
// value that is neither is an error rather than a zero: a silent zero reads as
// a real count of none.
func celIntImpl(value ref.Val) ref.Val {
	if isNullVal(value) {
		return types.NewErr("int() does not accept null")
	}
	switch typed := value.Value().(type) {
	case int64:
		return types.Int(typed)
	case uint64:
		if typed > math.MaxInt64 {
			return types.NewErr("int() cannot represent %d", typed)
		}
		return types.Int(int64(typed))
	case float64:
		return truncateToInt(typed)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return types.NewErr("int() cannot parse %q", typed)
		}
		return truncateToInt(parsed)
	default:
		return types.NewErr("int() expects a string or numeric value, got %T", value.Value())
	}
}

// truncateToInt rejects what int64 cannot hold instead of converting it.
// Converting a float64 outside int64's range is undefined in Go and yields the
// minimum int64 on amd64, so a byte count past 2^63 would otherwise arrive as a
// large negative number. The upper bound is strictly less than 2^63 because
// math.MaxInt64 is not representable as a float64 and rounds up to exactly that.
func truncateToInt(value float64) ref.Val {
	if math.IsNaN(value) {
		return types.NewErr("int() does not accept NaN")
	}
	truncated := math.Trunc(value)
	if truncated < math.MinInt64 || truncated >= 1<<63 {
		return types.NewErr("int() cannot represent %v", value)
	}
	return types.Int(int64(truncated))
}

func celFloatImpl(value ref.Val) ref.Val {
	if isNullVal(value) {
		return types.NewErr("float() does not accept null")
	}
	switch typed := value.Value().(type) {
	case float64:
		return types.Double(typed)
	case int64:
		return types.Double(float64(typed))
	case uint64:
		return types.Double(float64(typed))
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return types.NewErr("float() cannot parse %q", typed)
		}
		// Go parses "NaN" happily, and JSON has no way to write the result --
		// a value that cannot be serialised is not a number a caller can use.
		if math.IsNaN(parsed) {
			return types.NewErr("float() does not accept NaN")
		}
		return types.Double(parsed)
	default:
		return types.NewErr("float() expects a string or numeric value, got %T", value.Value())
	}
}

// celDateImpl accepts a timestamp or the RFC3339 string one usually arrives as,
// so a caller does not have to wrap every scraped field in timestamp() before
// asking which day it was.
func celDateImpl(value ref.Val) ref.Val {
	if isNullVal(value) {
		return types.NewErr("date() does not accept null")
	}
	switch typed := value.Value().(type) {
	case time.Time:
		return types.String(typed.UTC().Format(castDateLayout))
	case string:
		parsed, err := time.Parse(time.RFC3339, typed)
		if err != nil {
			return types.NewErr("date() expects an RFC3339 timestamp, got %q", typed)
		}
		return types.String(parsed.UTC().Format(castDateLayout))
	default:
		return types.NewErr("date() expects a timestamp or RFC3339 string, got %T", value.Value())
	}
}

// isNullVal covers both shapes a null reaches a binding as. Under the nilsafe
// library a null argument short-circuits the call and these branches are never
// reached, which is why they state the contract for every other environment
// rather than being the primary defence.
func isNullVal(value ref.Val) bool {
	return value == nil || value == types.NullValue || value.Value() == nil
}
