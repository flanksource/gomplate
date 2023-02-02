package gencel

import (
	"go/ast"
	"log"
	"regexp"
)

var blacklistedFuncs = []regexp.Regexp{
	*regexp.MustCompile("^Create"),
	*regexp.MustCompile("^init"),
}

type File struct {
	pkg   *Package
	path  string
	file  *ast.File
	decls []FuncDecl
}

// visitor visits all the ast nodes
// and extracts function declarations that're suitable
// for conversion.
func (t *File) visitor(n ast.Node) bool {
	switch v := n.(type) {
	case *ast.FuncDecl:
		return t.handleFuncDecl(v)
	default:
		return true
	}
}

func (t *File) handleFuncDecl(n *ast.FuncDecl) bool {
	for _, blf := range blacklistedFuncs {
		if blf.MatchString(n.Name.Name) {
			log.Printf("Ignoring func [%s]. Blacklisted pattern", n.Name.Name)
			return false
		}
	}

	if n.Type.Results == nil || len(n.Type.Results.List) == 0 {
		log.Printf("Ignoring func [%s]. Returns nothing", n.Name.Name)
		return false
	}

	decl := FuncDecl{
		Name: n.Name.Name,
	}

	if n.Type.Params != nil {
		for _, l := range n.Type.Params.List {
			if _, ok := l.Type.(*ast.Ellipsis); ok {
				log.Printf("Variadic function [%s] detected. Ignoring", n.Name.Name)
				return false
			}

			for range l.Names {
				decl.Args = append(decl.Args, l.Type)
			}
		}
	}

	for _, l := range n.Type.Results.List {
		decl.ReturnTypes = append(decl.ReturnTypes, l.Type)
	}

	if n.Recv != nil && n.Recv.List != nil {
		for _, x := range n.Recv.List {
			switch v := x.Type.(type) {
			case *ast.Ident:
				decl.RecvType = v.Name
			case *ast.StarExpr:
				switch y := v.X.(type) {
				case *ast.Ident:
					decl.RecvType = y.Name
				}
			}
		}
	}

	if decl.RecvType != "" {
		t.decls = append(t.decls, decl)
	}

	return true
}
