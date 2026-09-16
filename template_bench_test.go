package gomplate

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
)

func BenchmarkRunTemplateWithoutCache(b *testing.B) {
	environment := map[string]any{
		"name": "node-a",
		"tags": map[string]string{"cluster": "production"},
	}
	template := Template{
		Template:       "node://kubernetes/$(tags.cluster)/$(.name)",
		ValueFunctions: true,
		DelimSets: []Delims{
			{Left: "{{", Right: "}}"},
			{Left: "$(", Right: ")"},
		},
	}

	output, err := RunTemplate(environment, template)
	if err != nil {
		b.Fatal(err)
	}
	if output != "node://kubernetes/production/node-a" {
		b.Fatalf("unexpected warm-up result: %q", output)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := RunTemplate(environment, template); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRunTemplateWithCache(b *testing.B) {
	environment := map[string]any{
		"name": "node-a",
		"tags": map[string]string{"cluster": "production"},
	}
	template := Template{
		Template:       "node://kubernetes/$(tags.cluster)/$(.name)",
		CacheKey:       "benchmark-go-template-cache",
		CacheTime:      time.Hour,
		ValueFunctions: true,
		DelimSets: []Delims{
			{Left: "{{", Right: "}}"},
			{Left: "$(", Right: ")"},
		},
	}

	output, err := RunTemplate(environment, template)
	if err != nil {
		b.Fatal(err)
	}
	if output != "node://kubernetes/production/node-a" {
		b.Fatalf("unexpected warm-up result: %q", output)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := RunTemplate(environment, template); err != nil {
			b.Fatal(err)
		}
	}
}

type runTemplateBenchmarkScenario struct {
	name        string
	environment func(uint64) map[string]any
	template    func(uint64) Template
	expected    func(uint64) string
}

// Go templates ignore CelEnvs, so this option changes only cache eligibility.
var runTemplateBenchmarkNoopEnvOption cel.EnvOption = func(environment *cel.Env) (*cel.Env, error) {
	return environment, nil
}

func BenchmarkRunTemplateCacheMatrix(b *testing.B) {
	staticEnvironment := map[string]any{
		"service": "api",
		"region":  "us-east-1",
	}
	unusedFunction := func() string { return "unused" }
	longEnvironment := map[string]any{
		"service": "api",
		"items": []any{
			map[string]any{"name": "alpha", "value": "one"},
			map[string]any{"name": "beta", "value": ""},
			map[string]any{"name": "gamma", "value": "three"},
		},
	}
	largeEnvironment := runTemplateBenchmarkLargeEnvironment()

	scenarios := []runTemplateBenchmarkScenario{
		{
			name:        "static-dot",
			environment: func(uint64) map[string]any { return staticEnvironment },
			template: func(uint64) Template {
				return Template{
					Template:  "{{ .service }}@{{ .region }}",
					Functions: map[string]any{"benchmarkUnused": unusedFunction},
				}
			},
			expected: func(uint64) string { return "api@us-east-1" },
		},
		{
			name: "value-functions",
			environment: func(iteration uint64) map[string]any {
				if iteration%2 == 0 {
					return map[string]any{"greeting": "hello", "name": "Ada", "sequence": 17}
				}
				return map[string]any{"greeting": "welcome", "name": "Linus", "sequence": 42}
			},
			template: func(uint64) Template {
				return Template{
					Template:       "{{ greeting }}, {{ name }} #{{ .sequence }}",
					ValueFunctions: true,
				}
			},
			expected: func(iteration uint64) string {
				if iteration%2 == 0 {
					return "hello, Ada #17"
				}
				return "welcome, Linus #42"
			},
		},
		{
			name: "custom-functions",
			environment: func(iteration uint64) map[string]any {
				if iteration%2 == 0 {
					return map[string]any{"name": "Ada"}
				}
				return map[string]any{"name": "Linus"}
			},
			template: func(iteration uint64) Template {
				decorate := func(value string) string { return "A<" + value + ">" }
				if iteration%2 != 0 {
					decorate = func(value string) string { return "B[" + value + "]" }
				}
				return Template{
					Template:  "{{ decorate .name }}",
					Functions: map[string]any{"decorate": decorate},
				}
			},
			expected: func(iteration uint64) string {
				if iteration%2 == 0 {
					return "A<Ada>"
				}
				return "B[Linus]"
			},
		},
		{
			name:        "named-template-builtins",
			environment: func(uint64) map[string]any { return longEnvironment },
			template: func(uint64) Template {
				return Template{
					Template: `{{ define "label" }}{{ . | toUpper }}{{ end }}{{ define "row" }}{{ template "label" .name }}={{ .value | default "unknown" | quote }}{{ end }}service={{ template "label" .service }};{{ range .items }}[{{ template "row" . }}]{{ end }}`,
					CelEnvs:  []cel.EnvOption{runTemplateBenchmarkNoopEnvOption},
				}
			},
			expected: func(uint64) string {
				return `service=API;[ALPHA="one"][BETA="unknown"][GAMMA="three"]`
			},
		},
		{
			name: "changing-environment",
			environment: func(iteration uint64) map[string]any {
				if iteration%2 == 0 {
					return map[string]any{"region": "eu-west-1", "node": "worker-a", "replicas": 3}
				}
				return map[string]any{"region": "ap-south-1", "node": "worker-z", "replicas": 11}
			},
			template: func(uint64) Template {
				return Template{
					Template: "{{ .region }}/{{ .node }}/{{ .replicas }}",
					CelEnvs:  []cel.EnvOption{runTemplateBenchmarkNoopEnvOption},
				}
			},
			expected: func(iteration uint64) string {
				if iteration%2 == 0 {
					return "eu-west-1/worker-a/3"
				}
				return "ap-south-1/worker-z/11"
			},
		},
		{
			name: "two-delimiter-expansion",
			environment: func(uint64) map[string]any {
				return map[string]any{
					"service": "payments",
					"region":  "us-west-2",
					"version": "v7",
					"zone":    "b",
				}
			},
			template: func(uint64) Template {
				return Template{
					Template: "{{ .service }}/$(.region)/{{ .version }}-$(.zone)",
					DelimSets: []Delims{
						{Left: "{{", Right: "}}"},
						{Left: "$(", Right: ")"},
					},
					CelEnvs: []cel.EnvOption{runTemplateBenchmarkNoopEnvOption},
				}
			},
			expected: func(uint64) string { return "payments/us-west-2/v7-b" },
		},
		{
			name:        "large-environment",
			environment: func(uint64) map[string]any { return largeEnvironment },
			template: func(uint64) Template {
				return Template{
					Template: "{{ .key000 }}|{{ .key127 }}|{{ .key255 }}|{{ .nested.value }}",
					CelEnvs:  []cel.EnvOption{runTemplateBenchmarkNoopEnvOption},
				}
			},
			expected: func(uint64) string { return "value-000|value-127|value-255|deep-value" },
		},
	}

	// Every pair keeps the fixture fixed; cache=explicit only adds CacheKey.
	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			for _, explicitCache := range []bool{false, true} {
				cacheName := "disabled"
				if explicitCache {
					cacheName = "explicit"
				}
				b.Run("cache="+cacheName, func(b *testing.B) {
					executeRunTemplateBenchmarkScenario(b, scenario, explicitCache)
				})
			}
		})
	}

	benchmarkRunTemplateParallel(b)
}

