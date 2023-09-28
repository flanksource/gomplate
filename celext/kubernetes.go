package celext

import (
	"strings"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/gomplate/v3/k8s"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

func k8sHealth() cel.EnvOption {
	return cel.Function("k8s.health",
		cel.Overload("k8s.health_any",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				jsonObj, _ := anyToMapStringAny(k8s.GetHealth(obj.Value()))
				return types.NewDynamicMap(types.DefaultTypeAdapter, jsonObj)
			}),
		),
	)
}

func k8sIsHealthy() cel.EnvOption {
	return cel.Function("k8s.is_healthy",
		cel.Overload("k8s.is_healthy_any",
			[]*cel.Type{cel.AnyType},
			cel.StringType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				return types.Bool(k8s.GetHealth(obj.Value()).OK)
			}),
		),
	)
}

func k8sCpuAsMillicores() cel.EnvOption {
	return cel.Function("k8s.cpuAsMillicores",
		cel.Overload("k8s.cpuAsMillicores_string",
			[]*cel.Type{cel.StringType},
			cel.IntType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				objVal := conv.ToString(obj.Value())
				var cpu int64
				if strings.HasSuffix(objVal, "m") {
					cpu = conv.ToInt64(strings.ReplaceAll(objVal, "m", ""))
				} else {
					cpu = int64(conv.ToFloat64(objVal) * 1000)
				}
				return types.Int(cpu)
			}),
		),
	)
}

func k8sMemoryAsBytes() cel.EnvOption {
	return cel.Function("k8s.memoryAsBytes",
		cel.Overload("k8s.memoryAsBytes_string",
			[]*cel.Type{cel.StringType},
			cel.IntType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				objVal := conv.ToString(obj.Value())
				var memory int64
				if strings.HasSuffix(objVal, "Gi") {
					memory = int64(conv.ToFloat64(strings.ReplaceAll(objVal, "Gi", "")) * 1024 * 1024 * 1024)
				} else if strings.HasSuffix(objVal, "Mi") {
					memory = int64(conv.ToFloat64(strings.ReplaceAll(objVal, "Mi", "")) * 1024 * 1024)
				} else if strings.HasSuffix(objVal, "Ki") {
					memory = int64(conv.ToFloat64(strings.ReplaceAll(objVal, "Ki", "")) * 1024)
				} else {
					memory = conv.ToInt64(objVal)
				}
				return types.Int(memory)
			}),
		),
	)
}
