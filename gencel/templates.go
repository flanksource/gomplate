package gencel

import (
	"fmt"
	"strings"
)

var tplFuncs = map[string]any{
	"castReturn": celTypeToObj,
	"getArgs": func(args []Ident) string {
		output := []string{}
		for i := range args {
			var a string
			switch args[i].GoType {
			case "time.Time":
				a = fmt.Sprintf("args[%d].Value().(time.Time)", i)
			case "time.Duration":
				a = fmt.Sprintf("args[%d].Value().(time.Duration)", i)
			default:
				a = fmt.Sprintf("args[%d]", i)
			}

			output = append(output, a)
		}

		return strings.Join(output, ", ")
	},
}

type funcDefTemplateView struct {
	FnName     string
	Args       []Ident
	ReturnType Ident
	RecvType   string
}

const funcDefTemplate = `
var {{.FnName}}Gen = cel.Function("{{.FnName}}",
	cel.Overload("{{.FnName}}_string_string",
	[]*cel.Type{
		{{ range $elem := .Args }} {{.Type}},	{{end}}
	},
	{{.ReturnType.Type}},
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			var x {{.RecvType}}
			{{if eq .ReturnType.Type "cel.DurationType"}}
			return types.Duration{Duration: x.{{.FnName}}({{getArgs .Args}})}
			{{else if eq .ReturnType.Type "cel.TimestampType"}}
			return types.Timestamp{Time: x.{{.FnName}}({{getArgs .Args}})}
			{{else}}
			return types.{{castReturn .ReturnType.Type}}(x.{{.FnName}}({{getArgs .Args}}))
			{{end}}
		}),
	),
)
`
