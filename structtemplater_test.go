package gomplate

import (
	"encoding/json"

	"github.com/flanksource/commons/test/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type Test struct {
	Template   string `template:"true"`
	NoTemplate string
	Inner      Inner
	Properties json.RawMessage   `template:"true" json:"properties,omitempty"`
	JSONMap    map[string]any    `template:"true"`
	Labels     map[string]string `template:"true"`
	LabelsRaw  map[string]string
	Slice      []string `template:"true"`
}

func (t *Test) Clone() Test {
	return Test{
		Template:   t.Template,
		NoTemplate: t.NoTemplate,
		Inner: Inner{
			Template:   t.Inner.Template,
			NoTemplate: t.Inner.NoTemplate,
		},
		Properties: append(json.RawMessage(nil), t.Properties...),
		JSONMap:    cloneMap(t.JSONMap),
		Labels:     cloneMap(t.Labels),
		LabelsRaw:  cloneMap(t.LabelsRaw),
		Slice:      append([]string(nil), t.Slice...),
	}
}
func cloneMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

type Inner struct {
	Template   string `template:"true"`
	NoTemplate string
}

type test struct {
	name string
	StructTemplater
	Input, Output *Test
	Vars          map[string]string
}

var tests = []test{
	{
		name: "template byte slice",
		StructTemplater: StructTemplater{
			RequiredTag: "template",
			Values: map[string]any{
				"msg": "world",
			},
		},
		Input: &Test{
			Properties: json.RawMessage(`{"name": "{{.msg}}"}`),
		},
		Output: &Test{
			Properties: json.RawMessage(`{"name": "world"}`),
		},
	},
	{
		name: "template and no template",
		StructTemplater: StructTemplater{
			RequiredTag: "template",
			Values: map[string]any{
				"msg": "world",
			},
		},
		Input: &Test{
			Template:   "hello {{.msg}}",
			NoTemplate: "hello {{.msg}}",
		},
		Output: &Test{
			Template:   "hello world",
			NoTemplate: "hello {{.msg}}",
			Properties: json.RawMessage{},
		},
	},

	{
		name: "pod manifest",
		StructTemplater: StructTemplater{
			RequiredTag:    "template",
			ValueFunctions: true,
			Values: map[string]any{
				"msg": "world",
			},
		},
		Input: &Test{
			JSONMap: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name":      "httpbin-{{msg}}",
					"namespace": "development",
					"labels": map[string]any{
						"app": "httpbin",
					},
				},
				"spec": map[string]any{
					"containers": []any{
						map[string]any{
							"name":  "httpbin",
							"image": "kennethreitz/httpbin:latest",
							"ports": []any{
								map[string]any{
									"containerPort": 80,
								},
							},
						},
					},
				},
			},
		},
		Output: &Test{
			Properties: json.RawMessage{},
			JSONMap: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]any{
					"name":      "httpbin-world",
					"namespace": "development",
					"labels": map[string]any{
						"app": "httpbin",
					},
				},
				"spec": map[string]any{
					"containers": []any{
						map[string]any{
							"name":  "httpbin",
							"image": "kennethreitz/httpbin:latest",
							"ports": []any{
								map[string]any{
									"containerPort": 80,
								},
							},
						},
					},
				},
			},
		},
	},
	{
		name: "just template",
		StructTemplater: StructTemplater{
			DelimSets: []Delims{
				{Left: "{{", Right: "}}"},
				{Left: "$(", Right: ")"},
			},
			Values: map[string]any{
				"msg": "world",
			},
			ValueFunctions: true,
		},
		Input: &Test{
			Template: "hello $(msg)",
		},
		Output: &Test{
			Template:   "hello world",
			Properties: json.RawMessage{},
		},
	},
	{
		name: "template & no template but with maps",
		StructTemplater: StructTemplater{
			RequiredTag: "template",
			DelimSets: []Delims{
				{Left: "{{", Right: "}}"},
				{Left: "$(", Right: ")"},
			},
			Values: map[string]any{
				"name":    "James Bond",
				"colorOf": "eye",
				"color":   "blue",
				"code":    "007",
				"city":    "London",
				"country": "UK",
			},
			ValueFunctions: true,
		},
		Input: &Test{
			Template: "Special Agent - $(name)!",
			Labels: map[string]string{
				"address":           "{{city}}, {{country}}",
				"{{colorOf}} color": "light $(color)",
				"code":              "{{code}}",
				"operation":         "noop",
			},
			LabelsRaw: map[string]string{
				"address":           "{{city}}, {{country}}",
				"{{colorOf}} color": "light $(color)",
			},
		},
		Output: &Test{
			Template:   "Special Agent - James Bond!",
			Properties: json.RawMessage{},
			Labels: map[string]string{
				"address":   "London, UK",
				"eye color": "light blue",
				"code":      "007",
				"operation": "noop",
			},
			LabelsRaw: map[string]string{
				"address":           "{{city}}, {{country}}",
				"{{colorOf}} color": "light $(color)",
			},
		},
	},

	{
		name: "slice",
		StructTemplater: StructTemplater{
			RequiredTag:    "template",
			ValueFunctions: true,
			Values: map[string]any{
				"msg": "world",
			},
		},
		Input: &Test{
			Slice: []string{
				"hello {{.msg}}",
			},
		},
		Output: &Test{
			Slice: []string{
				"hello world",
			},
		},
	},

	{
		StructTemplater: StructTemplater{
			RequiredTag:    "template",
			ValueFunctions: true,
			Values: map[string]any{
				"msg": "world",
			},
		},

		Input: &Test{
			Template: "{{msg}}",
			JSONMap: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "{{msg}}",
					},
					"j": []map[string]any{
						{
							"l": "{{msg}}",
						},
					},
				},
				"e": "hello {{msg}}",
			},
		},

		Output: &Test{
			Template:   "world",
			Properties: json.RawMessage{},
			JSONMap: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": "world",
					},
					"j": []any{
						map[string]any{
							"l": "world",
						},
					},
				},
				"e": "hello world",
			},
		},
	},
}

var _ = Describe("StructTemplater", func() {

	for _, test := range tests {
		It(test.name, func() {
			clone := test.Input.Clone()
			err := test.Walk(&clone)
			Expect(err).NotTo(HaveOccurred())
			Expect(clone).To(matchers.MatchJson(*test.Output))
		})

	}
})
