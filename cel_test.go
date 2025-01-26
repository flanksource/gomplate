package gomplate

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"
)

func TestIsValidCELIdentifier(t *testing.T) {
	testCases := []struct {
		identifier string
		valid      bool
	}{
		{"variable", true},
		{"_var123", true},
		{"someVariable", true},

		{"123var", false},
		{"var-name", false},
		{"", false},
		{"var$", false},
		{"if", false},
		{"Σ_variable", false},
	}

	g := gomega.NewWithT(t)

	for i, tc := range testCases {
		result := IsValidCELIdentifier(tc.identifier)
		g.Expect(result).To(gomega.Equal(tc.valid), fmt.Sprintf("%d %s", i+1, tc.identifier))
	}
}
