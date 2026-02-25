package coll

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirst(t *testing.T) {
	tests := []struct {
		name     string
		in       any
		expected any
	}{
		{"list first element", []any{1, 2, 3}, 1},
		{"single-element list", []any{"x"}, "x"},
		{"empty list", []any{}, nil},
		{"nil input", nil, nil},
		{"string first char", "hello", "h"},
		{"empty string", "", ""},
		{"map first by sorted key", map[string]any{"b": 2, "a": 1}, 1},
		{"empty map", map[string]any{}, nil},
		{"typed string slice", []string{"a", "b"}, "a"},
		{"typed int slice", []int{10, 20, 30}, 10},
		{"multibyte string", "héllo", "h"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, First(tc.in))
		})
	}
}

func TestLast(t *testing.T) {
	tests := []struct {
		name     string
		in       any
		expected any
	}{
		{"list last element", []any{1, 2, 3}, 3},
		{"single-element list", []any{"x"}, "x"},
		{"empty list", []any{}, nil},
		{"nil input", nil, nil},
		{"string last char", "hello", "o"},
		{"empty string", "", ""},
		{"map last by sorted key", map[string]any{"b": 2, "a": 1}, 2},
		{"empty map", map[string]any{}, nil},
		{"typed string slice", []string{"a", "b"}, "b"},
		{"typed int slice", []int{10, 20, 30}, 30},
		{"multibyte string", "héllo", "o"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Last(tc.in))
		})
	}
}
