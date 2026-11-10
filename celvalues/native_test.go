package celvalues_test

import (
	"fmt"
	"io"
	"reflect"

	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"github.com/flanksource/commons/decimal"
	"github.com/flanksource/gomplate/v4/celvalues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type displayString string

var _ = Describe("native scalar conversions", func() {
	It("returns values assignable to the requested native type and rejects unsupported interfaces", func() {
		date, err := celvalues.NewDate("2026-11-10")
		Expect(err).NotTo(HaveOccurred())
		money, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "125", Currency: "USD", DecimalPlaces: 2})
		Expect(err).NotTo(HaveOccurred())
		for _, value := range []ref.Val{date, celvalues.NewAdapter(types.DefaultTypeAdapter).NativeToValue(money)} {
			for _, target := range []reflect.Type{reflect.TypeOf(value.Value()), reflect.TypeFor[any](), reflect.TypeFor[fmt.Stringer](), reflect.TypeFor[string](), reflect.TypeFor[displayString]()} {
				native, err := value.ConvertToNative(target)
				Expect(err).NotTo(HaveOccurred())
				Expect(reflect.TypeOf(native).AssignableTo(target)).To(BeTrue(), "%T to %s", value, target)
			}
			_, err := value.ConvertToNative(reflect.TypeFor[io.Writer]())
			Expect(err).To(HaveOccurred(), "%T to io.Writer", value)
		}
	})
})
