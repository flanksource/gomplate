package gomplate

import (
	"encoding/json"
)

// Serialize iterates over each key-value pair in the input map
// serializes any struct value to map[string]any.
func Serialize(in map[string]any) (map[string]any, error) {

	if data, err := json.Marshal(in); err != nil {
		return nil, err
	} else {
		var out = make(map[string]any)
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
}
