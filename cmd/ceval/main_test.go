package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvironmentYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env.yaml")
	require.NoError(t, os.WriteFile(path, []byte("labels:\n  a/b/c: gg\n"), 0o644))

	env, err := loadEnvironment(path)
	require.NoError(t, err)

	labels, ok := env["labels"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "gg", labels["a/b/c"])
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "env.yaml")
	require.NoError(t, os.WriteFile(path, []byte("labels:\n  a/b/c: gg\n"), 0o644))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"-f", path, "-e", "labels['a/b/c']"}, &stdout, &stderr)

	assert.Equal(t, 0, code)
	assert.Equal(t, "gg\n", stdout.String())
	assert.Equal(t, "", stderr.String())
}
