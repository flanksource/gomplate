package genmonarch

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"

	gomplate "github.com/flanksource/gomplate/v3"
	"github.com/flanksource/gomplate/v3/genmonarch/grammar"
)

// ExtractGoTemplate reads the go-template surface from the FuncMap gomplate
// actually installs, plus the vocabulary of text/template itself.
//
// Namespaces are registered as a zero-argument function returning a shared
// `*XFuncs` value (`f["strings"] = func() any { return ns }`), so the namespaced
// names are the exported methods of whatever that call returns.
func ExtractGoTemplate() (GoTemplateSpec, error) {
	g, err := grammar.ParseGoTemplate()
	if err != nil {
		return GoTemplateSpec{}, err
	}

	spec := GoTemplateSpec{
		Keywords: g.Keywords,
		Builtins: g.Builtins,
		Delimiters: Delimiters{
			Left:         g.LeftDelim,
			Right:        g.RightDelim,
			LeftComment:  g.LeftComment,
			RightComment: g.RightComment,
			TrimMarker:   g.TrimMarker,
		},
	}

	namespaces := map[string]bool{}
	ordinary := map[string]any{}
	for name, entry := range gomplate.CreateFuncs(context.Background()) {
		ns, isNamespace := namespaceValue(entry)
		if !isNamespace {
			ordinary[name] = entry
			continue
		}
		namespaces[name] = true
		methods, err := methodsOf(ns, name)
		if err != nil {
			return GoTemplateSpec{}, err
		}
		spec.Functions = append(spec.Functions, methods...)
	}
	for name, entry := range ordinary {
		sig, err := functionSignature(entry, 0)
		if err != nil {
			return GoTemplateSpec{}, fmt.Errorf("reading signature for %s: %w", name, err)
		}
		spec.Functions = append(spec.Functions, Function{
			Name:      name,
			Signature: sig,
		})
	}

	for ns := range namespaces {
		spec.Namespaces = append(spec.Namespaces, ns)
	}
	sort.Strings(spec.Namespaces)
	sort.Slice(spec.Functions, func(i, j int) bool { return spec.Functions[i].Name < spec.Functions[j].Name })

	if len(spec.Functions) == 0 {
		return GoTemplateSpec{}, fmt.Errorf("gomplate.CreateFuncs returned no functions")
	}
	return spec, nil
}

// namespaceValue calls a `func() any` namespace accessor and returns the value
// it yields. Reports false for anything that is an ordinary template function.
func namespaceValue(entry any) (reflect.Value, bool) {
	v := reflect.ValueOf(entry)
	t := v.Type()
	if t.Kind() != reflect.Func || t.NumIn() != 0 || t.NumOut() != 1 || t.IsVariadic() {
		return reflect.Value{}, false
	}
	out := v.Call(nil)[0]
	// The accessor is declared as `func() any`, so unwrap the interface to
	// reach the concrete *XFuncs before looking for methods.
	for out.Kind() == reflect.Interface {
		out = out.Elem()
	}
	if !out.IsValid() || out.NumMethod() == 0 {
		return reflect.Value{}, false
	}
	return out, true
}

// methodsOf lists the exported methods of a namespace value as `ns.Method`.
func methodsOf(ns reflect.Value, namespace string) ([]Function, error) {
	t := ns.Type()
	out := make([]Function, 0, t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if !m.IsExported() {
			continue
		}
		sig, err := methodSignature(t, m)
		if err != nil {
			return nil, fmt.Errorf("reading signature for %s.%s: %w", namespace, m.Name, err)
		}
		out = append(out, Function{
			Name:      namespace + "." + m.Name,
			Namespace: namespace,
			// Method values carry the receiver in Type.In(0); the signature an
			// author writes does not.
			Signature: sig,
		})
	}
	return out, nil
}

// functionSignature renders a Go function the way a template author sees it,
// retaining source parameter names that reflection itself discards.
func functionSignature(fn any, skip int) (string, error) {
	t := reflect.TypeOf(fn)
	if t == nil || t.Kind() != reflect.Func {
		return "", fmt.Errorf("expected a function, got %T", fn)
	}
	names, err := sourceParameterNames(fn, t.NumIn()-skip)
	if err != nil {
		return "", err
	}
	return renderSignature(t, names, skip), nil
}

func methodSignature(receiver reflect.Type, method reflect.Method) (string, error) {
	names, err := methodParameterNames(receiver, method.Name, method.Type.NumIn()-1)
	if err != nil {
		return "", err
	}
	base := receiver
	for base.Kind() == reflect.Pointer {
		base = base.Elem()
	}
	sourceParameterCache.Store(base.PkgPath()+"."+base.Name()+"."+method.Name, names)
	sourceParameterCache.Store(base.PkgPath()+".(*"+base.Name()+")."+method.Name, names)
	if runtimeFn := runtime.FuncForPC(method.Func.Pointer()); runtimeFn != nil {
		sourceParameterCache.Store(strings.TrimSuffix(runtimeFn.Name(), "-fm"), names)
	}
	return renderSignature(method.Type, names, 1), nil
}

func renderSignature(t reflect.Type, names []string, skip int) string {
	var args []string
	for i := skip; i < t.NumIn(); i++ {
		arg := t.In(i).String()
		if t.IsVariadic() && i == t.NumIn()-1 {
			arg = "..." + strings.TrimPrefix(arg, "[]")
		}
		args = append(args, names[i-skip]+" "+arg)
	}

	var results []string
	for i := 0; i < t.NumOut(); i++ {
		results = append(results, t.Out(i).String())
	}

	sig := "(" + strings.Join(args, ", ") + ")"
	switch len(results) {
	case 0:
	case 1:
		sig += " " + results[0]
	default:
		sig += " (" + strings.Join(results, ", ") + ")"
	}
	return sig
}

