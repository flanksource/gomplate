package gomplate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("cache key boundaries", func() {
	It("keeps delimiter and body boundaries distinct on cache hits", func() {
		for _, label := range []string{"", "explicit-delimiter-key"} {
			for _, template := range []Template{
				{LeftDelim: "$(", RightDelim: ")", Template: "Hi", CacheKey: label},
				{LeftDelim: "$", RightDelim: ")", Template: "(Hi", CacheKey: label},
			} {
				for range 2 {
					value, err := RunTemplate(nil, template)
					Expect(err).NotTo(HaveOccurred())
					Expect(value).To(Equal(template.Template))
				}
			}
		}
	})

	It("keeps explicit Go template cache labels separate", func() {
		first := Template{Template: "hello", CacheKey: "first"}
		second := first
		second.CacheKey = "second"
		Expect(first.autoCacheKey(nil)).NotTo(Equal(second.autoCacheKey(nil)))
	})

	It("keeps environment variable boundaries distinct", func() {
		template := Template{Expression: "1"}
		Expect(template.autoCacheKey(map[string]any{"a-b": 1})).
			NotTo(Equal(template.autoCacheKey(map[string]any{"a": 1, "b": 1})))
	})

	It("keeps strict semantics separate from explicit key prefixes", func() {
		for _, strictFirst := range []bool{false, true} {
			celExpressionCache.Flush()
			for _, strict := range []bool{strictFirst, !strictFirst, strictFirst} {
				key := "strict:boundary"
				if strict {
					key = "boundary"
				}
				value, err := RunExpression(map[string]any{"row": map[string]any{}}, Template{
					Expression: "row.missing == 0", Strict: strict, CacheKey: key,
				})
				if strict {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(value).To(BeTrue())
				}
			}
		}
	})

	It("keeps native generations separate from explicit key prefixes", func() {
		Expect((Template{CacheKey: "1:boundary"}).celCacheKey(nil, 0)).
			NotTo(Equal((Template{CacheKey: "boundary"}).celCacheKey(nil, 1)))
	})
})
