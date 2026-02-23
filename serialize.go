package gomplate

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ohler55/ojg"
	"github.com/ohler55/ojg/alt"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

var opts = oj.Options{
	Color:        false,
	InitSize:     256,
	CreateKey:    "",
	FullTypePath: false,
	OmitNil:      false,
	OmitEmpty:    false,
	UseTags:      true,
	KeyExact:     true,
	NestEmbed:    false,

	BytesAs:    ojg.BytesAsString,
	TimeFormat: "time",
	WriteLimit: 1024,
}

type AsMapper interface {
	AsMap(fields ...string) map[string]any
}

// Serialize iterates over each key-value pair in the input map
// serializes any struct value to map[string]any.
func Serialize(in map[string]any) (out map[string]any, err error) {
	if in == nil {
		return nil, nil
	}

	defer func() {
		if r := recover(); r != nil {
			if _err, ok := r.(error); ok {
				err = _err
			}
			err = fmt.Errorf("panic during serialization: %v", r)
		}
	}()

	// cel supports time.Duration natively - save original and then replace it after decomposition
	// FIXME: This does not work for anything inside Structs
	nativeTypes := make(map[string]any, len(in))
	jp.Walk(in, func(path jp.Expr, value any) {
		switch v := value.(type) {
		case AsMapper:
			nativeTypes[path.String()] = v.AsMap()
		case uuid.UUID:
			nativeTypes[path.String()] = v.String()
		case *uuid.UUID:
			nativeTypes[path.String()] = v.String()
		case time.Duration:
			nativeTypes[path.String()] = v
		}
	})

	out = alt.Alter(in, &opts).(map[string]any)

	for path, v := range nativeTypes {
		expr, err := jp.ParseString(path)
		if err != nil {
			return nil, err
		}
		if err := expr.SetOne(out, v); err != nil {
			return nil, err
		}
	}
	return out, nil
}
