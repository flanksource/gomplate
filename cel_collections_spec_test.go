package gomplate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("verification collection functions", func() {
	It("returns grouped native integers without a JSON number round trip", func() {
		value, err := RunExpression(nil, Template{Expression: `[{"amount":9007199254740993}].groupBy(r,"fund").one().rows`, Strict: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal([]any{map[string]any{"amount": int64(9007199254740993)}}))
	})
	It("groups collections supplied by the host environment", func() {
		value, err := RunExpression(map[string]any{"rows": []map[string]any{{"fund": "Employee", "amount": 20}}}, Template{
			Expression: `rows.groupBy(r,r.fund).one().rows.one().amount == 20`, Strict: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(BeTrue())
	})
	It("keeps strict verification separate from nil-safe cached expressions", func() {
		env := map[string]any{"row": map[string]any{}}
		_, err := RunExpression(env, Template{Expression: "row.missing == 0", Strict: true})
		Expect(err).To(HaveOccurred())
		value, err := RunExpression(env, Template{Expression: "row.missing == 0"})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(BeTrue())
		_, err = RunExpression(env, Template{Expression: "row.missing == 0", Strict: true})
		Expect(err).To(HaveOccurred())
	})
	DescribeTable("preserves keys, order and native row values",
		func(expression string) {
			value, err := RunExpression(nil, Template{Expression: expression})
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(BeTrue())
		},
		Entry("grouped rows", `[{'fund':'a','amount':2}, {'fund':'b','amount':4}, {'fund':'a','amount':3}].groupBy(r, r.fund).map(g, {'fund':g.key, 'total':g.rows.map(r,r.amount).sum()}) == [{'fund':'a','total':5}, {'fund':'b','total':4}]`),
		Entry("empty grouping", `[].groupBy(r, r.fund) == []`),
		Entry("typed keys", `[{'key':1}, {'key':'1'}].groupBy(r,r.key).size() == 2`),
		Entry("first distinct", `[{'key':'a','value':1}, {'key':'a','value':2}, {'key':'b','value':3}].distinctBy(r,r.key).map(r,r.value) == [1,3]`),
		Entry("latest row", `[{'date':2,'value':'new'}, {'date':1,'value':'old'}].latestBy(r,r.date).value == 'new'`),
		Entry("one row", `[{'amount':12}].one().amount == 12`),
		Entry("exact decimal sum", `['9007199254740993.01','0.09'].sum() == '9007199254740993.1'`),
		Entry("decimal cancellation", `['1.20','-1.20'].sum() == '0'`),
		Entry("decimal input rounding", `['9007199254740993.0000000001'].sum() == '9007199254740993'`),
		Entry("decimal half even", `['0.12345678901234567885'].sum() == '0.1234567890123456788'`),
		Entry("decimal underflow", `['0.00000000000000000005'].sum() == '0'`),
		Entry("no intermediate rounding", `['9999999999999999999','1','-9999999999999999999'].sum() == '1'`),
		Entry("join", `[{'fund':'a','amount':3}].joinBy([{'id':'a','amount':7}, {'id':'b','amount':4}], l,l.fund,r,r.id).one().right.amount == 7`),
	)
	DescribeTable("rejects ambiguous or invalid evidence",
		func(expression string, message string) {
			_, err := RunExpression(nil, Template{Expression: expression})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(message))
		},
		Entry("missing single row", `[].one()`, "exactly one row"),
		Entry("multiple single rows", `[1,2].one()`, "exactly one row"),
		Entry("empty latest", `[].latestBy(r,r.date)`, "latestBy requires at least one row"),
		Entry("tied latest", `[{'date':1}, {'date':1}].latestBy(r,r.date)`, "latestBy has multiple rows"),
		Entry("null group key", `[{'key':null}].groupBy(r,r.key)`, "groupBy key"),
		Entry("invalid decimal sum", `['1300','KES'].sum()`, "sum requires decimal strings"),
		Entry("decimal overflow", `['9999999999999999999','1'].sum()`, "overflow"),
	)
})
