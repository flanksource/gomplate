/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// URLs provides a CEL function library extension of URL parsing functions.
//
// url
//
// Converts a string to a URL or results in an error if the string is not a valid URL. The URL must be an absolute URI
// or an absolute path.
//
//	url(<string>) <URL>
//
// Examples:
//
//	url('https://user:pass@example.com:80/path?query=val#fragment') // returns a URL
//	url('/absolute-path') // returns a URL
//	url('https://a:b:c/') // error
//	url('../relative-path') // error
//
// isURL
//
// Returns true if a string is a valid URL. The URL must be an absolute URI or an absolute path.
//
//	isURL( <string>) <bool>
//
// Examples:
//
//	isURL('https://user:pass@example.com:80/path?query=val#fragment') // returns true
//	isURL('/absolute-path') // returns true
//	isURL('https://a:b:c/') // returns false
//	isURL('../relative-path') // returns false
//
// getScheme / getHost / getHostname / getPort / getEscapedPath / getQuery
//
// Return the parsed components of a URL.
//
//   - getScheme: If absent in the URL, returns an empty string.
//
//   - getHostname: IPv6 addresses are returned without braces, e.g. "::1". If absent in the URL, returns an empty string.
//
//   - getHost: IPv6 addresses are returned with braces, e.g. "[::1]". If absent in the URL, returns an empty string.
//
//   - getEscapedPath: The string returned by getEscapedPath is URL escaped, e.g. "with space" becomes "with%20space".
//     If absent in the URL, returns an empty string.
//
//   - getPort: If absent in the URL, returns an empty string.
//
//   - getQuery: Returns the query parameters in "matrix" form where a repeated query key is interpreted to
//     mean that there are multiple values for that key. The keys and values are returned unescaped.
//     If absent in the URL, returns an empty map.
//
//     <URL>.getScheme() <string>
//     <URL>.getHost() <string>
//     <URL>.getHostname() <string>
//     <URL>.getPort() <string>
//     <URL>.getEscapedPath() <string>
//     <URL>.getQuery() <map <string>, <list <string>>
//
// Examples:
//
//	url('/path').getScheme() // returns ''
//	url('https://example.com/').getScheme() // returns 'https'
//	url('https://example.com:80/').getHost() // returns 'example.com:80'
//	url('https://example.com/').getHost() // returns 'example.com'
//	url('https://[::1]:80/').getHost() // returns '[::1]:80'
//	url('https://[::1]/').getHost() // returns '[::1]'
//	url('/path').getHost() // returns ''
//	url('https://example.com:80/').getHostname() // returns 'example.com'
//	url('https://127.0.0.1:80/').getHostname() // returns '127.0.0.1'
//	url('https://[::1]:80/').getHostname() // returns '::1'
//	url('/path').getHostname() // returns ''
//	url('https://example.com:80/').getPort() // returns '80'
//	url('https://example.com/').getPort() // returns ''
//	url('/path').getPort() // returns ''
//	url('https://example.com/path').getEscapedPath() // returns '/path'
//	url('https://example.com/path with spaces/').getEscapedPath() // returns '/path%20with%20spaces/'
//	url('https://example.com').getEscapedPath() // returns ''
//	url('https://example.com/path?k1=a&k2=b&k2=c').getQuery() // returns { 'k1': ['a'], 'k2': ['b', 'c']}
//	url('https://example.com/path?key with spaces=value with spaces').getQuery() // returns { 'key with spaces': ['value with spaces']}
//	url('https://example.com/path?').getQuery() // returns {}
//	url('https://example.com/path').getQuery() // returns {}
func URLs() cel.EnvOption {
	return cel.Lib(urlsLib)
}

// URL provides a CEL representation of a URL.
type URL struct {
	*url.URL
}

var (
	URLObject = decls.NewObjectType("kubernetes.URL")
	typeValue = types.NewTypeValue("kubernetes.URL")
	URLType   = cel.ObjectType("kubernetes.URL")
)

// ConvertToNative implements ref.Val.ConvertToNative.
func (d URL) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	if reflect.TypeOf(d).AssignableTo(typeDesc) {
		return d, nil
	}
	if reflect.TypeOf("").AssignableTo(typeDesc) {
		return d.String(), nil
	}
	return nil, fmt.Errorf("type conversion error from 'URL' to '%v'", typeDesc)
}

// ConvertToType implements ref.Val.ConvertToType.
func (d URL) ConvertToType(typeVal ref.Type) ref.Val {
	switch typeVal {
	case typeValue:
		return d
	case types.TypeType:
		return typeValue
	}
	return types.NewErr("type conversion error from '%s' to '%s'", typeValue, typeVal)
}

// Equal implements ref.Val.Equal.
func (d URL) Equal(other ref.Val) ref.Val {
	otherDur, ok := other.(URL)
	if !ok {
		return types.MaybeNoSuchOverloadErr(other)
	}
	return types.Bool(d.String() == otherDur.String())
}

