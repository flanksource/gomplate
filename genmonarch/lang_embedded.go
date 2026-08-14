package genmonarch

import "fmt"

// Host is a language that gomplate templates are embedded in. Real Mission
// Control configuration is YAML with `{{ }}` actions inside it, and neither
// half is readable when the editor only understands the other.
type Host string

const (
	HostYAML Host = "yaml"
	HostJSON Host = "json"
	HostText Host = "text"
)

// EmbeddedLanguageID is the Monaco language id for a host with templates in it.
func EmbeddedLanguageID(h Host) string { return string(h) + "-gomplate" }

// BuildEmbedded assembles a host language whose every state can open a template
// action.
//
// Monarch cannot delegate to another *registered* language mid-state, so the
// host tokenizer is inlined here rather than composed at runtime. It is
// deliberately light: enough structure to read a config file, with the template
// actions -- the part gomplate owns -- fully tokenized by the shared states.
func BuildEmbedded(h Host, spec GoTemplateSpec, d Delimiters) (Language, Configuration, error) {
	states, entry, err := actionStates(spec, d)
	if err != nil {
		return Language{}, Configuration{}, err
	}

	hostRules, hostStates, err := hostTokenizer(h)
	if err != nil {
		return Language{}, Configuration{}, err
	}

	// The action entry rules come first so `{{` wins over any host rule that
	// would otherwise swallow it as ordinary text.
	root := append([]Rule{directiveRule()}, entry...)
	root = append(root, hostRules...)

	lang := Language{
		ID:           EmbeddedLanguageID(h),
		DefaultToken: "",
		TokenPostfix: "." + EmbeddedLanguageID(h),
		Brackets: []Bracket{
			{Open: "{", Close: "}", Token: "delimiter.curly"},
			{Open: "[", Close: "]", Token: "delimiter.square"},
			{Open: "(", Close: ")", Token: "delimiter.parenthesis"},
		},
		Attributes: goTemplateAttributes(spec),
		Tokenizer:  NewStates().Add("root", root...),
	}
	for _, name := range hostStates.Names() {
		// A host state must also be able to open an action: a template can
		// appear inside a quoted YAML scalar just as easily as at top level.
		lang.Tokenizer.Add(name, append(append([]Rule{}, entry...), hostStates.Rules(name)...)...)
	}
	for _, name := range states.Names() {
		lang.Tokenizer.Add(name, states.Rules(name)...)
	}

	config := Configuration{
		Comments: hostComments(h),
		Brackets: [][2]string{{"{", "}"}, {"[", "]"}},
		AutoClosingPairs: []Pair{
			{Open: "{", Close: "}"}, {Open: "[", Close: "]"},
			{Open: `"`, Close: `"`}, {Open: "'", Close: "'"},
		},
		WordPattern: `[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*`,
	}
	return lang, config, nil
}

func hostComments(h Host) *Comments {
	switch h {
	case HostYAML:
		return &Comments{LineComment: "#"}
	default:
		return nil
	}
}

// hostTokenizer returns the host's root rules and any extra states it needs.
//
// Two Monarch constraints shape these rules:
//
//   - Every capture group in a rule must participate in the match. An optional
//     group is `undefined` when it does not, and Monarch throws while summing
//     the group lengths. So alternatives get their own rules instead.
//   - A quoted scalar has to be a state, not a single pattern, or a template
//     inside it (`name: "{{ .app }}-web"`, the common case in real configs) is
//     swallowed whole as a string.
func hostTokenizer(h Host) ([]Rule, *States, error) {
	states := NewStates()

	switch h {
	case HostYAML:
		states.Add("yamlDouble", quotedScalar(`"`)...)
		states.Add("yamlSingle", quotedScalar(`'`)...)
		return []Rule{
			Match(`#.*$`, "comment"),
			Match(`^---\s*$`, "keyword.directive"),
			Match(`^\.\.\.\s*$`, "keyword.directive"),
			// A mapping key, with and without a leading list marker.
			MatchGroups(`^(\s*)(-\s+)([^-\s#"'][^:#]*?)(\s*)(:)(?=\s|$)`,
				"white", "delimiter.list", "type.yaml", "white", "delimiter"),
			MatchGroups(`^(\s*)([^-\s#"'][^:#]*?)(\s*)(:)(?=\s|$)`,
				"white", "type.yaml", "white", "delimiter"),
			Match(`^\s*-\s`, "delimiter.list"),
			Match(`[&*][A-Za-z0-9_-]+`, "variable.anchor"),
			Match(`!!?[A-Za-z0-9_/-]*`, "type"),
			Match(`[|>][-+]?`, "keyword.scalar"),
			Push(`"`, "string", "@yamlDouble"),
			Push(`'`, "string", "@yamlSingle"),
			Match(`\b(?:true|false|null|~|yes|no|on|off)\b`, "keyword.constant"),
			Match(`[+-]?(?:0[xX][0-9a-fA-F]+|(?:\d+\.\d*|\.\d+|\d+)(?:[eE][+-]?\d+)?)\b`, "number"),
			Match(`[{}\[\]]`, "@brackets"),
			Match(`,`, "delimiter"),
			Match(`[\s\S]`, ""),
		}, states, nil

	case HostJSON:
		states.Add("jsonString", quotedScalar(`"`)...)
		return []Rule{
			// A key is a string immediately followed by a colon. Templates can
			// appear in values, so only the key form is matched whole here.
			MatchGroups(`("(?:[^"\\{]|\\.)*")(\s*)(:)`, "type.json", "white", "delimiter"),
			Push(`"`, "string", "@jsonString"),
			Match(`\b(?:true|false|null)\b`, "keyword.constant"),
			Match(`-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?`, "number"),
			Match(`[{}\[\]]`, "@brackets"),
			Match(`[,:]`, "delimiter"),
			Match(`[\s\S]`, ""),
		}, states, nil

	case HostText:
		return []Rule{Match(`[\s\S]`, "source")}, states, nil

	default:
		return nil, nil, fmt.Errorf("unknown host language %q", h)
	}
}

// quotedScalar is the body of a quoted string state. The caller prepends the
// template entry rules, so `{{` breaks out of the string and back in again.
//
// The run rule deliberately stops at `{` so the entry rules get a chance at
// `{{`; a lone brace then falls through to the catch-all.
func quotedScalar(quote string) []Rule {
	inClass := quote
	if quote == `\` || quote == `]` || quote == `^` || quote == `-` {
		inClass = `\` + quote
	}
	return []Rule{
		Pop(escapeRegexLiteral(quote), "string"),
		Match(`\\.`, "string.escape"),
		Match(`[^`+inClass+`\\{]+`, "string"),
		Match(`[\s\S]`, "string"),
	}
}
