package genmonarch

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("the go-template vocabulary read from the live FuncMap", func() {
	var spec GoTemplateSpec
	var byName map[string]Function

	ginkgo.BeforeEach(func() {
		var err error
		spec, err = ExtractGoTemplate()
		Expect(err).ToNot(HaveOccurred())
		byName = map[string]Function{}
		for _, f := range spec.Functions {
			byName[f.Name] = f
		}
	})

	ginkgo.It("finds every namespace CreateFuncs installs", func() {
		Expect(spec.Namespaces).To(ConsistOf(
			"base64", "coll", "conv", "crypto", "data", "filepath", "k8s",
			"math", "net", "path", "random", "regexp", "strings", "test",
			"time", "uuid",
		))
	})

	ginkgo.It("keeps namespaces go-template has but CEL does not", func() {
		// conv.* and path.* are registered for go templates while being
		// commented out of funcs/cel_exports.go. The two catalogues are
		// genuinely different and the specs must say so.
		celSpec, err := ExtractCEL()
		Expect(err).ToNot(HaveOccurred())
		Expect(spec.Namespaces).To(ContainElements("conv", "path"))
		Expect(celSpec.Namespaces).ToNot(ContainElement("conv"))
		Expect(celSpec.Namespaces).ToNot(ContainElement("path"))
	})

	ginkgo.DescribeTable("expands namespace accessors into their methods, with signatures",
		func(name, wantSignature string) {
			fn, found := byName[name]
			Expect(found).To(BeTrue(), "%s is not registered", name)
			Expect(fn.Signature).To(Equal(wantSignature))
		},
		// The receiver is dropped: an author writes `strings.ToUpper "x"`.
		ginkgo.Entry("string method", "strings.ToUpper", "(s interface {}) string"),
		ginkgo.Entry("variadic", "coll.Dict", "(in ...interface {}) (map[string]interface {}, error)"),
		ginkgo.Entry("typed return", "time.ParseDuration", "(n interface {}) (time.Duration, error)"),
		ginkgo.Entry("two args", "jq", "(jqExpr string, in interface {}) (interface {}, error)"),
		ginkgo.Entry("grouped parameters", "contains", "(s string, substr string) bool"),
	)

	ginkgo.It("uses the exported method names, which differ in case from CEL's", func() {
		// The same helper is `k8s.IsHealthy` in a go template and
		// `k8s.isHealthy` in CEL. Sharing one list across both would be wrong.
		Expect(byName).To(HaveKey("k8s.IsHealthy"))
		Expect(byName).ToNot(HaveKey("k8s.isHealthy"))
	})

	ginkgo.It("keeps the top-level aliases alongside the namespaced names", func() {
		Expect(byName).To(HaveKey("toJSON"))
		Expect(byName).To(HaveKey("isHealthy"))
		Expect(byName).To(HaveKey("strings.ToUpper"))
	})

	ginkgo.It("is deterministic, so a regenerated spec produces no diff", func() {
		again, err := ExtractGoTemplate()
		Expect(err).ToNot(HaveOccurred())
		Expect(again).To(Equal(spec))
	})
})