func BenchmarkRunTemplateNaturalStaticWarmHit(b *testing.B) {
	environment := map[string]any{"service": "api", "region": "us-east-1"}
	template := Template{Template: "{{ .service }}@{{ .region }}"}
	const expected = "api@us-east-1"

	output, err := RunTemplate(environment, template)
	if err != nil {
		b.Fatal(err)
	}
	if output != expected {
		b.Fatalf("unexpected warm-up result: got %q, want %q", output, expected)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		output, err := RunTemplate(environment, template)
		if err != nil {
			b.Fatal(err)
		}
		if output != expected {
			b.Fatalf("unexpected result: got %q, want %q", output, expected)
		}
	}
}

func BenchmarkRunTemplateColdCache(b *testing.B) {
	for _, valueFunctions := range []bool{false, true} {
		name, source := "static", "{{ .service }}@{{ .region }}"
		if valueFunctions {
			name, source = "value-functions", "{{ service }}@{{ region }}"
		}
		b.Run(name, func(b *testing.B) {
			environment := map[string]any{"service": "api", "region": "us-east-1"}
			template := Template{
				Template:       source,
				ValueFunctions: valueFunctions,
				CacheKey:       "benchmark-cold-" + name,
				// Expiration forces recompilation without accumulating unique cache keys.
				CacheTime: time.Nanosecond,
			}
			b.ReportAllocs()
			for b.Loop() {
				output, err := RunTemplate(environment, template)
				if err != nil {
					b.Fatal(err)
				}
				if output != "api@us-east-1" {
					b.Fatalf("unexpected result: %q", output)
				}
			}
		})
	}
}

