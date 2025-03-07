package kubernetes

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/is-healthy/pkg/health"
	"github.com/flanksource/is-healthy/pkg/lua"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Health  string `json:"health"`
	Ready   bool   `json:"ready"`

	// Deprecated: Use Health.
	OK bool `json:"ok"`
}

func GetUnstructuredMap(in interface{}) []byte {
	x := GetUnstructured(in)
	b, _ := json.Marshal(x)
	return b
}

func GetUnstructured(in interface{}) *unstructured.Unstructured {
	var err error
	obj := make(map[string]interface{})

	switch v := in.(type) {
	case string:
		err = yaml.Unmarshal([]byte(v), &obj)
	case []byte:
		err = yaml.Unmarshal(v, &obj)
	case types.Bytes:
		err = yaml.Unmarshal(v, &obj)
	case map[string]interface{}:
		if val, ok := v["Object"].(map[string]any); ok {
			obj = val
		} else {
			obj = v
		}
	case unstructured.Unstructured:
		obj = v.Object
	default:
		var data []byte
		if data, err = yaml.Marshal(in); err == nil {
			err = yaml.Unmarshal(data, &obj)
		}
	}

	if err != nil {
		return nil
	}

	return &unstructured.Unstructured{Object: obj}
}

func IsReady(in any) bool {
	return GetHealth(in).Ready
}

func IsHealthy(in interface{}) bool {
	return GetHealth(in).Health == string(health.HealthHealthy)
}

func GetStatus(in interface{}) string {
	health := GetHealth(in)
	return fmt.Sprintf("%s: %s", health.Status, health.Message)
}

func GetHealth(in interface{}) HealthStatus {
	var err error
	obj := GetUnstructured(in)

	if obj == nil {
		return HealthStatus{
			OK:      false,
			Health:  "unhealthy",
			Status:  "Error",
			Message: "Invalid spec",
		}
	}

	_health, err := health.GetResourceHealth(obj, lua.ResourceHealthOverrides{})
	if err != nil {
		return HealthStatus{
			OK:      false,
			Status:  "Error",
			Message: err.Error(),
		}
	}

	if _health == nil {
		return HealthStatus{
			OK:     false,
			Health: "unknown",
			Ready:  true,
		}
	}

	return HealthStatus{
		OK:      _health.Health == health.HealthHealthy,
		Health:  string(_health.Health),
		Ready:   _health.Ready,
		Status:  string(_health.Status),
		Message: _health.Message,
	}
}

func k8sGetHealth(fnName string) cel.EnvOption {
	return cel.Function(fnName,
		cel.Overload(fnName+"_overload",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				jsonObj, _ := conv.AnyToMapStringAny(GetHealth(obj.Value()))
				return types.NewDynamicMap(types.DefaultTypeAdapter, jsonObj)
			}),
		),
	)
}

func k8sGetStatus(fnName string) cel.EnvOption {
	return cel.Function(fnName,
		cel.Overload(fnName+"_overload",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				return types.String(GetStatus(obj.Value()))
			}),
		),
	)
}

func k8sIsHealthy(fnName string) cel.EnvOption {
	return cel.Function(fnName,
		cel.Overload(fnName+"_overload",
			[]*cel.Type{cel.AnyType},
			cel.BoolType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				return types.Bool(GetHealth(obj.Value()).OK)
			}),
		),
	)
}

func k8sIsReady(fnName string) cel.EnvOption {
	return cel.Function(fnName,
		cel.Overload(fnName+"_overload",
			[]*cel.Type{cel.AnyType},
			cel.BoolType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				return types.Bool(GetHealth(obj.Value()).Ready)
			}),
		),
	)
}
