package gcp

import (
	"encoding/json"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

func IncidentToCheckResult(in any) map[string]any {
	var obj map[string]any
	switch v := in.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &obj); err != nil {
			return nil
		}
	case map[string]any:
		obj = v
	default:
		return nil
	}

	checkResult := map[string]any{
		"name":        fmt.Sprintf("[%s] %s", obj["incident_id"], obj["summary"]),
		"pass":        false,
		"details":     obj,
		"description": obj["summary"],
		"message":     obj["summary"],
	}
	return checkResult
}

func gcpIncidentToCheckResult(fnName string) cel.EnvOption {
	return cel.Function(fnName,
		cel.Overload(fnName+"_overload",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				return types.NewDynamicMap(types.DefaultTypeAdapter, IncidentToCheckResult(obj.Value()))
			}),
		),
	)
}
