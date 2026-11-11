package gomplate

import (
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type celInputAddress struct {
	City    string `json:"city_name"`
	Country string `json:"country,omitempty"`
}

type celInputAudit struct {
	CreatedBy string `json:"created_by"`
}

type celInputFile struct {
	celInputAudit
	Name     string           `json:"name"`
	Size     uint32           `json:"size"`
	Timeout  time.Duration    `json:"timeout"`
	Modified *time.Time       `json:"modified"`
	Owner    *celInputAddress `json:"owner"`
	Missing  *celInputAddress `json:"missing"`
	Secret   string           `json:"-"`
	Mirrors  []celInputAddress
	internal string
}

type celInputMapper struct{ name string }

func (m celInputMapper) AsMap(...string) map[string]any { return map[string]any{"name": m.name} }

var _ = Describe("CEL evaluation of Go input values", func() {
	modified := time.Date(2026, time.November, 4, 11, 8, 19, 0, time.UTC)
	file := &celInputFile{
		celInputAudit: celInputAudit{CreatedBy: "ops"},
		Name:          "report.csv",
		Size:          10,
		Timeout:       90 * time.Second,
		Modified:      &modified,
		Owner:         &celInputAddress{City: "Kathmandu"},
		Secret:        "hidden",
		Mirrors:       []celInputAddress{{City: "Lalitpur"}, {City: "Bhaktapur"}},
		internal:      "unexported",
	}
	id := uuid.MustParse("962f999c-a9bd-40a4-80bf-47c84b1ad750")
	environment := map[string]any{
		"file":   file,
		"ids":    map[string]any{"container": id},
		"mapper": map[string]any{"item": celInputMapper{name: "mapped"}},
	}

	DescribeTable("reads struct fields the way encoding/json names them",
		func(expression string, expected any) {
			value, err := RunExpression(environment, Template{Expression: expression})
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(expected))
		},
		Entry("json tag names", `file.name`, "report.csv"),
		Entry("untagged fields keep their Go name", `file.Mirrors.map(m, m.city_name)`, []any{"Lalitpur", "Bhaktapur"}),
		Entry("embedded struct fields are promoted", `file.created_by`, "ops"),
		Entry("pointer fields are followed", `file.owner.city_name`, "Kathmandu"),
		Entry("nil pointer fields are null", `file.missing == null`, true),
		Entry("omitempty zero fields are absent", `has(file.owner.country)`, false),
		Entry("json:\"-\" fields are absent", `has(file.Secret) || has(file.secret)`, false),
		Entry("unexported fields are absent", `has(file.internal)`, false),
		Entry("unsigned fields behave as CEL ints", `file.size + 1`, int64(11)),
		Entry("durations inside structs stay durations", `file.timeout > duration("1m")`, true),
		Entry("time pointers inside structs are timestamps", `file.modified < timestamp("2030-01-01T00:00:00Z")`, true),
		Entry("UUIDs are strings", `ids.container.startsWith("962f999c")`, true),
		Entry("AsMapper values use their map", `mapper.item.name`, "mapped"),
	)

	It("returns struct-derived results as plain Go maps and lists", func() {
		value, err := RunExpression(environment, Template{Expression: `file.Mirrors`})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal([]any{map[string]any{"city_name": "Lalitpur"}, map[string]any{"city_name": "Bhaktapur"}}))

		value, err = RunExpression(environment, Template{Expression: `{"owner": file.owner}`})
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(map[string]any{"owner": map[string]any{"city_name": "Kathmandu"}}))
	})

	It("leaves the caller's values untouched", func() {
		_, err := RunExpression(environment, Template{Expression: `file.owner.city_name == "Kathmandu"`})
		Expect(err).NotTo(HaveOccurred())
		Expect(environment["file"]).To(BeIdenticalTo(file))
		Expect(file.Owner).To(Equal(&celInputAddress{City: "Kathmandu"}))
	})
})
