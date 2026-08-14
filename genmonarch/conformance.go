package genmonarch

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	celparser "github.com/google/cel-go/parser/gen"
)

// ConformanceCase is one snippet with the token boundaries the language's real
// lexer produces.
//
// The generator sits next to the parsers gomplate evaluates with, so it can
// produce the oracle rather than a snapshot of the tokenizer's own output. The
// browser test replays these through Monaco and asserts the boundaries agree.
//
// Boundaries, not token classes: Monarch says `namespace` and `function` where
// the lexer only says IDENTIFIER, so the classes are not comparable. The
// boundaries are, and they are where the subtle bugs live -- a triple-quoted
// string cut short, `0x1f` truncated to `0`, `123u` split into a number and an
// identifier.
type ConformanceCase struct {
	Language string `json:"language"`
	Source   string `json:"source"`
	// Boundaries are 0-based offsets where a token starts, excluding
	// whitespace, in ascending order.
	Boundaries []int `json:"boundaries"`
	// Origin records where the snippet came from, so a failure is traceable.
	Origin string `json:"origin"`
}

// celEdgeCases are the lexical corners no documentation example happens to
// cover. Each one is a shape that a hand-written tokenizer gets wrong.
var celEdgeCases = []string{
	"`escaped.identifier-1`",
	`"""triple "quoted" string"""`,
	`'''triple 'quoted' string'''`,
	`r"raw\dstring"`,
	`R'raw\dstring'`,
	`r"""raw triple \d"""`,
	`b"bytes"`,
	`B'bytes'`,
	`"\x41A\U0001F600\101"`,
	`123u + 0x1fU + 0x1f + 1.5e-3 + .5`,
	`a.?b.orValue("x")`,
	`m[?"k"]`,
	`cond ? "yes" : "no"`,
	`[1, 2, 3].fold(e, acc, acc + e)`,
	`k8s.isHealthy(pod) && pod.status.?phase.orValue("") == "Running"`,
	`"a" + // trailing comment`,
}

// goTemplateEdgeCases exercise delimiters, trimming and comments.
var goTemplateEdgeCases = []string{
	`{{ .name | strings.ToUpper }}`,
	`{{- if .enabled -}}on{{- else -}}off{{- end -}}`,
	`{{/* a comment with {{ braces }} in it */}}`,
	"{{ $x := coll.Dict \"a\" 1 }}{{ $x }}",
	"{{ printf \"%s-%d\" .name 3 }}",
	"{{ `raw string` }}",
}

var jsonPathEdgeCases = []string{
	`$.store.book[0].title`,
	`$..author`,
	`$.store.book[?(@.price < 10)]`,
	`$.items[*]`,
	`$['quoted key']`,
	`$.items[0:2]`,
}

// fencedBlock matches a fenced code block in the Markdown references.
var fencedBlock = regexp.MustCompile("(?s)```[a-zA-Z]*\n(.*?)```")

// BuildConformance assembles the corpus. docs maps a filename to its contents;
// the caller reads them so this stays testable without touching the disk.
func BuildConformance(docs map[string]string) ([]ConformanceCase, error) {
	var cases []ConformanceCase

	for _, source := range celEdgeCases {
		c, err := celCase(source, "edge-case")
		if err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}

	for _, name := range sortedKeys(docs) {
		for _, source := range celSnippetsFrom(docs[name]) {
			c, err := celCase(source, name)
			if err != nil {
				// A snippet the real lexer rejects is prose, not code.
				continue
			}
			cases = append(cases, c)
		}
	}

	for _, source := range goTemplateEdgeCases {
		cases = append(cases, ConformanceCase{
			Language: GoTemplateLanguageID,
			Source:   source,
			Origin:   "edge-case",
		})
	}
	for _, source := range jsonPathEdgeCases {
		cases = append(cases, ConformanceCase{
			Language: JSONPathLanguageID,
			Source:   source,
			Origin:   "edge-case",
		})
	}

	return dedupeCases(cases), nil
}

// celCase lexes a snippet with cel-go's own ANTLR lexer and records where each
// token starts.
func celCase(source, origin string) (ConformanceCase, error) {
	boundaries, err := celTokenBoundaries(source)
	if err != nil {
		return ConformanceCase{}, err
	}
	return ConformanceCase{
		Language:   CELLanguageID,
		Source:     source,
		Boundaries: boundaries,
		Origin:     origin,
	}, nil
}

// celTokenBoundaries runs the generated CEL lexer and returns the start offset
// of every non-whitespace token.
func celTokenBoundaries(source string) ([]int, error) {
	lexer := celparser.NewCELLexer(antlr.NewInputStream(source))
	lexer.RemoveErrorListeners()

	failed := &lexErrorListener{}
	lexer.AddErrorListener(failed)

	var boundaries []int
	for {
		token := lexer.NextToken()
		if token == nil || token.GetTokenType() == antlr.TokenEOF {
			break
		}
		if token.GetTokenType() == celparser.CELLexerWHITESPACE {
			continue
		}
		boundaries = append(boundaries, token.GetStart())
	}
	if failed.err != nil {
		return nil, failed.err
	}
	if len(boundaries) == 0 {
		return nil, fmt.Errorf("no tokens in %q", source)
	}
	return boundaries, nil
}

// lexErrorListener records the first lexical error, so a snippet the real lexer
// rejects never reaches the corpus.
type lexErrorListener struct {
	*antlr.DefaultErrorListener
	err error
}

func (l *lexErrorListener) SyntaxError(_ antlr.Recognizer, _ any, line, column int, msg string, _ antlr.RecognitionException) {
	if l.err == nil {
		l.err = fmt.Errorf("%d:%d: %s", line, column, msg)
	}
}

// celSnippetsFrom pulls single-line CEL expressions out of a Markdown
// reference. Comment-only and prose lines are skipped; anything the lexer
// rejects is dropped by the caller.
func celSnippetsFrom(markdown string) []string {
	var out []string
	for _, block := range fencedBlock.FindAllStringSubmatch(markdown, -1) {
		for _, line := range strings.Split(block[1], "\n") {
			line = strings.TrimSpace(line)
			// `//` opens a CEL comment, and the references use it to show the
			// expected value, so keep only what precedes it.
			if i := strings.Index(line, "//"); i >= 0 {
				line = strings.TrimSpace(line[:i])
			}
			if len(line) < 3 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "$") {
				continue
			}
			// Anything with template or shell syntax is not a CEL expression.
			if strings.ContainsAny(line, "{}") && strings.Contains(line, "{{") {
				continue
			}
			out = append(out, line)
		}
	}
	return out
}

func dedupeCases(cases []ConformanceCase) []ConformanceCase {
	sort.SliceStable(cases, func(i, j int) bool {
		if cases[i].Language != cases[j].Language {
			return cases[i].Language < cases[j].Language
		}
		return cases[i].Source < cases[j].Source
	})
	out := cases[:0]
	for i, c := range cases {
		if i > 0 && c.Language == cases[i-1].Language && c.Source == cases[i-1].Source {
			continue
		}
		out = append(out, c)
	}
	return out
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
