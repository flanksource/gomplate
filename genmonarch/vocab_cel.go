package genmonarch

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common"

	gomplate "github.com/flanksource/gomplate/v3"
)

// celBuiltinTypes are the type names CEL exposes as identifiers. They come from
// the checker's standard declarations rather than from a rule in CEL.g4, which
// only knows IDENTIFIER.
var celBuiltinTypes = []string{
	"bool", "bytes", "double", "dyn", "duration", "int", "list", "map",
	"null_type", "string", "timestamp", "type", "uint",
}

// ExtractCEL reads the CEL surface out of a live environment built from the
// same options RunExpression compiles against, so the catalogue is whatever
// gomplate actually registers -- not a list maintained alongside it.
//
// extra layers a caller's own options on top. Because this reads a live
// cel.Env rather than a maintained list, a host that registers
// `cel.Function("catalog.query", cel.Overload(...))` gets it back here with its
// typed overloads, and the editor can highlight and complete it without any
// change to the grammar.
func ExtractCEL(extra ...cel.EnvOption) (CELSpec, error) {
	opts := gomplate.GetCelEnv(nil)
	opts = append(opts, extra...)
	env, err := cel.NewEnv(opts...)
	if err != nil {
		return CELSpec{}, fmt.Errorf("building the CEL environment: %w", err)
	}

	spec := CELSpec{
		Keywords: gomplate.CELKeywords(),
		Types:    celBuiltinTypes,
	}

	namespaces := map[string]bool{}
	for name, decl := range env.Functions() {
		if !isAuthorable(name) {
			continue
		}
		fn := Function{Name: name, Doc: decl.Description()}
		if ns, _, ok := splitNamespace(name); ok {
			fn.Namespace = ns
			namespaces[ns] = true
		}

		memberOnly := true
		for _, o := range decl.OverloadDecls() {
			args := make([]string, 0, len(o.ArgTypes()))
			for _, a := range o.ArgTypes() {
				args = append(args, a.String())
			}
			fn.Overloads = append(fn.Overloads, Overload{
				ID:     o.ID(),
				Args:   args,
				Result: o.ResultType().String(),
				Member: o.IsMemberFunction(),
			})
			fn.Examples = append(fn.Examples, o.Examples()...)
			if !o.IsMemberFunction() {
				memberOnly = false
			}
		}
		if len(fn.Overloads) == 0 {
			continue // a declaration with no overload is not callable
		}
		fn.MemberOnly = memberOnly
		sort.Slice(fn.Overloads, func(i, j int) bool { return fn.Overloads[i].ID < fn.Overloads[j].ID })
		spec.Functions = append(spec.Functions, fn)
	}

	for _, m := range env.Macros() {
		macro := Macro{
			Name:          m.Function(),
			ArgCount:      m.ArgCount(),
			ReceiverStyle: m.IsReceiverStyle(),
		}
		// Docs are attached with cel.MacroDocs/MacroExamples -- gomplate's own
		// `fold` macro sets both (cel_fold.go) -- but the Macro interface does
		// not require them, so probe rather than assume.
		if documented, ok := m.(interface{ Documentation() *common.Doc }); ok {
			if doc := documented.Documentation(); doc != nil {
				macro.Doc = doc.Description
				for _, child := range doc.Children {
					if child.Kind == common.DocExample {
						macro.Examples = append(macro.Examples, child.Description)
					}
				}
			}
		}
		spec.Macros = append(spec.Macros, macro)
	}

	for _, v := range env.Variables() {
		spec.Variables = append(spec.Variables, v.Name())
	}

	for ns := range namespaces {
		spec.Namespaces = append(spec.Namespaces, ns)
	}

	sortSpec(&spec)
	return spec, nil
}

// isAuthorable reports whether a declared name is one a person can actually
// type. The checker also declares operators (`_+_`, `_?._`, `_[?_]`), internal
// helpers (`__not_strictly_false__`) and macro-expansion targets (`cel.@fold*`)
// as functions; none of them belong in a tokenizer or a completion list.
func isAuthorable(name string) bool {
	for _, segment := range strings.Split(name, ".") {
		if segment == "" {
			return false
		}
		for i, r := range segment {
			isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
			isDigit := r >= '0' && r <= '9'
			if i == 0 && !isLetter {
				return false
			}
			if !isLetter && !isDigit && r != '_' {
				return false
			}
		}
	}
	return name != ""
}

// splitNamespace splits `k8s.isHealthy` into its namespace and leaf. Reports
// false for un-namespaced names such as `toJSON`.
func splitNamespace(name string) (ns, leaf string, ok bool) {
	i := strings.Index(name, ".")
	if i <= 0 || i == len(name)-1 {
		return "", name, false
	}
	return name[:i], name[i+1:], true
}

func sortSpec(spec *CELSpec) {
	sort.Strings(spec.Namespaces)
	sort.Strings(spec.Keywords)
	sort.Strings(spec.Variables)
	sort.Slice(spec.Functions, func(i, j int) bool { return spec.Functions[i].Name < spec.Functions[j].Name })
	sort.Slice(spec.Macros, func(i, j int) bool {
		if spec.Macros[i].Name != spec.Macros[j].Name {
			return spec.Macros[i].Name < spec.Macros[j].Name
		}
		return spec.Macros[i].ArgCount < spec.Macros[j].ArgCount
	})
}
