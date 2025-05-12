package gcp

import "github.com/google/cel-go/cel"

func Library() []cel.EnvOption {
	return []cel.EnvOption{
		gcpIncidentToCheckResult("gcp.incidents.toCheckResult"),
	}
}
