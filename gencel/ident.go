package gencel

import "go/ast"

type Ident struct {
	Type   string
	GoType string
}

// getCelArgs converts native go types to cel-go types
func getCelArgs(args []ast.Expr) []Ident {
	var celArgs = make([]Ident, len(args))
	for i, a := range args {
		celArgs[i] = astToIdent(a)
	}

	return celArgs
}

func goTypeToIdent(name string) Ident {
	switch name {
	case "string":
		return Ident{Type: "cel.StringType", GoType: name}
	case "bool":
		return Ident{Type: "cel.BoolType", GoType: name}
	case "Duration":
		return Ident{Type: "cel.DurationType", GoType: "time.Duration"}
	case "Time":
		return Ident{Type: "cel.TimestampType", GoType: "time.Time"}
	case "int", "int32", "int64":
		return Ident{Type: "cel.IntType", GoType: name}
	case "float32", "float64":
		return Ident{Type: "cel.DoubleType", GoType: name}
	default:
		return Ident{Type: "cel.DynType", GoType: name}
	}
}

func astToIdent(a ast.Expr) Ident {
	switch v := a.(type) {
	case *ast.InterfaceType:
		return Ident{Type: "cel.DynType", GoType: "interface{}"}
	case *ast.Ident:
		return goTypeToIdent(v.Name)
	case *ast.Ellipsis:
		return Ident{Type: "cel.DynType", GoType: "interface{}"}
	case *ast.SelectorExpr:
		return goTypeToIdent(v.Sel.Name)
	default:
		return Ident{Type: "cel.StringType", GoType: ""}
	}
}
