package k8s

import (
	"fmt"

	"github.com/flanksource/is-healthy/pkg/health"
	"github.com/flanksource/is-healthy/pkg/lua"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type HealthStatus struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	OK      bool   `json:"ok,omitempty"`
}

func GetUnstructured(in interface{}) *unstructured.Unstructured {
	var err error
	obj := make(map[string]interface{})

	switch v := in.(type) {
	case string:
		err = yaml.Unmarshal([]byte(v), &obj)
	case []byte:
		err = yaml.Unmarshal(v, &obj)
	case map[string]interface{}:
		obj = v
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

func IsHealthy(in interface{}) bool {
	return GetHealth(in).OK
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
			OK:      false,
			Status:  "Missing",
			Message: "No health check found",
		}
	}

	return HealthStatus{
		OK:      _health.Status == health.HealthStatusHealthy || _health.Status == health.HealthStatusProgressing,
		Status:  string(_health.Status),
		Message: _health.Message,
	}
}
