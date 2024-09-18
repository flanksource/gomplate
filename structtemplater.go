package gomplate

import (
	"reflect"
	"strings"

	"github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/mitchellh/reflectwalk"
	"gopkg.in/yaml.v3"
)

type StructTemplater struct {
	Context context.Context
	Values  map[string]interface{}
	// IgnoreFields from walking where key is field name and value is field type
	IgnoreFields map[string]string
	Funcs        map[string]any
	DelimSets    []Delims
	// If specified create a function for each value so that is can be accessed via {{ value }} in addition to {{ .value }}
	ValueFunctions bool
	RequiredTag    string
}

type Delims struct {
	Left, Right string
}

// this func is required to fulfil the reflectwalk.StructWalker interface
func (w StructTemplater) Struct(reflect.Value) error {
	return nil
}

func (w StructTemplater) StructField(f reflect.StructField, v reflect.Value) error {
	if !v.CanSet() {
		return nil
	}

	for key, value := range w.IgnoreFields {
		if key == f.Name && value == f.Type.String() {
			return reflectwalk.SkipEntry
		}
	}

	if w.RequiredTag != "" && f.Tag.Get(w.RequiredTag) != "true" {
		return reflectwalk.SkipEntry
	}

	switch v.Kind() {
	case reflect.String:
		val, err := w.Template(v.String())
		if err != nil {
			return err
		}
		v.SetString(val)

	case reflect.Map:
		if len(v.MapKeys()) != 0 {
			newMap := reflect.MakeMap(v.Type())
			for _, key := range v.MapKeys() {
				val := v.MapIndex(key)
				newKey, err := w.templateKey(key)
				if err != nil {
					return err
				}

				concreteVal := reflect.ValueOf(val.Interface())
				switch concreteVal.Kind() {
				case reflect.String:
					newVal, err := w.Template(concreteVal.String())
					if err != nil {
						return err
					}
					newMap.SetMapIndex(newKey, reflect.ValueOf(newVal))

				case reflect.Map:
					marshalled, err := yaml.Marshal(val.Interface())
					if err != nil {
						newMap.SetMapIndex(newKey, val)
					} else {
						templated, err := w.Template(string(marshalled))
						if err != nil {
							return err
						}

						var unmarshalled map[string]any
						if err := yaml.Unmarshal([]byte(templated), &unmarshalled); err != nil {
							newMap.SetMapIndex(newKey, val)
						} else {
							newMap.SetMapIndex(newKey, reflect.ValueOf(unmarshalled))
						}
					}

				default:
					newMap.SetMapIndex(newKey, val)
				}
			}
			v.Set(newMap)
		}
	}

	return nil
}

func (w StructTemplater) templateKey(v reflect.Value) (reflect.Value, error) {
	if v.Kind() == reflect.String {
		key, err := w.Template(v.String())
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(key), nil
	}
	return v, nil
}

func (w StructTemplater) Walk(object interface{}) error {
	if w.Context.Logger == nil {
		w.Context = newContext()
	}
	w.Context.Logger.V(7).Infof("walking %s", logger.Pretty(object))
	return reflectwalk.Walk(object, w)
}

func (w StructTemplater) Template(val string) (string, error) {
	w.Context.Logger.V(8).Infof("templating %s", val)
	in := val
	if strings.TrimSpace(val) == "" {
		return val, nil
	}
	if w.Funcs == nil {
		w.Funcs = make(map[string]any)
	}
	if w.ValueFunctions {
		for k, v := range w.Values {
			_v := v
			w.Funcs[k] = func() interface{} {
				return _v
			}
		}
	}

	delimSets := w.DelimSets

	// parse go-template headers and override the delim set, as otherwise the header get stripped out
	// for the first set of delims and reverts back to the default for the subsequent delimeters
	template, err := parseAndStripTemplateHeader(Template{Template: val})
	if err != nil {
		return "", err
	}

	val = template.Template
	if template.LeftDelim != "" && template.RightDelim != "" {
		delimSets = []Delims{{Left: template.LeftDelim, Right: template.RightDelim}}
	}

	if len(delimSets) == 0 {
		delimSets = []Delims{{Left: "{{", Right: "}}"}}
	}

	for _, delims := range delimSets {
		val, err = goTemplate(w.Context, Template{
			Template:   val,
			Functions:  w.Funcs,
			RightDelim: delims.Right,
			LeftDelim:  delims.Left,
		}, w.Values)

		if err != nil {
			return val, err
		}
	}
	val = strings.TrimSpace(val)
	if strings.TrimSpace(val) != strings.TrimSpace(in) {
		w.Context.Logger.V(6).Infof("==> %s", val)
	}
	return val, nil
}
