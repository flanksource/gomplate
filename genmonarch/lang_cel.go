package genmonarch

import (
	"sort"
	"strings"

	"github.com/flanksource/gomplate/v3/genmonarch/grammar"
)

// CELLanguageID is the Monaco language id for CEL expressions.
const CELLanguageID = "cel"

// identifier is CEL's IDENTIFIER rule. It is read from the grammar rather than
// written here so it cannot disagree with the parser.
const identifierRule = "IDENTIFIER"

// BuildCEL assembles the CEL tokenizer from the lexical grammar and the live
// function catalogue.
func BuildCEL(g *grammar.CELGrammar, spec CELSpec) (Language, Configuration) {
	ident := g.Patterns[identifierRule]

	// Macros and functions are highlighted only in call position. A bare
	// `map` or `filter` is a perfectly ordinary variable name in CEL, and
	// colouring it as a keyword would be wrong more often than right.
	callAhead := `(?=\s*\()`

	lang := Language{
		ID:           CELLanguageID,
		DefaultToken: "",
		TokenPostfix: ".cel",
		Brackets: []Bracket{
			{Open: "{", Close: "}", Token: "delimiter.curly"},
			{Open: "[", Close: "]", Token: "delimiter.square"},
			{Open: "(", Close: ")", Token: "delimiter.parenthesis"},
		},
		Attributes: map[string][]string{
			"keywords":        reservedWords(spec.Keywords),
			"constants":       {"true", "false", "null"},
			"typeKeywords":    spec.Types,
			"macros":          macroNames(spec.Macros),
			"namespaces":      spec.Namespaces,
			"globalFunctions": leafNames(spec.GlobalNames()),
			"memberFunctions": spec.MemberNames(),
			"operators":       g.Operators,
		},
	}

	states := NewStates()

	states.Add("root",
		Include("@whitespace"),

		// Bytes before strings, and both before identifiers: `b"x"` starts with
		// a letter, so IDENTIFIER would otherwise claim the prefix.
		Match(g.Patterns["BYTES"], "string.bytes"),
		Match(g.Patterns["STRING"], "string"),
		Match(g.Patterns["ESC_IDENTIFIER"], "identifier.escaped"),

		// Float before uint before int: `1.5` must not tokenize as `1`, and
		// `123u` must not tokenize as `123` followed by an identifier.
		Match(g.Patterns["NUM_FLOAT"], "number.float"),
		Match(g.Patterns["NUM_UINT"], "number.uint"),
		Match(g.Patterns["NUM_INT"], "number"),

		// Optional-typed access, before the plain `.` and `[` operators. The
		// field after `.?` is still a field, so it is matched here rather than
		// being left to fall through to the bare-identifier rule.
		Rule{
			Regex: `(\.\?)(` + ident + `)` + callAhead,
			Action: Action{Cases: NewCases().
				Groups("$2@macros", "operator.optional", "keyword.macro").
				Groups("$2@memberFunctions", "operator.optional", "function.member").
				Groups("$2@globalFunctions", "operator.optional", "function").
				Default(Action{Tokens: []string{"operator.optional", "variable.field"}}),
			},
		},
		MatchGroups(`(\.\?)(`+ident+`)`, "operator.optional", "variable.field"),
		Match(`\.\?`, "operator.optional"),
		Match(`\[\?`, "operator.optional"),
		Match(`\?\.`, "operator.optional"),

		// `ns.fn(` / `x.member(` / `x.field`
		Rule{
			Regex: `(` + ident + `)(\.)(` + ident + `)` + callAhead,
			Action: Action{Cases: NewCases().
				Groups("$1@namespaces", "namespace", "delimiter", "function").
				Groups("$3@macros", "identifier", "delimiter", "keyword.macro").
				Groups("$3@memberFunctions", "identifier", "delimiter", "function.member").
				Groups("$3@globalFunctions", "identifier", "delimiter", "function").
				Default(Action{Tokens: []string{"identifier", "delimiter", "identifier"}}),
			},
		},
		// A member call. Receiver-style macros (`x.fold(...)`, `x.all(...)`)
		// are macros, not functions, and are checked first.
		Rule{
			Regex: `(\.)(` + ident + `)` + callAhead,
			Action: Action{Cases: NewCases().
				Groups("$2@macros", "delimiter", "keyword.macro").
				Groups("$2@memberFunctions", "delimiter", "function.member").
				Groups("$2@globalFunctions", "delimiter", "function").
				Default(Action{Tokens: []string{"delimiter", "variable.field"}}),
			},
		},
		MatchGroups(`(\.)(`+ident+`)`, "delimiter", "variable.field"),

		// A bare name in call position.
		Rule{
			Regex: `(` + ident + `)` + callAhead,
			Action: Action{Cases: NewCases().
				Token("$1@macros", "keyword.macro").
				Token("$1@globalFunctions", "function").
				Token("$1@keywords", "keyword").
				Default(Action{Token: "identifier"}),
			},
		},

		// A bare name anywhere else.
		Rule{
			Regex: ident,
			Action: Action{Cases: NewCases().
				Token("@constants", "keyword.constant").
				Token("@keywords", "keyword").
				Token("@typeKeywords", "type").
				Default(Action{Token: "identifier"}),
			},
		},

		Match(`[{}()\[\]]`, "@brackets"),
		Match(operatorPattern(g.Operators), "operator"),
		Match(`[,;]`, "delimiter"),
	)

	states.Add("whitespace",
		Match(g.Patterns["WHITESPACE"], "white"),
		Match(g.Patterns["COMMENT"], "comment"),
	)

	lang.Tokenizer = states

	config := Configuration{
		Comments: &Comments{LineComment: "//"},
		Brackets: [][2]string{{"{", "}"}, {"[", "]"}, {"(", ")"}},
		AutoClosingPairs: []Pair{
			{Open: "{", Close: "}"}, {Open: "[", Close: "]"}, {Open: "(", Close: ")"},
			{Open: `"`, Close: `"`}, {Open: "'", Close: "'"},
		},
		SurroundingPairs: []Pair{
			{Open: "{", Close: "}"}, {Open: "[", Close: "]"}, {Open: "(", Close: ")"},
			{Open: `"`, Close: `"`}, {Open: "'", Close: "'"},
		},
		// A dotted name is one word, so completing `k8s.isHealthy` replaces the
		// whole thing instead of appending to the namespace.
		WordPattern: `[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*`,
	}
	return lang, config
}

