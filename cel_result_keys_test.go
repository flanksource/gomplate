package gomplate

import (
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type resultKeyMap struct {
	traits.Mapper
	key ref.Val
}

func (m *resultKeyMap) Value() any { return m }
func (m *resultKeyMap) Iterator() traits.Iterator {
	return types.NewRefValList(types.DefaultTypeAdapter, []ref.Val{m.key}).Iterator()
}
func (m *resultKeyMap) Get(ref.Val) ref.Val { return types.String("value") }

var _ = Describe("CEL map result keys", func() {
	DescribeTable("preserves maps whose native keys cannot be hashed", func(key ref.Val) {
		value := &resultKeyMap{Mapper: types.NewStringInterfaceMap(types.DefaultTypeAdapter, nil), key: key}
		Expect(func() { Expect(CELValue(value)).To(BeIdenticalTo(value)) }).NotTo(Panic())
	},
		Entry("bytes", types.Bytes("key")),
		Entry("list", types.NewRefValList(types.DefaultTypeAdapter, []ref.Val{types.String("key")})),
		Entry("map", types.NewRefValMap(types.DefaultTypeAdapter, map[ref.Val]ref.Val{types.String("key"): types.True})),
	)
	It("converts comparable non-string keys", func() {
		value := types.NewRefValMap(types.DefaultTypeAdapter, map[ref.Val]ref.Val{types.Int(3): types.String("value")})
		Expect(CELValue(value)).To(Equal(map[any]any{int64(3): "value"}))
	})
})
