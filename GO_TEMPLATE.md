# Go Templates

`template` expressions use the [Go Text Template](https://pkg.go.dev/text/template) library with additional functions provided by the [gomplate](https://docs.gomplate.ca/) library.

## Escaping

If you need to pass a template through a Helm Chart and prevent Helm from processing it, escape it:

```
{{`{{ .secret }}`}}
```

Alternatively, [change the templating delimiters](#delimiters).

## Multiline YAML

When using a YAML multiline string, use `|` and not `>` (which strips newlines):

```yaml
# Wrong - strips newlines
exec:
  script: >
    #! pwsh
    Get-Items | ConvertTo-JSON

# Correct - preserves newlines
exec:
  script: |
    #! pwsh
    Get-Items | ConvertTo-JSON
```

## Delimiters

The template delimiters can be changed from the defaults (`{{` and `}}`) using `gotemplate` comments:

```
# gotemplate: left-delim=$[[ right-delim=]]
$message = "$[[.config.name]]"
Write-Host "{{ $message }}"
```

---

## base64

### Encode

Encodes data as a Base64 string (standard Base64 encoding, [RFC4648 §4](https://tools.ietf.org/html/rfc4648#section-4)).

```
{{ base64.Encode "hello world" }}      // aGVsbG8gd29ybGQ=
{{ "hello world" | base64.Encode }}    // aGVsbG8gd29ybGQ=
```

### Decode

Decodes a Base64 string. Supports both standard and URL-safe encodings.

```
{{ base64.Decode "aGVsbG8gd29ybGQ=" }}         // hello world
{{ "aGVsbG8gd29ybGQ=" | base64.Decode }}        // hello world
```

---

## Collection

### dict

Creates a map with string keys from key/value pairs.

```
{{ coll.Dict "name" "Frank" "age" 42 | data.ToYAML }}
// age: 42
// name: Frank

{{ dict 1 2 3 | toJSON }}   // {"1":2,"3":""}

{{ define "T1" }}Hello {{ .thing }}!{{ end -}}
{{ template "T1" (dict "thing" "world")}}
{{ template "T1" (dict "thing" "everybody")}}
// Hello world!
// Hello everybody!
```

### slice

Creates a slice (array). Useful when ranging over variables.

```
{{ range slice "Bart" "Lisa" "Maggie" }}Hello, {{ . }}{{ end }}
// Hello, Bart
// Hello, Lisa
// Hello, Maggie
```

### has

Reports whether a given object has a property with the given key, or whether a slice contains the given value.

```
{{ $l := slice "foo" "bar" "baz" }}there is {{ if has $l "bar" }}a{{else}}no{{end}} bar
// there is a bar

{{ $o := dict "foo" "bar" "baz" "qux" }}
{{ if has $o "foo" }}{{ $o.foo }}{{ else }}THERE IS NO FOO{{ end }}   // bar
```

### jmespath

Evaluates a [JMESPath](https://jmespath.org/) expression against an object or JSON string.

```
{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ $data | jmespath "foo" }}   // 1

{{ $data | toJSON | jmespath "foo" }}   // 1
```

### jsonpath

Evaluates a [JSONPath](https://datatracker.ietf.org/doc/draft-ietf-jsonpath-base) expression against an object or JSON string.

```
{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ $data | jsonpath "$.foo" }}   // 1

{{ $data | toJSON | jsonpath "$.foo" }}   // 1
```

### jq

Filters input using the [jq query language](https://stedolan.github.io/jq/), implemented via [gojq](https://github.com/itchyny/gojq).

Returns a single value for single results, or an array for multiple results.

```
{{ .books | jq `[.works[]|{"title":.title,"authors":[.authors[].name],"published":.first_publish_year}][0]` }}
// map[authors:[Lewis Carroll] published:1865 title:Alice's Adventures in Wonderland]
```

### Keys

Returns a list of keys from one or more maps (ordered by map position then alphabetically).

```
{{ coll.Keys (dict "foo" 1 "bar" 2) }}   // [bar foo]
```

### Values

Returns a list of values from one or more maps (ordered by map position then key alphabetically).

```
{{ coll.Values (dict "foo" 1 "bar" 2) }}   // [2 1]
```

### append

Appends a value to the end of a list (produces a new list).

```
{{ slice 1 1 2 3 | append 5 }}   // [1 1 2 3 5]
```

### prepend

Prepends a value to the beginning of a list (produces a new list).

```
{{ slice 4 3 2 1 | prepend 5 }}   // [5 4 3 2 1]
```

### uniq

Removes duplicate values from a list, without changing order (produces a new list).

```
{{ slice 1 2 3 2 3 4 1 5 | uniq }}   // [1 2 3 4 5]
```

### flatten

Flattens a nested list. Defaults to complete flattening; can be limited with `depth`.

```
{{ "[[1,2],[],[3,4],[[[5],6],7]]" | jsonArray | flatten }}         // [1 2 3 4 5 6 7]
{{ coll.Flatten 2 ("[[1,2],[],[3,4],[[[5],6],7]]" | jsonArray) }}  // [1 2 3 4 [[5] 6] 7]
```

### reverse

Reverses a list (produces a new list).

```
{{ slice 4 3 2 1 | reverse }}   // [1 2 3 4]
```

### Sort

Sorts a list using natural sort order. Maps and structs can be sorted by a named key.

```
{{ slice "foo" "bar" "baz" | coll.Sort }}    // [bar baz foo]
{{ sort (slice 3 4 1 2 5) }}                 // [1 2 3 4 5]
```

### Merge

Combines multiple maps. The first map is the base; subsequent maps provide overrides (left-to-right precedence).

```
{{ $default := dict "foo" 1 "bar" 2}}
{{ $config := dict "foo" 8 }}
{{ merge $config $default }}
// map[bar:2 foo:8]
```

### Pick

Returns a new map containing only the specified keys.

```
{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ coll.Pick "foo" "baz" $data }}   // map[baz:3 foo:1]
```

### Omit

Returns a new map without the specified keys.

```
{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ coll.Omit "foo" "baz" $data }}   // map[bar:2]
```

---

## Convert

### bool

Converts a true-ish string to a boolean.

```
{{ $FOO := true }}
{{ if $FOO }}foo{{ else }}bar{{ end }}   // foo
```

### default

Provides a default value for empty inputs.

```
{{ "" | default "foo" }}      // foo
{{ "bar" | default "baz" }}   // bar
```

### Dict

Same as [`coll.Dict`](#dict) — creates a map with string keys.

```
{{ $dict := conv.Dict "name" "Frank" "age" 42 }}
{{ $dict | data.ToYAML }}
// age: 42
// name: Frank
```

### slice

Creates a slice. Same as [`coll.slice`](#slice).

```
{{ range slice "Bart" "Lisa" "Maggie" }}Hello, {{ . }}{{ end }}
```

### has

Same as [`coll.has`](#has).

### join

Concatenates array elements into a string with a separator.

```
{{ $a := slice 1 2 3 }}{{ join $a "-" }}   // 1-2-3
```

### urlParse

Parses a URL string.

```
{{ ($u := conv.URL "https://example.com:443/foo/bar").Host }}   // example.com:443
{{ (conv.URL "https://user:supersecret@example.com").Redacted }}  // https://user:xxxxx@example.com
```

### ParseInt

Parses a string as an `int64` with a given base.

```
{{ $val := conv.ParseInt "7C0" 16 32 }}   // 1984
```

### ParseFloat

Parses a string as a `float64`.

```
{{ $pi := conv.ParseFloat "3.14159265359" 64 }}
```

### ParseUint

Parses a string as a `uint64`.

```
{{ conv.ParseUint "FFFFFFFFFFFFFFFF" 16 64 }}   // 18446744073709551615
```

### ToBool

Converts input to a boolean. `true` values: `1`, `"t"`, `"true"`, `"yes"` (any case).

```
{{ conv.ToBool "yes" }}    // true
{{ conv.ToBool true }}     // true
{{ conv.ToBool false }}    // false
{{ conv.ToBool "blah" }}   // false
```

### ToBools

Converts a list of inputs to an array of booleans.

```
{{ conv.ToBools "yes" true "0x01" }}   // [true true true]
```

### ToInt64

Converts input to an `int64`.

```
{{ conv.ToInt64 "9223372036854775807" }}   // 9223372036854775807
{{ conv.ToInt64 "0x42" }}                 // 66
{{ conv.ToInt64 true }}                   // 1
```

### ToInt

Converts input to an `int`.

```
{{ conv.ToInt "0x42" }}   // 66
```

### ToInt64s / ToInts

Converts multiple inputs to an array of `int64`s / `int`s.

```
{{ conv.ToInt64s true 0x42 "123,456.99" "1.2345e+3" }}   // [1 66 123456 1234]
```

### ToFloat64

Converts input to a `float64`.

```
{{ conv.ToFloat64 "8.233e-1" }}   // 0.8233
{{ conv.ToFloat64 "9,000.09" }}   // 9000.09
```

### ToFloat64s

Converts multiple inputs to an array of `float64`s.

```
{{ conv.ToFloat64s true 0x42 "123,456.99" "1.2345e+3" }}   // [1 66 123456.99 1234.5]
```

### ToString

Converts any input to a `string`.

```
{{ conv.ToString 0xFF }}             // 255
{{ dict "foo" "bar" | conv.ToString }} // map[foo:bar]
```

### ToStrings

Converts multiple inputs to an array of `string`s.

```
{{ conv.ToStrings nil 42 true 0xF (slice 1 2 3) }}   // [nil 42 true 15 [1 2 3]]
```

---

## Cryptography

### crypto.SHA1 / SHA256 / SHA384 / SHA512

Computes a checksum as a hexadecimal string.

```
{{ crypto.SHA1 "foo" }}     // f1d2d2f924e986ac86fdf7b36c94bcdf32beec15
{{ crypto.SHA512 "bar" }}   // cc06808cbbee0510331aa97974132e8dc296aeb...
```

---

## Data

### json

Converts a JSON string into an object.

```
{{ ('{"hello":"world"}' | json).hello }}   // world
```

### jsonArray

Converts a JSON string into a slice. Only works for JSON Arrays.

```
{{ ('[ "you", "world" ]' | jsonArray) 1 }}   // world
```

### yaml

Converts a YAML string into an object.

```
{{ $FOO := "hello: world" }}
Hello {{ (yaml $FOO).hello }}   // Hello world
```

### yamlArray

Converts a YAML string into a slice.

```
{{ $FOO := "- hello\n- world" }}
{{ (yamlArray $FOO) 0 }}   // hello
```

### toml

Converts a TOML document into an object.

```
{{ $t := `[data]
hello = "world"` }}
Hello {{ (toml $t).data.hello }}   // Hello world
```

### csv

Converts a CSV-format string into a 2-dimensional string array.

```
{{ $c := `C,32
Go,25
COBOL,357` }}
{{ range ($c | csv) }}
{{ index . 0 }} has {{ index . 1 }} keywords.
{{ end }}
// C has 32 keywords.
// Go has 25 keywords.
// COBOL has 357 keywords.
```

### csvByRow

Converts a CSV-format string into a slice of maps.

```
{{ $c := `lang,keywords
C,32
Go,25` }}
{{ range ($c | csvByRow) }}
{{ .lang }} has {{ .keywords }} keywords.
{{ end }}
```

### csvByColumn

Like `csvByRow`, but returns a columnar map.

```
{{ $langs := ($c | csvByColumn ";" "lang,keywords").lang }}
```

### toJSON

Converts an object to a JSON document.

```
{{ (`{"foo":{"hello":"world"}}` | json).foo | toJSON }}   // {"hello":"world"}
```

### toJSONPretty

Converts an object to a pretty-printed JSON document.

```
{{ `{"hello":"world"}` | data.JSON | data.ToJSONPretty "  " }}
// {
//   "hello": "world"
// }
```

### toYAML

Converts an object to a YAML document.

```
{{ (`{"foo":{"hello":"world"}}` | data.JSON).foo | data.ToYAML }}   // hello: world
```

### toTOML

Converts an object to a TOML document.

```
{{ `{"foo":"bar"}` | data.JSON | data.ToTOML }}   // foo = "bar"
```

### toCSV

Converts a `[][]string` to a CSV document.

```
{{ $rows := (jsonArray `[["first","second"],["1","2"],["3","4"]]`) -}}
{{ data.ToCSV ";" $rows }}
```

---

## filepath

### Base

Returns the last element of path.

```
{{ filepath.Base "/tmp/foo" }}   // foo
```

### Clean

Returns the shortest equivalent path by lexical processing.

```
{{ filepath.Clean "/tmp//foo/../" }}   // /tmp
```

### Dir

Returns all but the last element of path.

```
{{ filepath.Dir "/tmp/foo" }}   // /tmp
```

### Ext

Returns the file name extension.

```
{{ filepath.Ext "/tmp/foo.csv" }}   // .csv
```

### IsAbs

Reports whether the path is absolute.

```
{{ filepath.IsAbs "/tmp/foo.csv" }}   // true
```

### Join

Joins path elements into a single path.

```
{{ filepath.Join "/tmp" "foo" "bar" }}   // /tmp/foo/bar
```

### Match

Reports whether name matches the shell file name pattern.

```
{{ filepath.Match "*.csv" "foo.csv" }}   // true
```

### Rel

Returns a relative path from basepath to targetpath.

```
{{ filepath.Rel "/a" "/a/b/c" }}   // b/c
```

### Split

Splits path into directory and file name.

```
{{ $p := filepath.Split "/tmp/foo" }}{{ index $p 0 }} / {{ index $p 1 }}
// /tmp/ / foo
```

### ToSlash / FromSlash

Converts path separators.

```
{{ filepath.ToSlash "/foo/bar" }}      // /foo/bar
{{ filepath.FromSlash "/foo/bar" }}    // /foo/bar
```

### VolumeName

Returns the leading volume name (Windows).

```
{{ filepath.VolumeName "C:/foo/bar" }}   // C:
```

---

## math

### Abs

Returns the absolute value.

```
{{ math.Abs -3.5 }}   // 3.5
{{ math.Abs -42 }}    // 42
```

### Add

Adds all given operands.

```
{{ math.Add 1 2 3 4 }}     // 10
{{ math.Add 1.5 2 3 }}     // 6.5
```

### Ceil

Returns the least integer ≥ the given value.

```
{{ math.Ceil 5.1 }}   // 6
{{ math.Ceil 42 }}    // 42
```

### Div

Divides the first number by the second (result is `float64`).

```
{{ math.Div 8 2 }}   // 4
{{ math.Div 3 2 }}   // 1.5
```

### Floor

Returns the greatest integer ≤ the given value.

```
{{ math.Floor 5.1 }}   // 4
{{ math.Floor 42 }}    // 42
```

### IsFloat

Returns whether the given number is a floating-point literal.

```
{{ math.IsFloat 1.0 }}    // true
{{ math.IsFloat 42 }}     // false
```

### IsInt

Returns whether the given number is an integer.

```
{{ math.IsInt 42 }}     // true
{{ math.IsInt 3.14 }}   // false
```

### IsNum

Returns whether the given input is a number.

```
{{ math.IsNum "foo" }}       // false
{{ math.IsNum 0xDeadBeef }}  // true
```

### Max

Returns the largest number provided.

```
{{ math.Max 0 8.0 4.5 "-1.5e-11" }}   // 8
```

### Min

Returns the smallest number provided.

```
{{ math.Min 0 8 4.5 "-1.5e-11" }}   // -1.5e-11
```

### Mul

Multiplies all given operands.

```
{{ math.Mul 8 8 2 }}   // 128
```

### Pow

Calculates an exponent.

```
{{ math.Pow 10 2 }}    // 100
{{ math.Pow 2 32 }}    // 4294967296
{{ math.Pow 1.5 2 }}   // 2.25
```

### Rem

Returns the remainder from integer division.

```
{{ math.Rem 5 3 }}    // 2
{{ math.Rem -5 3 }}   // -2
```

### Round

Returns the nearest integer, rounding half away from zero.

```
{{ math.Round 5.1 }}    // 5
{{ math.Round 42.9 }}   // 43
{{ math.Round 6.5 }}    // 7
```

### Seq

Returns a sequence from `start` to `end` in steps of `step`.

```
{{ range (math.Seq 5) }}{{.}} {{end}}                   // 1 2 3 4 5
{{ conv.Join (math.Seq 10 -3 2) ", " }}                 // 10, 8, 6, 4, 2, 0, -2
```

### Sub

Subtracts the second from the first operand.

```
{{ math.Sub 3 1 }}   // 2
```

---

## Path

### Base

Returns the last element of path.

```
{{ path.Base "/tmp/foo" }}   // foo
```

### Clean

Returns the shortest equivalent path.

```
{{ path.Clean "/tmp//foo/../" }}   // /tmp
```

### Dir

Returns the directory component of path.

```
{{ path.Dir "/tmp/foo" }}   // /tmp
```

### Ext

Returns the file name extension.

```
{{ path.Ext "/tmp/foo.csv" }}   // .csv
```

### IsAbs

Reports whether the path is absolute.

```
{{ path.IsAbs "/tmp/foo.csv" }}    // true
{{ path.IsAbs "../foo.csv" }}      // false
```

### Join

Joins path elements.

```
{{ path.Join "/tmp" "foo" "bar" }}   // /tmp/foo/bar
```

### Match

Reports whether name matches the shell file name pattern.

```
{{ path.Match "*.csv" "foo.csv" }}   // true
```

### Split

Splits path into directory and file name.

```
{{ index (path.Split "/tmp/foo") 0 }}   // /tmp/
```

---

## Random

### ASCII

Generates a random string from the printable 7-bit ASCII set.

```
{{ random.ASCII 8 }}   // _woJ%D&K
```

### Alpha

Generates a random alphabetical string.

```
{{ random.Alpha 42 }}   // oAqHKxHiytYicMxTMGHnUnAfltPVZDhFkVkgDvatJK
```

### AlphaNum

Generates a random alphanumeric string.

```
{{ random.AlphaNum 16 }}   // 4olRl9mRmVp1nqSm
```

### String

Generates a random string of a desired length with an optional character set.

```
{{ random.String 8 }}                      // FODZ01u_
{{ random.String 16 `[[:xdigit:]]` }}      // B9e0527C3e45E1f3
{{ random.String 8 "c" "m" }}              // ffmidgjc
```

### Item

Picks an element at random from a slice.

```
{{ random.Item (slice "red" "green" "blue") }}   // blue
```

### Number

Picks a random integer (default: 0–100).

```
{{ random.Number }}          // 55
{{ random.Number -10 10 }}   // -3
{{ random.Number 5 }}        // 2
```

### Float

Picks a random floating-point number.

```
{{ random.Float }}           // 0.2029946480303966
{{ random.Float 100 }}       // 71.28595374161743
{{ random.Float -100 200 }}  // 105.59119437834909
```

---

## regexp

### Find

Returns the leftmost match in `input` for the given regular expression.

```
{{ regexp.Find "[a-z]{3}" "foobar" }}          // foo
{{ "will not match" | regexp.Find "[0-9]" }}   // (empty)
```

### FindAll

Returns all successive matches of a regular expression.

```
{{ regexp.FindAll "[a-z]{3}" "foobar" | toJSON }}          // ["foo","bar"]
{{ "foo bar baz qux" | regexp.FindAll "[a-z]{3}" 3 | toJSON }}  // ["foo","bar","baz"]
```

### Match

Returns `true` if the regular expression matches the input.

```
{{ "hairyhenderson" | regexp.Match `^h` }}   // true
```

### QuoteMeta

Escapes all regular expression metacharacters in the input.

```
{{ `{hello}` | regexp.QuoteMeta }}   // \{hello\}
```

### Replace

Replaces matches of a regular expression with a replacement string (with `$` variable expansion).

```
{{ regexp.Replace "(foo)bar" "$1" "foobar" }}   // foo
{{ regexp.Replace "(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)" "${last}, ${first}" "Alan Turing" }}   // Turing, Alan
```

### ReplaceLiteral

Replaces matches with a literal replacement string (no `$` expansion).

```
{{ regexp.ReplaceLiteral "(foo)bar" "$1" "foobar" }}   // $1
```

### Split

Splits a string into sub-strings separated by the expression.

```
{{ regexp.Split `[\s,.]` "foo bar,baz.qux" | toJSON }}   // ["foo","bar","baz","qux"]
```

---

## Strings

### Abbrev

Abbreviates a string using `...`.

```
{{ "foobarbazquxquux" | strings.Abbrev 9 }}        // foobar...
{{ "foobarbazquxquux" | strings.Abbrev 6 9 }}      // ...baz...
```

### Contains

Reports whether a substring is contained within a string.

```
{{ "foo" | strings.Contains "f" }}   // true
```

### HasPrefix

Tests whether a string begins with a certain prefix.

```
{{ "http://example.com" | strings.HasPrefix "http://" }}   // true
```

### HasSuffix

Tests whether a string ends with a certain suffix.

```
{{ if not ("http://example.com" | strings.HasSuffix ":80") }}:80{{ end }}
// :80
```

### Indent

Indents each line of a string.

```
{{ `{"bar": {"baz": 2}}` | json | toYAML | strings.Indent "  " }}
//   bar:
//     baz: 2
```

### Split

Creates a slice by splitting a string on a delimiter.

```
{{ "Bart,Lisa,Maggie" | strings.Split "," }}   // [Bart Lisa Maggie]

{{ range ("Bart,Lisa,Maggie" | strings.Split ",") }}Hello, {{.}}{{ end }}
// Hello, Bart
// Hello, Lisa
// Hello, Maggie

{{ index ("Bart,Lisa,Maggie" | strings.Split ",") 0 }}   // Bart
```

### SplitN

Creates a slice by splitting a string on a delimiter with a count limit.

```
{{ range ("foo:bar:baz" | strings.SplitN ":" 2) }}{{.}}{{ end }}
// foo
// bar:baz
```

### Quote

Surrounds a string with double-quote characters.

```
{{ "in" | quote }}          // "in"
{{ strings.Quote 500 }}     // "500"
```

### Repeat

Returns `count` copies of the input string.

```
{{ "hello " | strings.Repeat 5 }}   // hello hello hello hello hello
```

### ReplaceAll

Replaces all occurrences of a substring.

```
{{ strings.ReplaceAll "." "-" "172.21.1.42" }}   // 172-21-1-42
{{ "172.21.1.42" | strings.ReplaceAll "." "-" }}  // 172-21-1-42
```

### Slug

Creates a URL-friendly slug from a string.

```
{{ "Hello, world!" | strings.Slug }}   // hello-world
```

### shellQuote

Quotes a string for safe use in a POSIX shell.

```
{{ slice "one word" "foo='bar baz'" | shellQuote }}
// 'one word' 'foo='"'"'bar baz'"'"''
```

### squote

Surrounds a string with single-quote characters.

```
{{ "in" | squote }}                // 'in'
{{ "it's a banana" | squote }}     // 'it''s a banana'
```

### Title

Converts a string to title-case.

```
{{ strings.Title "hello, world!" }}   // Hello, World!
```

### ToLower

Converts to lower-case.

```
{{ strings.ToLower "HELLO, WORLD!" }}   // hello, world!
```

### ToUpper

Converts to upper-case.

```
{{ strings.ToUpper "hello, world!" }}   // HELLO, WORLD!
```

### Trim

Removes the given characters from the beginning and end of a string.

```
{{ "_-foo-_" | strings.Trim "_-" }}   // foo
```

### TrimPrefix

Removes a leading prefix from a string.

```
{{ "hello, world" | strings.TrimPrefix "hello, " }}   // world
```

### TrimSpace

Removes leading and trailing whitespace.

```
{{ "  \n\t foo" | strings.TrimSpace }}   // foo
```

### TrimSuffix

Removes a trailing suffix from a string.

```
{{ "hello, world" | strings.TrimSuffix "world" }}   // hello, 
```

### Trunc

Truncates a string to the given length.

```
{{ "hello, world" | strings.Trunc 5 }}   // hello
```

### CamelCase

Converts a sentence to CamelCase.

```
{{ "Hello, World!" | strings.CamelCase }}   // HelloWorld
{{ "hello jello" | strings.CamelCase }}     // helloJello
```

### SnakeCase

Converts a sentence to snake_case.

```
{{ "Hello, World!" | strings.SnakeCase }}   // Hello_world
{{ "hello jello" | strings.SnakeCase }}     // hello_jello
```

### KebabCase

Converts a sentence to kebab-case.

```
{{ "Hello, World!" | strings.KebabCase }}   // Hello-world
{{ "hello jello" | strings.KebabCase }}     // hello-jello
```

### WordWrap

Inserts line breaks so that lines are at most `width` characters wide.

```
{{ "Hello, World!" | strings.WordWrap 7 }}
// Hello,
// World!
```

### RuneCount

Returns the number of Unicode code-points in the input.

```
{{ strings.RuneCount "Ω" }}   // 1
```

### contains (alias)

Reports whether the second string is contained within the first.

```
{{ if contains "foo" "f" }}yes{{ else }}no{{ end }}   // yes
```

### HasPrefix (alias)

Tests whether the string begins with a certain substring.

```
{{ if hasPrefix "http://example.com" "https" }}foo{{ else }}bar{{ end }}   // bar
```

### HasSuffix (alias)

Tests whether the string ends with a certain substring.

```
{{ if not (hasSuffix "http://example.com" ":80") }}:80{{ end }}   // :80
```

### split (alias)

Same as `strings.Split`.

```
{{ range split "Bart,Lisa,Maggie" "," }}Hello, {{ . }}{{ end }}
```

### splitN (alias)

Same as `strings.SplitN`.

```
{{ range splitN "foo:bar:baz" ":" 2 }}{{ . }}{{ end }}
// foo
// bar:baz
```

### Trim (alias)

```
{{ trim "  world " " " }}   // world
```

---

## Test

### Fail

Causes template generation to fail immediately, with an optional message.

```
{{ fail }}
{{ test.Fail "something is wrong!" }}
```

### IsKind

Reports whether the argument is of the given Kind.

```
{{ $data := "hello world" }}
{{ if isKind "string" $data }}{{ $data }} is a string{{ end }}
// hello world is a string
```

### Kind

Reports the _kind_ of the given argument.

```
{{ kind "hello world" }}                   // string
{{ dict "key1" true | test.Kind }}         // map
```

### ternary

Returns one of two values depending on whether the third is truthy.

```
{{ ternary "FOO" "BAR" false }}    // BAR
{{ ternary "FOO" "BAR" "yes" }}    // FOO
```

---

## Time

### Now

Returns the current local time as `time.Time`.

```
{{ (time.Now).UTC.Format "Day 2 of month 1 in year 2006 (timezone MST)" }}
// Day 14 of month 10 in year 2017 (timezone UTC)

{{ ((time.Now).AddDate 0 1 0).Format "Mon Jan 2 15:04:05 MST 2006" }}
// Tue Nov 14 09:57:02 EST 2017
```

### Parse

Parses a timestamp string using a given layout.

```
{{ (time.Parse "2006-01-02" "1993-10-23").Format "Monday January 2, 2006 MST" }}
// Saturday October 23, 1993 UTC
```

### ParseDuration

Parses a duration string. Valid units: `ns`, `us`, `ms`, `s`, `m`, `h`.

```
{{ ((time.Now).Add (time.ParseDuration "2h30m")).Format time.Kitchen }}   // 3:13AM
```

### ParseLocal

Same as `time.Parse`, but interprets the time in the local time zone.

```
{{ (time.ParseLocal time.Kitchen "6:00AM").Format "15:04 MST" }}   // 06:00 EST
```

### ParseInLocation

Same as `time.Parse`, but interprets the time in the given location's time zone.

```
{{ (time.ParseInLocation time.Kitchen "Africa/Luanda" "6:00AM").Format "15:04 MST" }}   // 06:00 LMT
```

### Since

Returns the time elapsed since a given time.

```
{{ $t := time.Parse time.RFC3339 "1970-01-01T00:00:00Z" }}
{{ time.Since $t }}   // 423365h0m24.353828924s
```

### Unix

Returns the `time.Time` corresponding to a Unix timestamp (seconds since epoch).

```
{{ (time.Unix 42).UTC.Format time.Stamp }}   // Jan  1, 00:00:42
```

### Until

Returns the duration until a given time.

```
{{ $t := time.Parse time.RFC3339 "2020-01-01T00:00:00Z" }}
{{ time.Until $t }}   // 14922h56m46.578625891s
```

### ZoneName

Returns the local system's time zone name.

```
{{ time.ZoneName }}   // UTC
```

### ZoneOffset

Returns the local system's time zone offset in seconds.

```
{{ time.ZoneOffset }}   // 0
```