// operatorPattern renders the operator literals as one alternation, longest
// first so `<` never shadows `<=`. Brackets are excluded: they are matched
// separately so Monaco can pair and colour them.
func operatorPattern(operators []string) string {
	var parts []string
	for _, op := range operators {
		if strings.ContainsAny(op, "{}()[]") {
			continue
		}
		parts = append(parts, escapeRegexLiteral(op))
	}
	return "(?:" + strings.Join(parts, "|") + ")"
}

// reservedWords drops the literals that are better highlighted as constants.
func reservedWords(keywords []string) []string {
	var out []string
	for _, k := range keywords {
		switch k {
		case "true", "false", "null":
		default:
			out = append(out, k)
		}
	}
	return out
}

func macroNames(macros []Macro) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range macros {
		if !seen[m.Name] {
			seen[m.Name] = true
			out = append(out, m.Name)
		}
	}
	sort.Strings(out)
	return out
}

// leafNames reduces `k8s.isHealthy` to `isHealthy` and keeps un-namespaced
// names as they are. A tokenizer matches the leaf after the namespace has
// already been consumed as its own capture group.
func leafNames(names []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, name := range names {
		leaf := name
		if i := strings.LastIndex(name, "."); i >= 0 {
			leaf = name[i+1:]
		}
		if !seen[leaf] {
			seen[leaf] = true
			out = append(out, leaf)
		}
	}
	sort.Strings(out)
	return out
}
