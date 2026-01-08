package tests

import (
	"testing"

	"github.com/flanksource/commons/properties"
	"github.com/stretchr/testify/require"
)

func loadTestProperties(t *testing.T) {
	t.Helper()

	err := properties.LoadFile("testdata/test.properties")
	require.NoError(t, err)
}
