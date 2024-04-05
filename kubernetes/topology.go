package kubernetes

import (
	"strings"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func celk8sLabels() cel.EnvOption {
	return cel.Function("k8s.labels",
		cel.Overload("k8s.labels_map_map",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				val := k8sLabels(obj.Value())
				return types.NewStringStringMap(types.DefaultTypeAdapter, val)
			}),
		),
	)
}

func k8sLabels(input any) map[string]string {
	labels := make(map[string]string)

	obj := GetUnstructured(input)
	if obj == nil {
		return labels
	}

	if ns := obj.GetNamespace(); ns != "" {
		labels["namespace"] = ns
	}

	for k, v := range obj.GetLabels() {
		if strings.HasSuffix(k, "-hash") {
			continue
		}
		labels[k] = v
	}

	return labels
}

func celPodProperties() cel.EnvOption {
	return cel.Function("k8s.podProperties",
		cel.Overload("k8s.podProperties_list_dyn_map",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				jsonObj, _ := conv.AnyToListMapStringAny(PodComponentProperties(obj.Value()))
				return types.NewDynamicList(types.DefaultTypeAdapter, jsonObj)
			}),
		),
	)
}

func celPodMaxMemoryBytes() cel.EnvOption {
	return cel.Function("k8s.podMaxMemoryBytes",
		cel.Overload("k8s.podMaxMemoryBytes_obj_int",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				val := conv.ToInt(podMaxMemory(obj.Value()))
				return types.Int(val)
			}),
		),
	)
}

func podMaxMemory(input any) int64 {
	obj := GetUnstructured(input)
	if obj == nil {
		return 0
	}

	var pod corev1.Pod
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &pod)
	if err != nil {
		return 0
	}
	var totalMemBytes int64
	for _, container := range pod.Spec.Containers {
		mem := container.Resources.Limits.Memory()
		if mem != nil {
			totalMemBytes += _k8sMemoryAsBytes(mem.String())
		}
	}
	return totalMemBytes
}

func celPodMaxCPUMillicores() cel.EnvOption {
	return cel.Function("k8s.podMaxCPUMillicores",
		cel.Overload("k8s.podMaxCPUMillicores_obj_int",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				val := conv.ToInt(podMaxCPU(obj.Value()))
				return types.Int(val)
			}),
		),
	)
}

func podMaxCPU(input any) int64 {
	obj := GetUnstructured(input)
	if obj == nil {
		return 0
	}

	var pod corev1.Pod
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &pod)
	if err != nil {
		return 0
	}
	var totalCPU int64
	for _, container := range pod.Spec.Containers {
		cpu := container.Resources.Limits.Cpu()
		if cpu != nil {
			totalCPU += _k8sCPUAsMillicores(cpu.String())
		}
	}
	return totalCPU
}

func PodComponentProperties(input any) []map[string]any {
	obj := GetUnstructured(input)
	if obj == nil {
		return nil
	}

	var pod corev1.Pod
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &pod)
	if err != nil {
		return nil
	}

	var totalCPU int64
	for _, container := range pod.Spec.Containers {
		cpu := container.Resources.Limits.Cpu()
		if cpu != nil {
			totalCPU += _k8sCPUAsMillicores(cpu.String())
		}
	}

	var totalMemBytes int64
	for _, container := range pod.Spec.Containers {
		mem := container.Resources.Limits.Memory()
		if mem != nil {
			totalMemBytes += _k8sMemoryAsBytes(mem.String())
		}
	}

	rootContainer := pod.Spec.Containers[0]
	return []map[string]any{
		{"name": "image", "text": rootContainer.Image},
		{"name": "cpu", "max": totalCPU, "unit": "millicores", "headline": true},
		{"name": "memory", "max": totalMemBytes, "unit": "bytes", "headline": true},
		{"name": "node", "text": pod.Spec.NodeName},
		{"name": "created_at", "text": pod.ObjectMeta.CreationTimestamp.String()},
		{"name": "namespace", "text": pod.ObjectMeta.Namespace},
	}
}

func celNodeProperties() cel.EnvOption {
	return cel.Function("k8s.nodeProperties",
		cel.Overload("k8s.nodeProperties_list_dyn_map",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				jsonObj, _ := conv.AnyToListMapStringAny(NodeComponentProperties(obj.Value()))
				return types.NewDynamicList(types.DefaultTypeAdapter, jsonObj)
			}),
		),
	)
}

func NodeComponentProperties(input any) []map[string]any {
	obj := GetUnstructured(input)
	if obj == nil {
		return nil
	}

	var node corev1.Node
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &node)
	if err != nil {
		return nil
	}

	totalCPU := _k8sCPUAsMillicores(node.Status.Allocatable.Cpu().String())
	totalMemBytes := _k8sMemoryAsBytes(node.Status.Allocatable.Memory().String())
	totalStorage := _k8sMemoryAsBytes(node.Status.Allocatable.StorageEphemeral().String())

	return []map[string]any{
		{"name": "cpu", "max": totalCPU, "unit": "millicores", "headline": true},
		{"name": "memory", "max": totalMemBytes, "unit": "bytes", "headline": true},
		{"name": "ephemeral-storage", "max": totalStorage, "unit": "bytes", "headline": true},
		{"name": "zone", "text": node.GetLabels()["topology.kubernetes.io/zone"]},
	}
}
