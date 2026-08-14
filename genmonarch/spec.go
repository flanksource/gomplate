// Package genmonarch builds Monaco language definitions for the expression
// languages gomplate runs, from the grammars and registries gomplate itself
// uses -- so the editor cannot drift from the evaluator.
package genmonarch

// Spec is the machine-readable catalogue of everything gomplate exposes to an
// author. It drives the generated tokenizers, the completion and hover
// providers, and the playground's function browser.
type Spec struct {
	CEL        CELSpec        `json:"cel"`
	GoTemplate GoTemplateSpec `json:"gotemplate"`
}

// CELSpec is the CEL surface, read from a live cel.Env.
type CELSpec struct {
	// Namespaces are the dotted prefixes in use: k8s, math, time, ...
	Namespaces []string `json:"namespaces"`
	// Keywords are CEL's reserved words.
	Keywords []string `json:"keywords"`
	// Types are the built-in type names usable as identifiers.
	Types []string `json:"types"`
	// Variables are the identifiers the base environment declares.
	Variables []string   `json:"variables,omitempty"`
	Macros    []Macro    `json:"macros"`
	Functions []Function `json:"functions"`
}

// GoTemplateSpec is the Go text/template surface.
type GoTemplateSpec struct {
	// Namespaces are the dotted prefixes: strings, coll, conv, ...
	Namespaces []string `json:"namespaces"`
	// Keywords are text/template's action keywords.
	Keywords []string `json:"keywords"`
	// Builtins are the functions text/template provides itself.
	Builtins []string `json:"builtins"`
	// Delimiters are text/template's defaults.
	Delimiters Delimiters `json:"delimiters"`
	Functions  []Function `json:"functions"`
}

// Delimiters is an action-delimiter pair.
type Delimiters struct {
	Left  string `json:"left"`
	Right string `json:"right"`
	// LeftComment and RightComment open and close a comment *inside* an action.
	LeftComment  string `json:"leftComment"`
	RightComment string `json:"rightComment"`
	// TrimMarker abuts a delimiter to trim surrounding whitespace.
	TrimMarker string `json:"trimMarker"`
}

// Function is one callable name, with every registered overload.
type Function struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	// MemberOnly marks a function callable only as `x.f()`, never as `f(x)`.
	// The tokenizer uses this to colour `x.sum()` without colouring a bare
	// `sum`, which in CEL is just an identifier.
	MemberOnly bool   `json:"memberOnly,omitempty"`
	Doc        string `json:"doc,omitempty"`
	// Signature is the Go signature, for go-template functions.
	Signature string     `json:"signature,omitempty"`
	Overloads []Overload `json:"overloads,omitempty"`
	Examples  []string   `json:"examples,omitempty"`
}

// Overload is one typed signature of a CEL function.
type Overload struct {
	ID     string   `json:"id"`
	Args   []string `json:"args"`
	Result string   `json:"result"`
	Member bool     `json:"member,omitempty"`
}

// Macro is a CEL macro -- expanded at parse time, so it is never a Function.
type Macro struct {
	Name          string   `json:"name"`
	ArgCount      int      `json:"argCount"`
	ReceiverStyle bool     `json:"receiverStyle"`
	Doc           string   `json:"doc,omitempty"`
	Examples      []string `json:"examples,omitempty"`
}

// MemberNames returns the names callable only in member position.
func (s CELSpec) MemberNames() []string {
	var out []string
	for _, f := range s.Functions {
		if f.MemberOnly {
			out = append(out, f.Name)
		}
	}
	return out
}

// GlobalNames returns the names callable in global position, namespace included.
func (s CELSpec) GlobalNames() []string {
	var out []string
	for _, f := range s.Functions {
		if !f.MemberOnly {
			out = append(out, f.Name)
		}
	}
	return out
}
