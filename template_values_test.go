package gomplate

import (
	"reflect"
	"strings"
	stdtemplate "text/template"
	"time"

	"github.com/flanksource/commons/decimal"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type tvMeta struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels,omitempty"`
	Count  uint              `json:"count"`
	Note   string            `json:"note,omitempty"`
}

type tvBase struct {
	ID     uuid.UUID  `json:"id"`
	Parent *uuid.UUID `json:"parent"`
}

type tvItem struct {
	tvBase
	Meta    tvMeta        `json:"metadata"`
	Tags    []string      `json:"tags"`
	Raw     []byte        `json:"raw"`
	Created time.Time     `json:"created"`
	Updated *time.Time    `json:"updated"`
	Age     time.Duration `json:"age"`
	Untag   string
	Skipped string `json:"-"`
	hidden  string
}

type tvMapper struct{ calls *int }

func (m tvMapper) AsMap(...string) map[string]any {
	*m.calls++
	return map[string]any{"kind": "mapped", "nested": map[string]any{"deep": 1}}
}

// upperProvider resolves like the default provider and upper-cases strings it hands to consumers.
type upperProvider struct{ DefaultValuesProvider }

func (p upperProvider) Value(v reflect.Value) (reflect.Value, error) {
	if v.Kind() == reflect.String {
		return reflect.ValueOf(strings.ToUpper(v.String())), nil
	}
	return p.DefaultValuesProvider.Value(v)
}

func tvEnvironment() map[string]any {
	price, err := decimal.NewMoney(decimal.MoneyOptions{Amount: "12.50", Currency: "USD"})
	Expect(err).ToNot(HaveOccurred())
	calls := 0
	updated := time.Date(2026, 11, 13, 8, 0, 0, 0, time.UTC)
	item := tvItem{
		tvBase:  tvBase{ID: uuid.MustParse("6f1c0d6a-1b2c-4c3d-8e9f-0a1b2c3d4e5f")},
		Meta:    tvMeta{Name: "web", Labels: map[string]string{"app": "web", "tier": "front"}, Count: 3},
		Tags:    []string{"a", "b"},
		Raw:     []byte("bytes"),
		Created: time.Date(2026, 11, 12, 10, 30, 0, 0, time.UTC),
		Updated: &updated,
		Age:     90 * time.Second,
		Untag:   "untagged",
		Skipped: "skipped",
		hidden:  "hidden",
	}
	return map[string]any{
		"item":   item,
		"ptr":    &item,
		"items":  []tvItem{item, {Meta: tvMeta{Name: "db"}}},
		"mapper": tvMapper{calls: &calls},
		"labels": map[string]string{"env": "prod"},
		"price":  price,
		"count":  2,
		"name":   "plain",
		"id":     item.ID,
		"age":    item.Age,
	}
}

