package gomplate

import (
	"testing"
	"time"
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
