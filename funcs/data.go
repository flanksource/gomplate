package funcs

import (
	"context"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/gomplate/v3/data"
)

// DataNS -
// Deprecated: don't use
func DataNS() *DataFuncs {
	return &DataFuncs{}
}

// CreateDataFuncs -
func CreateDataFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}

	ns := &DataFuncs{ctx}

	f["data"] = func() interface{} { return ns }

	f["json"] = ns.JSON
	f["jsonArray"] = ns.JSONArray
	f["yaml"] = ns.YAML
	f["yamlArray"] = ns.YAMLArray
	f["toml"] = ns.TOML
	f["csv"] = ns.CSV
	f["csvByRow"] = ns.CSVByRow
	f["csvByColumn"] = ns.CSVByColumn
	f["toJSON"] = ns.ToJSON
	f["toJSONPretty"] = ns.ToJSONPretty
	f["toYAML"] = ns.ToYAML
	f["toTOML"] = ns.ToTOML
	f["toCSV"] = ns.ToCSV
	return f
}

// DataFuncs -
type DataFuncs struct {
	ctx context.Context
}

// JSON -
func (f *DataFuncs) JSON(in interface{}) (map[string]interface{}, error) {
	return data.JSON(conv.ToString(in))
}

// JSONArray -
func (f *DataFuncs) JSONArray(in interface{}) ([]interface{}, error) {
	return data.JSONArray(conv.ToString(in))
}

// YAML -
func (f *DataFuncs) YAML(in interface{}) (map[string]interface{}, error) {
	return data.YAML(conv.ToString(in))
}

// YAMLArray -
func (f *DataFuncs) YAMLArray(in interface{}) ([]interface{}, error) {
	return data.YAMLArray(conv.ToString(in))
}

// TOML -
func (f *DataFuncs) TOML(in interface{}) (interface{}, error) {
	return data.TOML(conv.ToString(in))
}

// CSV -
func (f *DataFuncs) CSV(args ...string) ([][]string, error) {
	return data.CSV(args...)
}

// CSVByRow -
func (f *DataFuncs) CSVByRow(args ...string) (rows []map[string]string, err error) {
	return data.CSVByRow(args...)
}

// CSVByColumn -
func (f *DataFuncs) CSVByColumn(args ...string) (cols map[string][]string, err error) {
	return data.CSVByColumn(args...)
}

// ToCSV -
func (f *DataFuncs) ToCSV(args ...interface{}) (string, error) {
	return data.ToCSV(args...)
}

// ToJSON -
func (f *DataFuncs) ToJSON(in interface{}) (string, error) {
	return data.ToJSON(in)
}

// ToJSONPretty -
func (f *DataFuncs) ToJSONPretty(indent string, in interface{}) (string, error) {
	return data.ToJSONPretty(indent, in)
}

// ToYAML -
func (f *DataFuncs) ToYAML(in interface{}) (string, error) {
	return data.ToYAML(in)
}

// ToTOML -
func (f *DataFuncs) ToTOML(in interface{}) (string, error) {
	return data.ToTOML(in)
}
