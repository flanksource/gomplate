package gomplate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	gotemplate "text/template"
)

var funcMap gotemplate.FuncMap

func init() {
	funcMap = CreateFuncs(context.Background())
}

type Template struct {
	Template   string `yaml:"template,omitempty" json:"template,omitempty"`
	JSONPath   string `yaml:"jsonPath,omitempty" json:"jsonPath,omitempty"`
	Expression string `yaml:"expr,omitempty" json:"expr,omitempty"`
	Javascript string `yaml:"javascript,omitempty" json:"javascript,omitempty"`
}

func (t Template) IsEmpty() bool {
	return t.Template == "" && t.JSONPath == "" && t.Expression == "" && t.Javascript == ""
}

func RunTemplate(environment map[string]interface{}, template Template) (string, error) {
	// javascript
	if template.Javascript != "" {
		// // FIXME: whitelist allowed files
		// vm := otto.New()
		// for k, v := range environment {
		// 	if err := vm.Set(k, v); err != nil {
		// 		return "", errors.Wrapf(err, "error setting %s", k)
		// 	}
		// }

		// if err != nil {
		// 	return "", errors.Wrapf(err, "error setting findConfigItem function")
		// }

		// out, err := vm.Run(template.Javascript)
		// if err != nil {
		// 	return "", errors.Wrapf(err, "failed to run javascript")
		// }

		// if s, err := out.ToString(); err != nil {
		// 	return "", errors.Wrapf(err, "failed to cast output to string")
		// } else {
		// 	return s, nil
		// }
	}

	// gotemplate
	if template.Template != "" {
		tpl := gotemplate.New("")
		tpl, err := tpl.Funcs(funcMap).Parse(template.Template)
		if err != nil {
			return "", err
		}

		// marshal data from interface{} to map[string]interface{}
		data, _ := json.Marshal(environment)
		unstructured := make(map[string]interface{})
		if err := json.Unmarshal(data, &unstructured); err != nil {
			return "", err
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, unstructured); err != nil {
			return "", fmt.Errorf("error executing template %s: %v", strings.Split(template.Template, "\n")[0], err)
		}
		return strings.TrimSpace(buf.String()), nil
	}

	// exprv
	if template.Expression != "" {
		// program, err := expr.Compile(template.Expression, text.MakeExpressionOptions(environment)...)
		// if err != nil {
		// 	return "", err
		// }
		// output, err := expr.Run(program, text.MakeExpressionEnvs(environment))
		// if err != nil {
		// 	return "", err
		// }
		// return fmt.Sprint(output), nil
	}
	return "", nil
}
