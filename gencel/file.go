package gencel

import (
	"go/ast"
	"strings"
)

type File struct {
	pkg   *Package
	file  *ast.File
	decls []FuncDecl
}

func (t *File) handleFuncDecl(n *ast.FuncDecl) bool {
	if strings.Index(n.Name.Name, "Create") == 0 {
		return false
	}

	decl := FuncDecl{
		Name: n.Name.Name,
	}

	if n.Type.Params != nil {
		for _, l := range n.Type.Params.List {
			for range l.Names {
				decl.Args = append(decl.Args, l.Type)
			}
		}
	}

	if n.Type.Results != nil {
		if len(n.Type.Results.List) > 1 {
			// NOTE: Multiple returns not supported by cel-go
			return false
		}

		for _, l := range n.Type.Results.List {
			decl.ReturnType = l.Type
		}
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

func (t *File) genDecl(n ast.Node) bool {
	switch v := n.(type) {
	case *ast.FuncDecl:
		return t.handleFuncDecl(v)
	default:
		return true
	}
}