var (
	parsedSourceFiles    sync.Map
	sourceParameterCache sync.Map
)

type parsedSourceFile struct {
	file *ast.File
	fset *token.FileSet
}

func sourceParameterNames(fn any, count int) ([]string, error) {
	v := reflect.ValueOf(fn)
	pc := v.Pointer()
	runtimeFn := runtime.FuncForPC(pc)
	if runtimeFn == nil {
		return nil, fmt.Errorf("runtime has no function metadata")
	}
	cacheKey := strings.TrimSuffix(runtimeFn.Name(), "-fm")
	if cached, ok := sourceParameterCache.Load(cacheKey); ok {
		names := cached.([]string)
		if len(names) != count {
			return nil, fmt.Errorf("source metadata for %s has %d parameters, expected %d", cacheKey, len(names), count)
		}
		return append([]string(nil), names...), nil
	}
	filename, line := runtimeFn.FileLine(pc)
	if filename == "<autogenerated>" {
		packagePath, receiver, method, ok := runtimeMethodIdentity(cacheKey)
		if !ok {
			return nil, fmt.Errorf("source metadata for %s was not indexed", cacheKey)
		}
		names, err := namedMethodParameterNames(packagePath, receiver, method, count)
		if err != nil {
			return nil, err
		}
		sourceParameterCache.Store(cacheKey, names)
		return append([]string(nil), names...), nil
	}
	source, err := parseSourceFile(filename)
	if err != nil {
		return nil, err
	}

	name := sourceFunctionName(runtimeFn.Name())
	var matchingName, containingLine *ast.FuncDecl
	for _, decl := range source.file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if function.Name.Name == name && fieldCount(function.Type.Params) == count {
			matchingName = function
		}
		if source.fset.Position(function.Pos()).Line <= line && line <= source.fset.Position(function.End()).Line && fieldCount(function.Type.Params) == count {
			containingLine = function
		}
	}
	selected := containingLine
	if selected == nil {
		selected = matchingName
	}
	if selected == nil {
		return nil, fmt.Errorf("could not locate %s with %d parameters in %s:%d", name, count, filename, line)
	}
	names := fieldNames(selected.Type.Params)
	sourceParameterCache.Store(cacheKey, names)
	return append([]string(nil), names...), nil
}

func methodParameterNames(receiver reflect.Type, name string, count int) ([]string, error) {
	for receiver.Kind() == reflect.Pointer {
		receiver = receiver.Elem()
	}
	return namedMethodParameterNames(receiver.PkgPath(), receiver.Name(), name, count)
}

func namedMethodParameterNames(packagePath, receiver, name string, count int) ([]string, error) {
	dir, err := packageSourceDir(packagePath)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		source, err := parseSourceFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		for _, decl := range source.file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Recv == nil || function.Name.Name != name {
				continue
			}
			if receiverName(function.Recv.List[0].Type) != receiver || fieldCount(function.Type.Params) != count {
				continue
			}
			return fieldNames(function.Type.Params), nil
		}
	}
	return nil, fmt.Errorf("could not locate %s.%s with %d parameters", receiver, name, count)
}

func runtimeMethodIdentity(runtimeName string) (packagePath, receiver, method string, ok bool) {
	methodSeparator := strings.LastIndex(runtimeName, ".")
	if methodSeparator < 0 {
		return "", "", "", false
	}
	method = runtimeName[methodSeparator+1:]
	prefix := runtimeName[:methodSeparator]
	receiverSeparator := strings.LastIndex(prefix, ".")
	lastSlash := strings.LastIndex(prefix, "/")
	if receiverSeparator <= lastSlash {
		return "", "", "", false
	}
	packagePath = prefix[:receiverSeparator]
	receiver = strings.Trim(prefix[receiverSeparator+1:], "()")
	receiver = strings.TrimPrefix(receiver, "*")
	return packagePath, receiver, method, packagePath != "" && receiver != "" && method != ""
}

func packageSourceDir(packagePath string) (string, error) {
	const modulePath = "github.com/flanksource/gomplate/v3"
	if packagePath != modulePath && !strings.HasPrefix(packagePath, modulePath+"/") {
		return "", fmt.Errorf("package %s is outside %s", packagePath, modulePath)
	}
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime has no source location for genmonarch")
	}
	root := filepath.Dir(filepath.Dir(sourceFile))
	relative := strings.TrimPrefix(packagePath, modulePath)
	return filepath.Join(root, strings.TrimPrefix(relative, "/")), nil
}

func receiverName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return receiverName(typed.X)
	case *ast.IndexExpr:
		return receiverName(typed.X)
	case *ast.IndexListExpr:
		return receiverName(typed.X)
	default:
		return ""
	}
}

func parseSourceFile(filename string) (*parsedSourceFile, error) {
	if cached, ok := parsedSourceFiles.Load(filename); ok {
		return cached.(*parsedSourceFile), nil
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}
	source := &parsedSourceFile{file: file, fset: fset}
	parsedSourceFiles.Store(filename, source)
	return source, nil
}

func sourceFunctionName(runtimeName string) string {
	runtimeName = strings.TrimSuffix(runtimeName, "-fm")
	name := runtimeName[strings.LastIndex(runtimeName, ".")+1:]
	if generic := strings.IndexByte(name, '['); generic >= 0 {
		name = name[:generic]
	}
	return name
}

func fieldCount(fields *ast.FieldList) int {
	if fields == nil {
		return 0
	}
	count := 0
	for _, field := range fields.List {
		count += len(field.Names)
		if len(field.Names) == 0 {
			count++
		}
	}
	return count
}

func fieldNames(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var names []string
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			names = append(names, fmt.Sprintf("arg%d", len(names)+1))
			continue
		}
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}
