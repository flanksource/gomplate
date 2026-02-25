package coll

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoalesce(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		expected any
	}{
		{"first non-empty string", []any{"", "b"}, "b"},
		{"nil skip", []any{nil, "x"}, "x"},
		{"all empty returns nil", []any{"", nil}, nil},
		{"all nil returns nil", []any{nil, nil}, nil},
		{"zero is valid", []any{nil, 0}, 0},
		{"false is valid", []any{nil, false}, false},
		{"empty slice skip", []any{[]any{}, "ok"}, "ok"},
		{"empty map skip", []any{map[string]any{}, "ok"}, "ok"},
		{"single non-empty", []any{"a"}, "a"},
		{"single nil", []any{nil}, nil},
		{"first wins", []any{"first", "second"}, "first"},
		{"non-empty slice is valid", []any{nil, []any{"x"}}, []any{"x"}},
		{"non-empty map is valid", []any{nil, map[string]any{"k": "v"}}, map[string]any{"k": "v"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Coalesce(tc.args...))
		})
	}
}
