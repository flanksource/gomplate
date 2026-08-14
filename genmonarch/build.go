package genmonarch

import (
	"fmt"

	"github.com/google/cel-go/cel"

	"github.com/flanksource/gomplate/v3/genmonarch/grammar"
)

// Bundle is everything the npm package ships: one Monarch definition and one
// language configuration per language id, plus the shared function catalogue
// and the conformance corpus.
type Bundle struct {
	Spec           Spec
	Languages      map[string]Language
	Configurations map[string]Configuration
	Conformance    []ConformanceCase
	// Order lists the language ids deterministically, for stable file output.
	Order []string
}

// Build assembles every language from gomplate's own grammars and registries.
// docs supplies the Markdown references the conformance corpus draws snippets
// from, keyed by filename. extraCEL layers a host's own CEL options in, so a
// binary that registers extra functions can generate a bundle that knows them.
func Build(docs map[string]string, extraCEL ...cel.EnvOption) (*Bundle, error) {
	celGrammar, err := grammar.ParseCEL()
	if err != nil {
		return nil, fmt.Errorf("reading the CEL grammar: %w", err)
	}
	celSpec, err := ExtractCEL(extraCEL...)
	if err != nil {
		return nil, fmt.Errorf("reading the CEL environment: %w", err)
	}
	goSpec, err := ExtractGoTemplate()
	if err != nil {
		return nil, fmt.Errorf("reading the go-template functions: %w", err)
	}

	bundle := &Bundle{
		Spec:           Spec{CEL: celSpec, GoTemplate: goSpec},
		Languages:      map[string]Language{},
		Configurations: map[string]Configuration{},
	}

	celLang, celConfig := BuildCEL(celGrammar, celSpec)
	bundle.add(celLang, celConfig)

	goLang, goConfig, err := BuildGoTemplate(goSpec, goSpec.Delimiters)
	if err != nil {
		return nil, fmt.Errorf("building the gomplate language: %w", err)
	}
	bundle.add(goLang, goConfig)

	for _, host := range []Host{HostYAML, HostJSON, HostText} {
		lang, config, err := BuildEmbedded(host, goSpec, goSpec.Delimiters)
		if err != nil {
			return nil, fmt.Errorf("building the %s host language: %w", host, err)
		}
		bundle.add(lang, config)
	}

	jsonPathLang, jsonPathConfig := BuildJSONPath()
	bundle.add(jsonPathLang, jsonPathConfig)

	bundle.Conformance, err = BuildConformance(docs)
	if err != nil {
		return nil, fmt.Errorf("building the conformance corpus: %w", err)
	}
	if err := ValidateCorpus(bundle.Conformance); err != nil {
		return nil, fmt.Errorf("validating the conformance corpus: %w", err)
	}

	return bundle, nil
}

func (b *Bundle) add(lang Language, config Configuration) {
	b.Languages[lang.ID] = lang
	b.Configurations[lang.ID] = config
	b.Order = append(b.Order, lang.ID)
}
