package gencel

import (
	"fmt"
	"strings"
)

func getArgs(args []Ident) string {
	var output []string
	for i := range args {
		var a string
		switch args[i].GoType {
		case "interface{}":
			a = fmt.Sprintf("args[%d]", i)
		default:
			a = fmt.Sprintf("args[%d].Value().(%s)", i, args[i].GoType)
		}

		output = append(output, a)
	}

	return strings.Join(output, ", ")
}

var tplFuncs = map[string]any{
	"fnSuffix": func(args []Ident) string {
		var output []string
		for _, a := range args {
			output = append(output, a.GoType)
		}

		return strings.Join(output, "_")
	},
	"getArgs": getArgs,
	"getReturnTypes": func(args []Ident) string {
		switch len(args) {
		case 0:
			return "nil"
		case 1:
			return goTypeToIdent(args[0].GoType).Type
		default:
			return "cel.DynType"
		}
	},
}

type funcDefTemplateView struct {
	ParentFileName string
	FnName         string
	Args           []Ident
	ReturnTypes    []Ident
	RecvType       string
}

const funcDefTemplate = `
var {{.FnName}}{{.ParentFileName}}Gen = cel.Function("{{.FnName}}",
	cel.Overload("{{.FnName}}_{{fnSuffix .Args}}",
	{{if .Args}}
	[]*cel.Type{
		{{ range $elem := .Args }} {{.Type}},	{{end}}
	}{{else}}nil{{end}},
	{{getReturnTypes .ReturnTypes}},
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			{{if gt (len .ReturnTypes) 1}}
				// Need to figure this out
				name := "Flanksource"
				return types.DefaultTypeAdapter.NativeToValue([]string{name, name + "suffix"})
			{{else}}
				var x {{.RecvType}}
				{{if eq (index .ReturnTypes 0).Type "cel.DurationType"}}
				return types.Duration{Duration: x.{{.FnName}}({{getArgs .Args}})}
				{{else if eq (index .ReturnTypes 0).Type "cel.TimestampType"}}
				return types.Timestamp{Time: x.{{.FnName}}({{getArgs .Args}})}
				{{else}}
				return types.DefaultTypeAdapter.NativeToValue(x.{{.FnName}}({{getArgs .Args}}))
				{{end}}
			{{end}}
		}),
	),
)
`
