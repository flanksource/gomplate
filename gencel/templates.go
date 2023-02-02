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
	"getReturnIdentifiers": func(args []Ident) string {
		var output []string
		for i := range args {
			output = append(output, fmt.Sprintf("a%d", i))
		}

		return strings.Join(output, ", ")
	},
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
	// IdentName is the name of the exported cel func
	// in this codebase.
	IdentName string

	// FnName is the name of the cel func inside the
	// cel environment.
	FnName string

	// Args is the list of arguments of the go func
	// that this cel func is encapsulating.
	Args []Ident

	// ReturnTypes is the list of all returns of the go func
	// that this cel func is encapsulating.
	ReturnTypes []Ident

	// RecvType is the parent type of the member func
	// that this cel func is encapsulating.
	RecvType string
}

const funcBodyTemplate = `
{{define "body"}}
		{{if gt (len .ReturnTypes) 1}}
			var x {{.RecvType}}
			{{getReturnIdentifiers .ReturnTypes}} := x.{{.FnName}}({{getArgs .Args}})
			return types.DefaultTypeAdapter.NativeToValue([]any{
				{{getReturnIdentifiers .ReturnTypes}},
			})
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
{{end}}
`

const funcDefTemplate = `
var {{.IdentName}} = cel.Function("{{.FnName}}",
	cel.Overload("{{.FnName}}_{{fnSuffix .Args}}",
	{{if .Args}}
	[]*cel.Type{
		{{ range $elem := .Args }} {{.Type}},	{{end}}
	}{{else}}nil{{end}},
	{{getReturnTypes .ReturnTypes}},
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			{{ block "body" . }}{{end}}
		}),
	),
)
`

type exportFuncsTemplateView struct {
	FnNames []string
}

const exportAllTemplate = `
var CelEnvOption = []cel.EnvOption{
	{{range $fnName := .FnNames}}{{$fnName}},
	{{end}}
}
`