// serializedRender is the go template path before the values provider: serialize, then text/template.
func serializedRender(text string, env map[string]any) (string, error) {
	data, err := Serialize(env)
	if err != nil {
		return "", err
	}
	tpl, err := stdtemplate.New("").Funcs(stdtemplate.FuncMap(funcMap)).Parse(text)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	if err := tpl.Execute(&out, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

var _ = Describe("Go template values provider", func() {
	DescribeTable("renders the same as serializing the environment first",
		func(text string) {
			expected, err := serializedRender(text, tvEnvironment())
			Expect(err).ToNot(HaveOccurred())

			out, err := RunTemplate(tvEnvironment(), Template{Template: text})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal(expected))
		},
		Entry("json tagged field", "{{ .item.metadata.name }}"),
		Entry("through a pointer", "{{ .ptr.metadata.name }}"),
		Entry("untagged field by Go name", "{{ .item.Untag }}"),
		Entry("unsigned count arithmetic", "{{ add .item.metadata.count 1 }}"),
		Entry("bytes", "{{ .item.raw }}"),
		Entry("time method", `{{ .item.created.Format "2006-01-02" }}`),
		Entry("time pointer", "{{ .item.updated }}"),
		Entry("top level uuid", "{{ .id }}"),
		Entry("top level duration", "{{ .age }}"),
		Entry("money", "{{ .price }}"),
		Entry("whole struct printed", "{{ .item.metadata }}"),
		Entry("struct to toJSON", "{{ .item.metadata | toJSON }}"),
		Entry("map[string]string keys", "{{ keys .item.metadata.labels | sort }}"),
		Entry("has on a struct map", `{{ has .item.metadata "name" }}`),
		Entry("index into a struct", `{{ index .item.metadata.labels "app" }}`),
		Entry("range over struct slice", "{{ range .items }}{{ .metadata.name }},{{ end }}"),
		Entry("range over struct map", "{{ range $k, $v := .item.metadata.labels }}{{ $k }}={{ $v }};{{ end }}"),
		Entry("join a string slice", `{{ join .item.tags "," }}`),
		Entry("as mapper", "{{ .mapper.kind }}-{{ .mapper.nested.deep }}"),
		Entry("if on an empty omitempty field", "{{ if .item.metadata.note }}yes{{ else }}no{{ end }}"),
		Entry("with on struct", "{{ with .item.metadata }}{{ .name }}{{ end }}"),
		Entry("plain values", "{{ .name }} {{ .count }}"),
		Entry("variable then field", "{{ $m := .item.metadata }}{{ $m.name }}"),
		Entry("whole environment to json", "{{ .labels | toJSON }}"),
	)

	// Serialize only preserved uuids and durations at the top level of the environment (its FIXME):
	// nested in a struct a uuid became a list of bytes and a duration nanoseconds. Read lazily they
	// render the same at any depth, as CEL already reads them.
	DescribeTable("renders nested uuids and durations like top level ones",
		func(text, expected string) {
			out, err := RunTemplate(tvEnvironment(), Template{Template: text})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal(expected))
		},
		Entry("promoted embedded uuid", "{{ .item.id }}", "6f1c0d6a-1b2c-4c3d-8e9f-0a1b2c3d4e5f"),
		Entry("nested uuid compared to a string", `{{ eq .item.id "6f1c0d6a-1b2c-4c3d-8e9f-0a1b2c3d4e5f" }}`, "true"),
		Entry("nested duration", "{{ .item.age }}", "1m30s"),
	)

	DescribeTable("renders a missing value as empty text, and <no value> when strict",
		func(text string) {
			out, err := RunTemplate(tvEnvironment(), Template{Template: "[" + text + "]"})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("[]"))

			out, err = RunTemplate(tvEnvironment(), Template{Template: "[" + text + "]", Strict: true})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("[<no value>]"))
		},
		Entry("missing top level key", "{{ .nope }}"),
		Entry("missing struct key", "{{ .item.metadata.nope }}"),
		Entry("missing map[string]string key", "{{ .labels.nope }}"),
		Entry("index a missing map[string]string key", `{{ index .labels "nope" }}`),
		Entry("as mapper missing key", "{{ .mapper.nope }}"),
		Entry("nil uuid pointer", "{{ .item.parent }}"),
	)

	It("does not materialise values the template never reads", func() {
		calls := 0
		env := map[string]any{"item": tvEnvironment()["item"], "mapper": tvMapper{calls: &calls}}
		out, err := RunTemplate(env, Template{Template: "{{ .item.metadata.name }}"})
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(Equal("web"))
		Expect(calls).To(Equal(0))
	})

	It("uses a custom values provider", func() {
		out, err := RunTemplate(map[string]any{"item": tvEnvironment()["item"]}, Template{
			Template:       "{{ .item.metadata.name }}",
			ValuesProvider: upperProvider{},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(Equal("WEB"))
	})

	DescribeTable("is nil-safe unless strict",
		func(text, strictError string) {
			env := map[string]any{"list": []any{1}, "null": nil, "item": tvEnvironment()["item"]}
			out, err := RunTemplate(env, Template{Template: text})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(BeEmpty())

			_, err = RunTemplate(env, Template{Template: text, Strict: true})
			Expect(err).To(MatchError(ContainSubstring(strictError)))
		},
		Entry("field through nil", "{{ .null.a.b }}", "nil pointer evaluating interface {}.a"),
		Entry("index past the end", "{{ index .list 3 }}", "index out of range: 3"),
		Entry("index through nil", `{{ index .null "a" }}`, "index of untyped nil"),
		Entry("nil uuid pointer field", "{{ .item.parent.x }}", "nil pointer evaluating *uuid.UUID.x"),
	)
})
