package lang

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/flanksource/gomplate/v4/genmonarch"
)

//go:embed src/generated/spec.json
var specJSON []byte

// Spec reads the same generated catalogue the editor ships, without requiring
// Go source files or a Go installation at runtime.
func Spec() (*genmonarch.Spec, error) {
	var spec genmonarch.Spec
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return nil, fmt.Errorf("decode bundled language catalogue: %w", err)
	}
	if len(spec.CEL.Functions) == 0 || len(spec.GoTemplate.Functions) == 0 {
		return nil, fmt.Errorf("bundled language catalogue has no functions")
	}
	return &spec, nil
}
