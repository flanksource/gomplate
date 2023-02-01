package gencel

import (
	"encoding/json"
	"go/ast"
)

type FuncDecl struct {
	Name       string     `json:"Name"`
	Args       []ast.Expr `json:"Args"`
	ReturnType ast.Expr   `json:"ReturnType"`
	Body       string     `json:"Body"`
	RecvType   string     `json:"RecvType"`
}

func (t FuncDecl) String() string {
	x, _ := json.MarshalIndent(t, "", "\t")
	return string(x)
}
