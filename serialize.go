package gomplate

import (
	"fmt"
	"reflect"
	"time"

	"github.com/flanksource/mapstructure"
)

var timeType = reflect.TypeOf(time.Time{})

// Serialize iterates over each key-value pair in the input map
// serializes any struct value to map[string]any.
func Serialize(in map[string]any) (map[string]any, error) {
	if in == nil {
		return nil, nil
	}

	newMap := make(map[string]any, len(in))
	for k, v := range in {
		var dec *mapstructure.Decoder
		var err error

		if v == nil {
			newMap[k] = nil
			continue
		}
		vt := reflect.TypeOf(v)
		if vt.Kind() == reflect.Ptr {
			vt = vt.Elem()
		}

		switch vt.Kind() {

		case timeType.Kind():
			newMap[k] = v
		case reflect.Struct:
			var result map[string]any

			dec, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				TagName: "json",
				Result:  &result,
				Squash:  true,
				Deep:    true})
			if err != nil {
				return nil, fmt.Errorf("error creating new mapstructure decoder: %w", err)
			}

			if err := dec.Decode(v); err != nil {
				return nil, fmt.Errorf("error decoding %T to map[string]any: %w", v, err)
			}

			newMap[k] = result

		case reflect.Slice:
			var result any
			if vt.Elem().Kind() == reflect.Struct {
				result = make([]map[string]any, 0)
			}
			dec, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				TagName: "json",
				Result:  &result,
				Squash:  true,
				Deep:    true})
			if err != nil {
				return nil, fmt.Errorf("error creating new mapstructure decoder: %w", err)
			}

			if err := dec.Decode(v); err != nil {
				return nil, fmt.Errorf("error decoding %T to map[string]any: %w", v, err)
			}
			newMap[k] = result

		default:
			newMap[k] = v
			continue
		}
	}

	return newMap, nil
}
