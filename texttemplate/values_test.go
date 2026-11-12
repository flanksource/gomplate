package texttemplate_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gomplate/v4/texttemplate"
)

func TestValues(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "texttemplate values")
}

// box is a host value only the provider knows how to read and materialise.
type box struct{ items map[string]any }

type person struct{ Name string }

type holder struct{ P *person }

func (person) Greeting() string { return "method" }
func (person) Other() string    { return "other-method" }

// recordingProvider resolves box keys and person.Greeting, and materialises a box into its map.
type recordingProvider struct{ boxValues int }

func (p *recordingProvider) Field(receiver reflect.Value, name string) (reflect.Value, bool, error) {
	switch v := receiver.Interface().(type) {
	case box:
		item, ok := v.items[name]
		if !ok {
			return reflect.Value{}, false, nil
		}
		return reflect.ValueOf(item), true, nil
	case person:
		if name == "Greeting" {
			return reflect.ValueOf("provider"), true, nil
		}
	}
	return reflect.Value{}, false, nil
}

func (p *recordingProvider) Value(v reflect.Value) (reflect.Value, error) {
	if !v.IsValid() || !v.CanInterface() {
		return v, nil
	}
	if b, ok := v.Interface().(box); ok {
		p.boxValues++
		return reflect.ValueOf(b.items), nil
	}
	return v, nil
}

var typeName = texttemplate.FuncMap{"typeName": func(v any) string { return fmt.Sprintf("%T", v) }}

func render(text string, data any, opts texttemplate.ExecuteOptions, options ...string) (string, error) {
	tpl := texttemplate.Must(texttemplate.New("t").Funcs(typeName).Option(options...).Parse(text))
	var out strings.Builder
	err := tpl.ExecuteWith(&out, data, opts)
	return out.String(), err
}

var _ = Describe("ValuesProvider", func() {
	full := box{items: map[string]any{"k": "v"}}
	data := map[string]any{"b": full, "empty": box{}, "p": person{Name: "n"}}

	DescribeTable("resolves fields through the provider before methods and struct fields",
		func(text, expected string) {
			out, err := render(text, data, texttemplate.ExecuteOptions{Values: &recordingProvider{}})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal(expected))
		},
		Entry("provider shadows a method", "{{ .p.Greeting }}", "provider"),
		Entry("found=false falls through to the method", "{{ .p.Other }}", "other-method"),
		Entry("found=false falls through to the struct field", "{{ .p.Name }}", "n"),
		Entry("provider-only receiver", "{{ .b.k }}", "v"),
		Entry("unresolved key on a provider-only receiver is missing", "{{ .b.nope }}", "<no value>"),
		Entry("unresolved key on a struct is missing", "{{ .p.nope }}", "<no value>"),
		Entry("index by string goes through the provider", `{{ index .b "k" }}`, "v"),
	)

	It("reports an unresolved struct key with missingkey=error", func() {
		_, err := render("{{ .p.nope }}", data, texttemplate.ExecuteOptions{Values: &recordingProvider{}}, "missingkey=error")
		Expect(err).To(MatchError(ContainSubstring(`no entry for key "nope"`)))
	})

	It("keeps text/template's error for an unknown struct field without a provider", func() {
		_, err := render("{{ .p.nope }}", data, texttemplate.ExecuteOptions{})
		Expect(err).To(MatchError(ContainSubstring("can't evaluate field nope")))
	})

	DescribeTable("materialises values at each consumer",
		func(text, expected string, materialised int) {
			provider := &recordingProvider{}
			out, err := render(text, data, texttemplate.ExecuteOptions{Values: provider})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal(expected))
			Expect(provider.boxValues).To(Equal(materialised))
		},
		Entry("print", "{{ .b }}", "map[k:v]", 1),
		Entry("function argument", "{{ typeName .b }}", "map[string]interface {}", 1),
		Entry("pipeline value", "{{ .b | typeName }}", "map[string]interface {}", 1),
		Entry("range", "{{ range $k, $v := .b }}{{ $k }}={{ $v }}{{ end }}", "k=v", 1),
		Entry("if on an empty box", "{{ if .empty }}yes{{ else }}no{{ end }}", "no", 1),
		Entry("with binds the materialised value", "{{ with .b }}{{ .k }}{{ end }}", "v", 1),
		Entry("not on a field chain intermediate", "{{ .b.k }}", "v", 0),
		Entry("not on a variable assignment", "{{ $x := .b }}{{ $x.k }}", "v", 0),
	)
})

var _ = Describe("NilSafe", func() {
	data := map[string]any{
		"list":   []any{1},
		"holder": holder{},
		"m":      map[string]any{},
		"null":   nil,
		"p":      (*person)(nil),
		"words":  map[string]any{"a": "x"},
	}

	It("keeps text/template's missing-key result for a chain through an absent key", func() {
		out, err := render("{{ .a.b.c }}", data, texttemplate.ExecuteOptions{})
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(Equal("<no value>"))
	})

	DescribeTable("nil and missing receivers",
		func(text, strictError string) {
			out, err := render(text, data, texttemplate.ExecuteOptions{NilSafe: true})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(BeEmpty())

			_, err = render(text, data, texttemplate.ExecuteOptions{})
			Expect(err).To(MatchError(ContainSubstring(strictError)))
		},
		Entry("field chain through a nil value", "{{ .null.b.c }}", "nil pointer evaluating interface {}.b"),
		Entry("field through a nil pointer", "{{ .p.Name }}", "nil pointer evaluating interface {}.Name"),
		Entry("field through a nil typed pointer", "{{ .holder.P.Name }}", "nil pointer evaluating *texttemplate_test.person.Name"),
		Entry("index of a missing value", `{{ index .missing "x" }}`, "index of untyped nil"),
		Entry("index past the end of a slice", "{{ index .list 5 }}", "index out of range: 5"),
		Entry("index through a missing nested key", `{{ index .m "a" "b" }}`, "index of nil pointer"),
		Entry("index through a nil pointer", `{{ index .p "Name" }}`, "index of nil pointer"),
	)

	DescribeTable("prints a missing value as empty text",
		func(text, expected string) {
			out, err := render(text, data, texttemplate.ExecuteOptions{NilSafe: true})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal(expected))
		},
		Entry("missing key between text", "a{{ .nope }}b", "ab"),
		Entry("chain through an absent key", "{{ .a.b.c }}", ""),
		Entry("missing key in a nested map", "{{ .words.nope }}", ""),
	)

	It("still resolves present values", func() {
		out, err := render(`{{ index .words "a" }}-{{ index .list 0 }}`, data, texttemplate.ExecuteOptions{NilSafe: true})
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(Equal("x-1"))
	})

	It("keeps type errors", func() {
		_, err := render(`{{ index .list "x" }}`, data, texttemplate.ExecuteOptions{NilSafe: true})
		Expect(err).To(MatchError(ContainSubstring("cannot index slice/array with type string")))
	})

	It("lets missingkey=error win over nil-safety", func() {
		_, err := render("{{ .a.b }}", data, texttemplate.ExecuteOptions{NilSafe: true}, "missingkey=error")
		Expect(err).To(MatchError(ContainSubstring(`map has no entry for key "a"`)))
	})

	It("is inherited by nested template invocations", func() {
		out, err := render(`{{ define "inner" }}{{ .null.b }}{{ end }}[{{ template "inner" . }}]`, data, texttemplate.ExecuteOptions{NilSafe: true})
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(Equal("[]"))
	})
})
