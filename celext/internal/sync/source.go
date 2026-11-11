package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

func copySource(path, header string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	positions := token.NewFileSet()
	file, err := parser.ParseFile(positions, path, data, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	if file.Name.Name != "ext" {
		return nil, fmt.Errorf("%s: expected package ext, got %s", path, file.Name.Name)
	}
	if err := checkInitializers(path, file); err != nil {
		return nil, err
	}
	// Change only the parsed package identifier, retaining the upstream license and source bytes.
	start, end := positions.Position(file.Name.Pos()).Offset, positions.Position(file.Name.End()).Offset
	result := append([]byte(header), data[:start]...)
	result = append(result, "celext"...)
	return append(result, data[end:]...), nil
}

func checkInitializers(path string, file *ast.File) error {
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			if declaration.Recv == nil && declaration.Name.Name == "init" {
				return fmt.Errorf("%s: init function requires review for global name registration; refusing to copy", path)
			}
		case *ast.GenDecl:
			if declaration.Tok == token.VAR {
				if err := checkVariableCalls(path, declaration); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkVariableCalls(path string, declaration *ast.GenDecl) error {
	var failure error
	ast.Inspect(declaration, func(node ast.Node) bool {
		if failure != nil {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		var name bytes.Buffer
		if err := format.Node(&name, token.NewFileSet(), call.Fun); err != nil {
			failure = fmt.Errorf("%s: inspect package initializer: %w", path, err)
			return false
		}
		switch name.String() {
		case "types.NewErr", "reflect.TypeFor[[]string]":
			return true
		default:
			failure = fmt.Errorf("%s: package initializer calls %s; review for global name registration before copying", path, name.String())
			return false
		}
	})
	return failure
}
