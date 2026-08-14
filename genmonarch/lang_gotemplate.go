package genmonarch

import (
	"fmt"
	"sort"
	"strings"
)

// GoTemplateLanguageID is the Monaco language id for a bare gomplate template.
const GoTemplateLanguageID = "gomplate"

// goTemplateIdent is text/template's identifier shape. Unlike CEL, this is not
// in a published grammar -- the lexer accepts any alphanumeric run -- so it is
// stated once here.
const goTemplateIdent = `[A-Za-z_][A-Za-z0-9_]*`

// actionStates are the tokenizer states that apply *inside* `{{ ... }}`. They
// are built once and shared by the bare gomplate language and every embedded
// host variant, so a fix to template highlighting lands in all of them at once.
//
// The state names are prefixed so they cannot collide with a host language's
// own states when the two are spliced into one tokenizer.
func actionStates(spec GoTemplateSpec, d Delimiters) (states *States, entry []Rule, err error) {
	if d.Left == "" || d.Right == "" {
		return nil, nil, fmt.Errorf("template delimiters must not be empty")
	}
	left, right := escapeRegexLiteral(d.Left), escapeRegexLiteral(d.Right)
	trim := escapeRegexLiteral(d.TrimMarker)
	leftComment, rightComment := escapeRegexLiteral(d.LeftComment), escapeRegexLiteral(d.RightComment)

	// A comment opens with the delimiter immediately followed by `/*`, so it
	// has to be tried before a plain action.
	entry = []Rule{
		Push(left+trim+"?"+leftComment, "comment", "@tmplComment"),
		Push(left+trim+"?", "delimiter.template", "@tmplAction"),
	}

	states = NewStates()

	states.Add("tmplComment",
		Pop(rightComment+trim+"?"+right, "comment"),
		Match(`[\s\S]`, "comment"),
	)

	states.Add("tmplAction",
		Pop(trim+"?"+right, "delimiter.template"),
		Match(`[ \t\r\n]+`, "white"),

		// Strings: interpreted, raw (back-quoted) and rune literals.
		Match(`"(?:[^"\\]|\\.)*"`, "string"),
		Match("`[^`]*`", "string"),
		Match(`'(?:[^'\\]|\\.)*'`, "string"),

		Match(`[+-]?(?:0[xX][0-9a-fA-F]+|(?:\d+\.\d*|\.\d+|\d+)(?:[eE][+-]?\d+)?)`, "number"),

		// `$`, `$x`, `$x :=`
		Match(`\$`+goTemplateIdent, "variable"),
		Match(`\$`, "variable"),

		// A leading dot is a field path, never a namespace.
		MatchGroups(`(\.)(`+goTemplateIdent+`)`, "delimiter", "variable.field"),
		Match(`\.`, "variable.field"),

		// `strings.ToUpper` -- only when the prefix is a real namespace.
		Rule{
			Regex: `(` + goTemplateIdent + `)(\.)(` + goTemplateIdent + `)`,
			Action: Action{Cases: NewCases().
				Groups("$1@namespaces", "namespace", "delimiter", "function").
				Default(Action{Tokens: []string{"identifier", "delimiter", "identifier"}}),
			},
		},

		Rule{
			Regex: goTemplateIdent,
			Action: Action{Cases: NewCases().
				Token("@keywords", "keyword").
				Token("@builtins", "function.builtin").
				Token("@functions", "function").
				Default(Action{Token: "identifier"}),
			},
		},

		Match(`[()\[\]]`, "@brackets"),
		Match(`\|`, "operator.pipe"),
		Match(`:=|=`, "operator"),
		Match(`,`, "delimiter"),
	)

	return states, entry, nil
}

// BuildGoTemplate assembles the bare gomplate language: literal text with
// `{{ ... }}` actions embedded in it.
func BuildGoTemplate(spec GoTemplateSpec, d Delimiters) (Language, Configuration, error) {
	states, entry, err := actionStates(spec, d)
	if err != nil {
		return Language{}, Configuration{}, err
	}

	root := append([]Rule{directiveRule()}, entry...)
	// Everything that is not an action is literal output.
	root = append(root, Match(`[\s\S]`, "source"))

	lang := Language{
		ID:           GoTemplateLanguageID,
		DefaultToken: "source",
		TokenPostfix: ".gomplate",
		Brackets: []Bracket{
			{Open: "(", Close: ")", Token: "delimiter.parenthesis"},
			{Open: "[", Close: "]", Token: "delimiter.square"},
		},
		Attributes: goTemplateAttributes(spec),
		Tokenizer:  NewStates().Add("root", root...),
	}
	for _, name := range states.Names() {
		lang.Tokenizer.Add(name, states.Rules(name)...)
	}

	config := Configuration{
		Comments: &Comments{BlockComment: [2]string{d.Left + d.LeftComment, d.RightComment + d.Right}},
		Brackets: [][2]string{{d.Left, d.Right}, {"(", ")"}, {"[", "]"}},
		AutoClosingPairs: []Pair{
			{Open: d.Left, Close: " " + d.Right}, {Open: "(", Close: ")"},
			{Open: `"`, Close: `"`}, {Open: "`", Close: "`"},
		},
		SurroundingPairs: []Pair{
			{Open: "(", Close: ")"}, {Open: `"`, Close: `"`}, {Open: "`", Close: "`"},
		},
		WordPattern: `[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*`,
	}
	return lang, config, nil
}

// directiveRule highlights gomplate's own delimiter header, which reconfigures
// the parser from inside the template (see Template.parseHeader).
func directiveRule() Rule {
	return Match(`^#\s*gotemplate:.*$`, "comment.directive")
}

func goTemplateAttributes(spec GoTemplateSpec) map[string][]string {
	var topLevel []string
	for _, f := range spec.Functions {
		if f.Namespace == "" {
			topLevel = append(topLevel, f.Name)
		}
	}
	sort.Strings(topLevel)

	return map[string][]string{
		"keywords":   spec.Keywords,
		"builtins":   spec.Builtins,
		"namespaces": spec.Namespaces,
		"functions":  topLevel,
	}
}

// namespacedLeaves lists the method names under each namespace, for completion.
func namespacedLeaves(spec GoTemplateSpec) []string {
	var out []string
	for _, f := range spec.Functions {
		if f.Namespace != "" {
			out = append(out, strings.TrimPrefix(f.Name, f.Namespace+"."))
		}
	}
	sort.Strings(out)
	return dedupe(out)
}
