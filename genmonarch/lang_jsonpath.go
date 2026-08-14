package genmonarch

// JSONPathLanguageID is the Monaco language id for JSONPath expressions.
const JSONPathLanguageID = "jsonpath"

// jsonPathFilters are the operators ojg's parser accepts inside `[?(...)]`.
//
// Unlike CEL and text/template, the JSONPath dialect gomplate evaluates
// (github.com/ohler55/ojg/jp, via coll.JSONPath) is a hand-written Go lexer with
// no grammar to read, so this vocabulary is declared rather than derived. The
// conformance corpus compensates: every snippet is round-tripped through the
// real jp.ParseString, so a token listed here that the parser rejects fails the
// build.
var jsonPathFilters = []string{
	"==", "!=", "<=", ">=", "&&", "||", "=~",
	"<", ">", "!", "+", "-", "*", "/",
}

// BuildJSONPath assembles the JSONPath tokenizer.
func BuildJSONPath() (Language, Configuration) {
	lang := Language{
		ID:           JSONPathLanguageID,
		DefaultToken: "",
		TokenPostfix: ".jsonpath",
		Brackets: []Bracket{
			{Open: "[", Close: "]", Token: "delimiter.square"},
			{Open: "(", Close: ")", Token: "delimiter.parenthesis"},
		},
		Attributes: map[string][]string{
			"filterOperators": jsonPathFilters,
		},
		Tokenizer: NewStates().Add("root",
			Match(`\$`, "variable.root"),
			Match(`@`, "variable.current"),

			// Recursive descent before the plain child separator, or `..name`
			// tokenizes as two separate steps.
			Match(`\.\.`, "operator.descendant"),
			Match(`\*`, "operator.wildcard"),
			MatchGroups(`(\.)([A-Za-z_][A-Za-z0-9_]*)`, "delimiter", "variable.field"),
			Match(`\.`, "delimiter"),

			// A filter opens with `?(`; the union/slice forms are plain brackets.
			Match(`\?\(`, "keyword.filter"),
			Match(`[\[\]()]`, "@brackets"),

			Match(`"(?:[^"\\]|\\.)*"`, "string"),
			Match(`'(?:[^'\\]|\\.)*'`, "string"),
			Match(`\b(?:true|false|null)\b`, "keyword.constant"),
			Match(`-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?`, "number"),

			Match(`(?:==|!=|<=|>=|&&|\|\||=~|[<>!+\-*/])`, "operator"),
			Match(`:`, "operator.slice"),
			Match(`,`, "delimiter"),
			Match(`[A-Za-z_][A-Za-z0-9_]*`, "variable.field"),
		),
	}

	config := Configuration{
		Brackets: [][2]string{{"[", "]"}, {"(", ")"}},
		AutoClosingPairs: []Pair{
			{Open: "[", Close: "]"}, {Open: "(", Close: ")"},
			{Open: `"`, Close: `"`}, {Open: "'", Close: "'"},
		},
		WordPattern: `[A-Za-z_][A-Za-z0-9_]*`,
	}
	return lang, config
}
