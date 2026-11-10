package celvalues

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
)

var dateType = types.NewObjectType("gomplate.Date", traits.ComparerType)

// Date is a civil calendar day. Construct it with NewDate.
type Date struct{ iso string }

// NewDate accepts an exact YYYY-MM-DD date without a time or timezone.
func NewDate(value string) (Date, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil || parsed.Format(time.DateOnly) != value {
		return Date{}, fmt.Errorf("date must be YYYY-MM-DD, got %q", value)
	}
	return Date{iso: value}, nil
}

func (d Date) String() string               { return d.iso }
func (d Date) Value() any                   { return d }
func (d Date) Type() ref.Type               { return dateType }
func (d Date) MarshalJSON() ([]byte, error) { return json.Marshal(d.iso) }

func (d Date) ConvertToNative(target reflect.Type) (any, error) {
	return scalarNative(d, target)
}

func (d Date) ConvertToType(target ref.Type) ref.Val {
	if d.iso == "" {
		return types.NewErr("uninitialized Date; use NewDate")
	}
	switch target {
	case dateType:
		return d
	case types.TypeType:
		return dateType
	case types.StringType:
		return types.String(d.iso)
	default:
		return types.NewErr("cannot convert Date to %s", target.TypeName())
	}
}

func (d Date) Compare(other ref.Val) ref.Val {
	if d.iso == "" {
		return types.NewErr("uninitialized Date; use NewDate")
	}
	day, err := dateOperand(other)
	if err != nil {
		return types.WrapErr(err)
	}
	return types.Int(strings.Compare(d.iso, day.iso))
}

func (d Date) Equal(other ref.Val) ref.Val {
	if other == types.NullValue {
		return types.False
	}
	return comparisonVerdict("_==_", d.Compare(other))
}

func dateOperand(value ref.Val) (Date, error) {
	switch typed := value.(type) {
	case Date:
		if typed.iso == "" {
			return Date{}, fmt.Errorf("uninitialized Date; use NewDate")
		}
		return typed, nil
	case types.String:
		return NewDate(string(typed))
	case types.Timestamp:
		return NewDate(typed.Time.UTC().Format(time.DateOnly))
	default:
		return Date{}, fmt.Errorf("Date comparison expects a date, ISO date string or timestamp, got %s", value.Type().TypeName())
	}
}
