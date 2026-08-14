package genmonarch

import (
	"context"
	"fmt"
	"text/template"

	"github.com/ohler55/ojg/jp"

	gomplate "github.com/flanksource/gomplate/v3"
)

// ValidateCorpus parses every non-CEL snippet with the parser that actually
// evaluates it, so the corpus cannot drift into asserting behaviour for input
// gomplate would reject.
//
// CEL snippets are validated as they are built: celTokenBoundaries fails on any
// lexical error.
func ValidateCorpus(cases []ConformanceCase) error {
	funcs := gomplate.CreateFuncs(context.Background())

	for _, c := range cases {
		switch c.Language {
		case GoTemplateLanguageID:
			if _, err := template.New("conformance").Funcs(funcs).Parse(c.Source); err != nil {
				return fmt.Errorf("go template %q (%s): %w", c.Source, c.Origin, err)
			}
		case JSONPathLanguageID:
			if _, err := jp.ParseString(c.Source); err != nil {
				return fmt.Errorf("jsonpath %q (%s): %w", c.Source, c.Origin, err)
			}
		}
	}
	return nil
}