func executeRunTemplateBenchmarkScenario(b *testing.B, scenario runTemplateBenchmarkScenario, explicitCache bool) {
	b.Helper()
	template := scenario.template(0)
	if explicitCache {
		template.CacheKey = "benchmark-run-template-" + scenario.name
	}
	output, err := RunTemplate(scenario.environment(0), template)
	if err != nil {
		b.Fatal(err)
	}
	if expected := scenario.expected(0); output != expected {
		b.Fatalf("unexpected warm-up result: got %q, want %q", output, expected)
	}

	b.ReportAllocs()
	b.ResetTimer()
	var iteration uint64
	for b.Loop() {
		template := scenario.template(iteration)
		if explicitCache {
			template.CacheKey = "benchmark-run-template-" + scenario.name
		}
		output, err := RunTemplate(scenario.environment(iteration), template)
		if err != nil {
			b.Fatal(err)
		}
		if expected := scenario.expected(iteration); output != expected {
			b.Fatalf("unexpected result at iteration %d: got %q, want %q", iteration, output, expected)
		}
		iteration++
	}
}

func benchmarkRunTemplateParallel(b *testing.B) {
	b.Run("parallel-rendering", func(b *testing.B) {
		for _, explicitCache := range []bool{false, true} {
			cacheName := "disabled"
			if explicitCache {
				cacheName = "explicit"
			}
			b.Run("cache="+cacheName, func(b *testing.B) {
				template := Template{
					Template: "{{ .worker }}:{{ .region }}:{{ .sequence }}",
					CelEnvs:  []cel.EnvOption{runTemplateBenchmarkNoopEnvOption},
				}
				if explicitCache {
					template.CacheKey = "benchmark-run-template-parallel-rendering"
				}
				warmEnvironment := map[string]any{"worker": "worker-a", "region": "east", "sequence": 13}
				output, err := RunTemplate(warmEnvironment, template)
				if err != nil {
					b.Fatal(err)
				}
				if output != "worker-a:east:13" {
					b.Fatalf("unexpected warm-up result: got %q, want %q", output, "worker-a:east:13")
				}

				b.ReportAllocs()
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					// Serialize may mutate nested maps, so workers never share an environment.
					environment := map[string]any{"worker": "worker-a", "region": "east", "sequence": 13}
					var iteration uint64
					for pb.Next() {
						expected := "worker-a:east:13"
						if iteration%2 != 0 {
							environment["worker"] = "worker-z"
							environment["region"] = "west"
							environment["sequence"] = 97
							expected = "worker-z:west:97"
						} else {
							environment["worker"] = "worker-a"
							environment["region"] = "east"
							environment["sequence"] = 13
						}
						output, err := RunTemplate(environment, template)
						if err != nil {
							b.Errorf("render failed: %v", err)
							return
						}
						if output != expected {
							b.Errorf("unexpected result: got %q, want %q", output, expected)
							return
						}
						iteration++
					}
				})
			})
		}
	})
}

func runTemplateBenchmarkLargeEnvironment() map[string]any {
	environment := make(map[string]any, 257)
	for index := range 256 {
		environment[fmt.Sprintf("key%03d", index)] = fmt.Sprintf("value-%03d", index)
	}
	environment["nested"] = map[string]any{
		"value": "deep-value",
		"items": []any{
			map[string]any{"id": 1, "enabled": true},
			map[string]any{"id": 2, "enabled": false},
		},
	}
	return environment
}
