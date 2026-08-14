package grammar

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"sort"
	"strconv"

	"golang.org/x/tools/go/packages"
)

// GoTemplateGrammar is the lexical vocabulary of Go's text/template, read out
// of the standard library's own lexer and function table. All three are
// unexported there, so they are recovered from the AST -- the same approach
// gencel already uses to read gomplate's own sources.
type GoTemplateGrammar struct {
	// Keywords are the action keywords: if, range, end, ...
	Keywords []string
	// Builtins are the functions text/template supplies itself.
	Builtins []string
	// LeftDelim and RightDelim open and close an action.
	LeftDelim, RightDelim string
	// LeftComment and RightComment open and close a comment inside an action.
	LeftComment, RightComment string
	// TrimMarker abuts a delimiter to trim adjacent whitespace.
	TrimMarker string
}

// ParseGoTemplate recovers the text/template vocabulary from the standard
// library sources of the toolchain building this package.
func ParseGoTemplate() (*GoTemplateGrammar, error) {
	pkgs, err := packages.Load(
		&packages.Config{Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedName},
		"text/template", "text/template/parse",
	)
	if err != nil {
		return nil, fmt.Errorf("loading text/template sources: %w", err)
	}
	files := map[string][]*ast.File{}
	for _, p := range pkgs {
		if len(p.Errors) > 0 {
			return nil, fmt.Errorf("loading %s: %v", p.PkgPath, p.Errors[0])
		}
		files[p.PkgPath] = p.Syntax
	}

	g := &GoTemplateGrammar{}

	// parse.key maps every action keyword to its item type.
	keywords, err := mapLiteralKeys(files["text/template/parse"], "key")
	if err != nil {
		return nil, fmt.Errorf("text/template/parse: %w", err)
	}
	// `.` is in the table as itemDot; it is a token, not a keyword.
	for _, k := range keywords {
		if k != "." {
			g.Keywords = append(g.Keywords, k)
		}
	}

	// builtins() returns the FuncMap literal of everything text/template
	// provides without any user Funcs call.
	builtins, err := funcReturnedMapKeys(files["text/template"], "builtins")
	if err != nil {
		return nil, fmt.Errorf("text/template: %w", err)
	}
	g.Builtins = builtins

	consts, err := stringConsts(files["text/template/parse"],
		"leftDelim", "rightDelim", "leftComment", "rightComment")
	if err != nil {
		return nil, fmt.Errorf("text/template/parse: %w", err)
	}
	g.LeftDelim, g.RightDelim = consts["leftDelim"], consts["rightDelim"]
	g.LeftComment, g.RightComment = consts["leftComment"], consts["rightComment"]

	marker, err := runeConst(files["text/template/parse"], "trimMarker")
	if err != nil {
		return nil, fmt.Errorf("text/template/parse: %w", err)
	}
	g.TrimMarker = string(marker)

	sort.Strings(g.Keywords)
	sort.Strings(g.Builtins)
	return g, g.validate()
}

// validate fails loudly if the standard library moved something, rather than
// letting an empty vocabulary silently produce a tokenizer that highlights
// nothing.
func (g *GoTemplateGrammar) validate() error {
	switch {
	case len(g.Keywords) == 0:
		return fmt.Errorf("no action keywords found in text/template/parse")
	case len(g.Builtins) == 0:
		return fmt.Errorf("no builtin functions found in text/template")
	case g.LeftDelim == "" || g.RightDelim == "":
		return fmt.Errorf("action delimiters not found in text/template/parse")
	case g.LeftComment == "" || g.RightComment == "":
		return fmt.Errorf("comment delimiters not found in text/template/parse")
	case g.TrimMarker == "":
		return fmt.Errorf("trim marker not found in text/template/parse")
	}
	return nil
}

// mapLiteralKeys returns the string keys of a package-level map literal.
func mapLiteralKeys(files []*ast.File, name string) ([]string, error) {
	lit, err := findValueSpec(files, name)
	if err != nil {
		return nil, err
	}
	composite, ok := lit.(*ast.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("%s is not a composite literal", name)
	}
	return compositeKeys(composite, name)
}

// funcReturnedMapKeys returns the string keys of the map literal a
// zero-argument function returns directly.
func funcReturnedMapKeys(files []*ast.File, funcName string) ([]string, error) {
	for _, f := range files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != funcName || fn.Body == nil {
				continue
			}
			for _, stmt := range fn.Body.List {
				ret, ok := stmt.(*ast.ReturnStmt)
				if !ok || len(ret.Results) != 1 {
					continue
				}
				composite, ok := ret.Results[0].(*ast.CompositeLit)
				if !ok {
					continue
				}
				return compositeKeys(composite, funcName)
			}
		}
	}
	return nil, fmt.Errorf("no function %s returning a map literal", funcName)
}

func compositeKeys(composite *ast.CompositeLit, context string) ([]string, error) {
	var keys []string
	for _, elt := range composite.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		lit, ok := kv.Key.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			return nil, fmt.Errorf("%s: unquoting key %s: %w", context, lit.Value, err)
		}
		keys = append(keys, s)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("%s has no string keys", context)
	}
	return keys, nil
}

// findValueSpec locates the value assigned to a package-level identifier.
func findValueSpec(files []*ast.File, name string) (ast.Expr, error) {
	for _, f := range files {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, ident := range vs.Names {
					if ident.Name == name && i < len(vs.Values) {
						return vs.Values[i], nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("no declaration of %s", name)
}

// stringConsts reads the values of untyped string constants.
func stringConsts(files []*ast.File, names ...string) (map[string]string, error) {
	out := map[string]string{}
	for _, name := range names {
		expr, err := findValueSpec(files, name)
		if err != nil {
			return nil, err
		}
		lit, ok := expr.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return nil, fmt.Errorf("%s is not a string literal", name)
		}
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			return nil, fmt.Errorf("unquoting %s: %w", name, err)
		}
		out[name] = s
	}
	return out, nil
}

// runeConst reads the value of an untyped rune constant.
func runeConst(files []*ast.File, name string) (rune, error) {
	expr, err := findValueSpec(files, name)
	if err != nil {
		return 0, err
	}
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.CHAR {
		return 0, fmt.Errorf("%s is not a rune literal", name)
	}
	v := constant.MakeFromLiteral(lit.Value, token.CHAR, 0)
	r, ok := constant.Int64Val(constant.ToInt(v))
	if !ok {
		return 0, fmt.Errorf("%s is not a constant rune", name)
	}
	return rune(r), nil
}
