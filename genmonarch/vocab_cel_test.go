package genmonarch

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("the CEL vocabulary read from a live environment", func() {
	var spec CELSpec
	var byName map[string]Function

	ginkgo.BeforeEach(func() {
		var err error
		spec, err = ExtractCEL()
		Expect(err).ToNot(HaveOccurred())
		byName = map[string]Function{}
		for _, f := range spec.Functions {
			byName[f.Name] = f
		}
	})

	ginkgo.It("finds the namespaces gomplate registers, not the ones CEL.md claims", func() {
		// conv.* and path.* are commented out of funcs/cel_exports.go, so they
		// must not appear however much the prose suggests otherwise.
		Expect(spec.Namespaces).To(ContainElements("k8s", "aws", "math", "time", "filepath", "crypto", "random", "regexp", "base64", "uuid", "test", "net"))
		Expect(spec.Namespaces).ToNot(ContainElement("conv"))
		Expect(spec.Namespaces).ToNot(ContainElement("gcp"))
	})

	ginkgo.It("carries the reserved words from the evaluator's own table", func() {
		Expect(spec.Keywords).To(ContainElements("true", "false", "null", "in", "as", "while"))
	})

	ginkgo.DescribeTable("registers the functions this fork adds",
		func(name, namespace string) {
			fn, found := byName[name]
			Expect(found).To(BeTrue(), "%s is not registered", name)
			Expect(fn.Namespace).To(Equal(namespace))
			Expect(fn.Overloads).ToNot(BeEmpty())
		},
		ginkgo.Entry("k8s health", "k8s.isHealthy", "k8s"),
		ginkgo.Entry("k8s quantity", "k8s.cpuAsMillicores", "k8s"),
		ginkgo.Entry("aws", "aws.arnToMap", "aws"),
		ginkgo.Entry("un-namespaced coll", "matchLabel", ""),
		ginkgo.Entry("go-template bridge", "f", ""),
		ginkgo.Entry("business hours", "in_business_hours", ""),
	)

	ginkgo.It("separates member-only functions from globally callable ones", func() {
		// `x.sum()` is valid, a bare `sum(x)` is not -- the tokenizer must only
		// colour the former, so this classification has to be right.
		Expect(byName["sum"].MemberOnly).To(BeTrue())
		Expect(byName["getHost"].MemberOnly).To(BeTrue())
		Expect(byName["k8s.isHealthy"].MemberOnly).To(BeFalse())
		Expect(byName["matchLabel"].MemberOnly).To(BeFalse())
	})

	ginkgo.It("keeps aliases, because both spellings are real", func() {
		for _, alias := range []string{"k8s.isHealthy", "k8s.is_healthy", "IsHealthy"} {
			Expect(byName).To(HaveKey(alias))
		}
	})

	ginkgo.It("excludes operator and internal declarations", func() {
		// The checker declares `_+_`, `_?._` and `__not_strictly_false__` as
		// functions; none of them can be typed by an author.
		for name := range byName {
			Expect(isAuthorable(name)).To(BeTrue(), "%q is not an authorable identifier", name)
		}
		Expect(byName).ToNot(HaveKey("_+_"))
		Expect(byName).ToNot(HaveKey("__not_strictly_false__"))
	})

	ginkgo.It("captures macros with their arities, including gomplate's own fold", func() {
		type key struct {
			name  string
			arity int
		}
		got := map[key]Macro{}
		for _, m := range spec.Macros {
			got[key{m.Name, m.ArgCount}] = m
		}
		Expect(got).To(HaveKey(key{"has", 1}))
		Expect(got).To(HaveKey(key{"all", 2}))
		Expect(got).To(HaveKey(key{"exists_one", 2}))
		Expect(got).To(HaveKey(key{"fold", 3}))
		Expect(got).To(HaveKey(key{"fold", 4}))

		// cel_fold.go attaches MacroDocs/MacroExamples; they should survive.
		Expect(got[key{"fold", 3}].Doc).To(ContainSubstring("accumulator"))
		Expect(got[key{"fold", 3}].Examples).ToNot(BeEmpty())
	})

	ginkgo.It("is deterministic, so a regenerated spec produces no diff", func() {
		again, err := ExtractCEL()
		Expect(err).ToNot(HaveOccurred())
		Expect(again).To(Equal(spec))
	})
})
