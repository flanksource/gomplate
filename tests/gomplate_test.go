package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/flanksource/gomplate/v3"
	_ "github.com/flanksource/gomplate/v3/js"
	"github.com/flanksource/gomplate/v3/kubernetes"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/stretchr/testify/assert"
)

func TestGomplateFunctions(t *testing.T) {
	funcs := map[string]any{
		"fn": func() any {
			return map[string]any{
				"a": "b",
				"c": 1,
			}
		},
		"fn1": func() any {
			return "c"
		},
	}

	out, err := gomplate.RunTemplate(map[string]interface{}{
		"hello": "hi",
	}, gomplate.Template{
		Template:  "{{.hello}} {{fn1}}",
		Functions: funcs,
	})

	assert.ErrorIs(t, nil, err)
	assert.Equal(t, "hi c", out)
}

func TestDelimeters(t *testing.T) {
	tests := []struct {
		env      map[string]interface{}
		template string
		out      string
	}{
		{map[string]interface{}{"hello": "world"}, "[[ .hello ]]", "world"},
		{map[string]interface{}{"hello": "world"}, "$(.hello)", "$(.hello)"},
	}

	for _, tt := range tests {
		out, err := gomplate.RunTemplate(tt.env, gomplate.Template{
			Template:   tt.template,
			RightDelim: "]]",
			LeftDelim:  "[[",
		})
		assert.NoError(t, err)
		assert.Equal(t, tt.out, out)
	}
}

func TestGomplate(t *testing.T) {
	tests := []struct {
		env      map[string]interface{}
		template string
		out      string
	}{
		{map[string]interface{}{"hello": "world"}, "{{ .hello }}", "world"},
		{map[string]interface{}{"hello": "hello world ?"}, "{{ .hello | urlencode }}", `hello+world+%3F`},
		{map[string]interface{}{"hello": "hello+world+%3F"}, "{{ .hello | urldecode }}", `hello world ?`},
		{map[string]interface{}{"age": 75 * time.Second}, "{{ .age | humanDuration  }}", "1m15s"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructured(kubernetes.TestHealthyCertificate)}, "{{(.healthySvc | isHealthy)}}", "true"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructured(kubernetes.TestHealthyCertificate)}, "{{(.healthySvc | isReady)}}", "true"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructured(kubernetes.TestDegradedCertificate)}, "{{(.healthySvc | isHealthy)}}", "false"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructured(kubernetes.TestHealthySvc)}, "{{ (.healthySvc | isHealthy) }}", "true"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructured(kubernetes.TestLuaStatus)}, "{{ (.healthySvc | getStatus) }}", ": found less than two generators, Merge requires two or more"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructured(kubernetes.TestHealthySvc)}, "{{ (.healthySvc | getHealth).Health  }}", "healthy"},
		{map[string]interface{}{"size": 123456}, "{{ .size | humanSize }}", "120.6K"},
		{map[string]interface{}{"v": "1.2.3-beta.1+c0ff33"}, "{{  (.v | semver).Prerelease  }}", "beta.1"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.3"}, "{{  .old | semverCompare .new }}", "true"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.4"}, "{{  .old | semverCompare .new }}", "false"},

		{map[string]interface{}{"i": person}, `{{ index (jsonpath "$.addresses[-1:].city_name" .i) 0}}`, "New York"},
		{map[string]interface{}{"i": person}, `{{ .i | jmespath "addresses[*].city_name | [0]"}}`, "Kathmandu"},
		{map[string]interface{}{"i": person}, `{{ .i | jmespath "length(addresses)"}}`, "3"},

		{map[string]interface{}{"kv": "a=b,c=d"}, "{{ (.kv | keyValToMap).a }}", "b"},
		{inline, `{{.Name}}`, "Jane Doe"},
		{inline, `{{ (.Data |  base64.Decode | json).name }}`, "John Doe"},
		{inline, `{{ (.Data |  base64.Decode ) | jq ".addresses[0].city_name" }}`, "Kathmandu"},
		{inline, `{{ $city := (.Data |  base64.Decode ) | jq ".addresses[0].city_name" }}
			{{ $city | strings.ToLower }}
		`, "kathmandu"},
		// {inline, `{{ (.Data |  base64.Decode | json) | jsonPath ".name" }}`, "John Doe"},
		{structEnv, `{{.results.name}} {{.results.Address.city_name}}`, "Aditya Kathmandu"},
		{
			map[string]any{"results": junitEnv},
			`{{.results.passed}}{{ range $r := .results.suites}}{{$r.name}} ✅ {{$r.passed}} ❌ {{$r.failed}} in 🕑 {{$r.duration}}{{end}}`,
			"1hi ✅ 0 ❌ 2 in 🕑 0",
		},
		{
			map[string]any{
				"results": SQLDetails{
					Rows: []map[string]any{{"name": "apm-hub"}, {"name": "config-db"}},
				},
			},
			`{{range $r := .results.rows }}{{range $x, $y := $r }}{{ $y }}{{end}}{{end}}`, "apm-hubconfig-db"},
	}

	for _, tc := range tests {
		t.Run(tc.template, func(t *testing.T) {
			out, err := gomplate.RunTemplate(tc.env, gomplate.Template{
				Template: tc.template,
			})
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestGomplateHeaders(t *testing.T) {
	tests := []struct {
		env       map[string]interface{}
		template  string
		out       string
		expectErr bool
	}{
		{map[string]interface{}{"name": "world"}, readFile(t, "testdata/gotemplate/template-multiple-headers.yaml"), readFile(t, "testdata/gotemplate/expected/template-multiple-headers.yaml"), false},
		{map[string]interface{}{"name": "world"}, readFile(t, "testdata/gotemplate/template-header.txt"), "Hello, world", false},
		{map[string]interface{}{"name": "world"}, readFile(t, "testdata/gotemplate/bad-template-header.txt"), "", true},
		{map[string]interface{}{"name": "world"}, readFile(t, "testdata/gotemplate/template-header-override.txt"), "Hello, world. This ${should} not be {{touched}}", false},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d", i+1), func(t *testing.T) {
			out, err := gomplate.RunTemplate(tc.env, gomplate.Template{
				Template: tc.template,
			})
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.ErrorIs(t, err, nil)
				assert.Equal(t, tc.out, out)
			}
		})
	}
}

func readFile(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filename)
	assert.ErrorIs(t, err, nil)
	return string(data)
}
