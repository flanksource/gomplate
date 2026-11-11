package celext_test

import (
	"runtime/debug"

	"github.com/flanksource/gomplate/v4/celext"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SourceVersion", func() {
	It("matches the cel-go dependency in the test binary", func() {
		info, ok := debug.ReadBuildInfo()
		Expect(ok).To(BeTrue(), "cannot read test binary build info; run make celext")
		for _, dependency := range info.Deps {
			if dependency.Path == "cel.dev/cel-go" {
				Expect(celext.SourceVersion).To(Equal(dependency.Version), "cel-go source drift: run make celext")
				return
			}
		}
		Fail("cel.dev/cel-go is missing from test binary build info; run make celext")
	})
})
