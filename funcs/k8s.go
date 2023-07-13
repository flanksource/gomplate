package funcs

import (
	"context"

	"github.com/flanksource/gomplate/v3/k8s"
)

// CreateFilePathFuncs -
func CreateKubernetesFuncs(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"isHealthy": k8s.IsHealthy,
		"getStatus": k8s.GetStatus,
		"getHealth": k8s.GetHealth,
	}
}