// Type implements ref.Val.Type.
func (d URL) Type() ref.Type {
	return typeValue
}

// Value implements ref.Val.Value.
func (d URL) Value() interface{} {
	return d.URL
}

var urlsLib = &urls{}

type urls struct{}

func (*urls) LibraryName() string {
	return "k8s.urls"
}

var urlLibraryDecls = map[string][]cel.FunctionOpt{
	"url": {
		cel.Overload("string_to_url", []*cel.Type{cel.StringType}, URLType,
			cel.UnaryBinding(stringToURL))},
	"getScheme": {
		cel.MemberOverload("url_get_scheme", []*cel.Type{URLType}, cel.StringType,
			cel.UnaryBinding(getScheme))},
	"getHost": {
		cel.MemberOverload("url_get_host", []*cel.Type{URLType}, cel.StringType,
			cel.UnaryBinding(getHost))},
	"getHostname": {
		cel.MemberOverload("url_get_hostname", []*cel.Type{URLType}, cel.StringType,
			cel.UnaryBinding(getHostname))},
	"getPort": {
		cel.MemberOverload("url_get_port", []*cel.Type{URLType}, cel.StringType,
			cel.UnaryBinding(getPort))},
	"getEscapedPath": {
		cel.MemberOverload("url_get_escaped_path", []*cel.Type{URLType}, cel.StringType,
			cel.UnaryBinding(getEscapedPath))},
	"getQuery": {
		cel.MemberOverload("url_get_query", []*cel.Type{URLType},
			cel.MapType(cel.StringType, cel.ListType(cel.StringType)),
			cel.UnaryBinding(getQuery))},
	"isURL": {
		cel.Overload("is_url_string", []*cel.Type{cel.StringType}, cel.BoolType,
			cel.UnaryBinding(isURL))},
}

func (*urls) CompileOptions() []cel.EnvOption {
	options := []cel.EnvOption{}
	for name, overloads := range urlLibraryDecls {
		options = append(options, cel.Function(name, overloads...))
	}
	return options
}

func (*urls) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func stringToURL(arg ref.Val) ref.Val {
	s, ok := arg.Value().(string)
	if !ok {
		return types.MaybeNoSuchOverloadErr(arg)
	}
	// Use ParseRequestURI to check the URL before conversion.
	// ParseRequestURI requires absolute URLs and is used by the OpenAPIv3 'uri' format.
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return types.NewErr("URL parse error during conversion from string: %v", err)
	}
	// We must parse again with Parse since ParseRequestURI incorrectly parses URLs that contain a fragment
	// part and will incorrectly append the fragment to either the path or the query, depending on which it was adjacent to.
	u, err := url.Parse(s)
	if err != nil {
		// Errors are not expected here since Parse is a more lenient parser than ParseRequestURI.
		return types.NewErr("URL parse error during conversion from string: %v", err)
	}
	return URL{URL: u}
}

func getScheme(arg ref.Val) ref.Val {
	u, ok := arg.Value().(*url.URL)
	if !ok {
		return types.MaybeNoSuchOverloadErr(arg)
	}
	return types.String(u.Scheme)
}

func getHost(arg ref.Val) ref.Val {
	u, ok := arg.Value().(*url.URL)
	if !ok {
		return types.MaybeNoSuchOverloadErr(arg)
	}
	return types.String(u.Host)
}

func getHostname(arg ref.Val) ref.Val {
	u, ok := arg.Value().(*url.URL)
	if !ok {
		return types.MaybeNoSuchOverloadErr(arg)
	}
	return types.String(u.Hostname())
}

func getPort(arg ref.Val) ref.Val {
	u, ok := arg.Value().(*url.URL)
	if !ok {
		return types.MaybeNoSuchOverloadErr(arg)
	}
	return types.String(u.Port())
}

func getEscapedPath(arg ref.Val) ref.Val {
	u, ok := arg.Value().(*url.URL)
	if !ok {
		return types.MaybeNoSuchOverloadErr(arg)
	}
	return types.String(u.EscapedPath())
}

func getQuery(arg ref.Val) ref.Val {
	u, ok := arg.Value().(*url.URL)
	if !ok {
		return types.MaybeNoSuchOverloadErr(arg)
	}

	result := map[ref.Val]ref.Val{}
	for k, v := range u.Query() {
		result[types.String(k)] = types.NewStringList(types.DefaultTypeAdapter, v)
	}
	return types.NewRefValMap(types.DefaultTypeAdapter, result)
}

func isURL(arg ref.Val) ref.Val {
	s, ok := arg.Value().(string)
	if !ok {
		return types.MaybeNoSuchOverloadErr(arg)
	}
	_, err := url.ParseRequestURI(s)
	return types.Bool(err == nil)
}
