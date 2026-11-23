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

var _ = Describe("order statistics", func() {
	It("summarizes a window with nearest-rank percentiles", func() {
		value, err := RunExpression(nil, Template{Expression: `[10,9,8,7,6,5,4,3,2,1].stats()`, Strict: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(map[string]any{
			"min": 1.0, "avg": 5.5, "max": 10.0, "p50": 5.0, "p90": 9.0, "p95": 10.0, "p99": 10.0, "samples": int64(10),
		}))
	})
	It("summarizes a single sample", func() {
		value, err := RunExpression(nil, Template{Expression: `[-2.5].stats()`, Strict: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(map[string]any{
			"min": -2.5, "avg": -2.5, "max": -2.5, "p50": -2.5, "p90": -2.5, "p95": -2.5, "p99": -2.5, "samples": int64(1),
		}))
	})
	DescribeTable("percentile returns the nearest-rank sample as a double",
		func(expression string, expected float64) {
			value, err := RunExpression(nil, Template{Expression: expression, Strict: true})
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(expected))
		},
		Entry("p90 of ten", `[1,2,3,4,5,6,7,8,9,10].percentile(90)`, 9.0),
		Entry("p95 of ten rounds up", `[1,2,3,4,5,6,7,8,9,10].percentile(95)`, 10.0),
		Entry("double q", `[1,2,3,4,5,6,7,8,9,10].percentile(50.0)`, 5.0),
		Entry("single sample", `[7].percentile(1)`, 7.0),
		Entry("q of 100 is the max", `[3,1,2].percentile(100)`, 3.0),
		Entry("q just above 0 is the min", `[3,1,2].percentile(0.001)`, 1.0),
		Entry("mixed int and double", `[3, 2.5, 1].percentile(50)`, 2.5),
		Entry("negative samples", `[2, -5].percentile(50)`, -5.0),
	)
	DescribeTable("rejects samples or ranks that have no percentile",
		func(expression string, message string) {
			_, err := RunExpression(nil, Template{Expression: expression})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(message))
		},
		Entry("empty percentile", `[].percentile(50)`, "percentile requires at least one sample"),
		Entry("empty stats", `[].stats()`, "stats requires at least one sample"),
		Entry("string sample", `[1,'2'].percentile(50)`, "percentile requires numeric samples"),
		Entry("bool sample", `[true].stats()`, "stats requires numeric samples"),
		Entry("NaN sample", `[double('NaN')].percentile(50)`, "percentile requires finite samples"),
		Entry("infinite sample", `[1, double('Inf')].stats()`, "stats requires finite samples"),
		Entry("negative infinite sample", `[double('-Inf')].percentile(50)`, "percentile requires finite samples"),
		Entry("q of 0", `[1].percentile(0)`, "percentile requires q in (0, 100]"),
		Entry("negative q", `[1].percentile(-1)`, "percentile requires q in (0, 100]"),
		Entry("q above 100", `[1].percentile(100.1)`, "percentile requires q in (0, 100]"),
		Entry("NaN q", `[1].percentile(double('NaN'))`, "percentile requires q in (0, 100]"),
		Entry("string q", `[1].percentile('50')`, "percentile requires a numeric q"),
	)
})
