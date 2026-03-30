# JSONata Functions Reference

This document lists every function available in gomplate JSONata expressions. Functions are divided into two groups:

1. **Built-in** -- provided natively by the [JSONata specification](https://docs.jsonata.org/) via [gnata](https://github.com/recolabs/gnata). These always take precedence when a name collision exists.
2. **Custom** -- gomplate-specific extensions prefixed with `$` in expressions (e.g. `$sha256("hello")`).

> **Type note:** JSONata passes all numbers as `float64`. Functions like `$mathIsInt` work best with string arguments (e.g. `$mathIsInt("42")`).

---

## Built-in Functions

These are native JSONata functions. See the [JSONata docs](https://docs.jsonata.org/overview) for full details.

### String

| Function | Description |
|----------|-------------|
| `$string(value)` | Cast to string |
| `$length(str)` | String length |
| `$substring(str, start[, length])` | Extract substring |
| `$substringBefore(str, chars)` | Text before first occurrence |
| `$substringAfter(str, chars)` | Text after first occurrence |
| `$uppercase(str)` | Convert to uppercase |
| `$lowercase(str)` | Convert to lowercase |
| `$trim(str)` | Remove leading/trailing whitespace |
| `$pad(str, width[, char])` | Pad string to width |
| `$contains(str, pattern)` | Test if string contains pattern |
| `$split(str, separator[, limit])` | Split string into array |
| `$join(array[, separator])` | Join array into string |
| `$match(str, pattern[, limit])` | Regex match returning objects |
| `$replace(str, pattern, replacement[, limit])` | Regex replace |
| `$base64encode(str)` | Base64 encode |
| `$base64decode(str)` | Base64 decode |
| `$encodeUrl(str)` | URL encode (full URL) |
| `$encodeUrlComponent(str)` | URL encode (component) |
| `$decodeUrl(str)` | URL decode (full URL) |
| `$decodeUrlComponent(str)` | URL decode (component) |

### Numeric

| Function | Description |
|----------|-------------|
| `$number(value)` | Cast to number |
| `$abs(n)` | Absolute value |
| `$floor(n)` | Round down |
| `$ceil(n)` | Round up |
| `$round(n[, precision])` | Round to precision |
| `$power(base, exp)` | Exponentiation |
| `$sqrt(n)` | Square root |
| `$random()` | Random float in [0, 1) |
| `$formatNumber(n, picture[, options])` | Format number as string |
| `$formatBase(n, radix)` | Format number in given base |
| `$formatInteger(n, picture)` | Format integer |
| `$parseInteger(str, picture)` | Parse integer from string |

### Array / Aggregation

| Function | Description |
|----------|-------------|
| `$count(array)` | Number of elements |
| `$sum(array)` | Sum of numeric elements |
| `$max(array)` | Maximum value |
| `$min(array)` | Minimum value |
| `$average(array)` | Arithmetic mean |
| `$append(arr1, arr2)` | Concatenate arrays |
| `$reverse(array)` | Reverse order |
| `$sort(array[, comparator])` | Sort elements |
| `$shuffle(array)` | Randomise order |
| `$distinct(array)` | Remove duplicates |
| `$flatten(array)` | Flatten nested arrays |
| `$zip(arr1, arr2, ...)` | Zip arrays together |
| `$filter(array, fn)` | Filter elements by predicate |
| `$map(array, fn)` | Transform each element |
| `$reduce(array, fn[, init])` | Reduce to a single value |
| `$sift(object, fn)` | Filter object entries |
| `$each(object, fn)` | Map over object entries |
| `$single(array, fn)` | Find exactly one match |

### Object

| Function | Description |
|----------|-------------|
| `$keys(object)` | Array of keys |
| `$values(object)` | Array of values |
| `$spread(object)` | Spread object into array of single-key objects |
| `$merge(array)` | Merge array of objects into one |
| `$lookup(object, key)` | Lookup a key |

### Boolean / Control

| Function | Description |
|----------|-------------|
| `$boolean(value)` | Cast to boolean |
| `$not(value)` | Logical NOT |
| `$exists(value)` | Test if value exists |
| `$assert(condition, message)` | Assert a condition |
| `$type(value)` | Return type name |
| `$error(message)` | Throw an error |
| `$eval(expr[, context])` | Evaluate a JSONata expression at runtime |

### Date / Time

| Function | Description |
|----------|-------------|
| `$now()` | Current timestamp (ISO 8601) |
| `$millis()` | Current time in epoch milliseconds |
| `$fromMillis(ms[, picture])` | Convert epoch ms to string |
| `$toMillis(str[, picture])` | Convert string to epoch ms |

---

## Custom Functions (gomplate extensions)

### Strings

| Function | Signature | Description |
|----------|-----------|-------------|
| `$replaceAll` | `(old, new, str)` | Replace all occurrences of `old` with `new` |
| `$trimPrefix` | `(prefix, str)` | Remove leading prefix |
| `$trimSuffix` | `(suffix, str)` | Remove trailing suffix |
| `$title` | `(str)` | Title Case conversion |
| `$toUpper` | `(str)` | Uppercase (via gomplate, not JSONata `$uppercase`) |
| `$toLower` | `(str)` | Lowercase (via gomplate, not JSONata `$lowercase`) |
| `$trimSpace` | `(str)` | Remove leading/trailing whitespace |
| `$trunc` | `(length, str)` | Truncate to length |
| `$slug` | `(str)` | URL-safe slug (`"Hello World!" -> "hello-world"`) |
| `$quote` | `(str)` | Double-quote a string |
| `$squote` | `(str)` | Single-quote a string |
| `$shellQuote` | `(str)` | Shell-safe quoting |
| `$snakeCase` | `(str)` | Convert to `snake_case` |
| `$camelCase` | `(str)` | Convert to `camelCase` |
| `$kebabCase` | `(str)` | Convert to `kebab-case` |
| `$abbrev` | `(maxWidth, str)` or `(offset, maxWidth, str)` | Abbreviate with ellipsis |
| `$repeat` | `(count, str)` | Repeat string N times |
| `$sortStrings` | `(list)` | Sort a list of strings |
| `$splitN` | `(sep, n, str)` | Split with limit |
| `$indent` | `(str)` or `(width, str)` or `(width, pad, str)` | Indent text |
| `$wordWrap` | `(str)` or `(width, str)` or `(width, lbseq, str)` | Wrap text to width |
| `$runeCount` | `(str, ...)` | Count Unicode runes |
| `$humanDuration` | `(value)` | Human-readable duration (e.g. `"3 hours"`) |
| `$humanSize` | `(bytes)` | Human-readable byte size (e.g. `"1.5 GiB"`) |
| `$semver` | `(str)` | Parse semver returning `{major, minor, patch, prerelease, metadata, original}` |
| `$semverCompare` | `(v1, v2)` | `true` if `v1` constraint satisfied by `v2` |
| `$hasPrefix` | `(prefix, str)` | Test if string starts with prefix |
| `$hasSuffix` | `(suffix, str)` | Test if string ends with suffix |

#### Examples

```jsonata
$slug("Hello World!")                       /* "hello-world" */
$replaceAll("o", "0", "foo")               /* "f00" */
$trimPrefix("v", "v1.2.3")                 /* "1.2.3" */
$humanDuration(3600)                        /* "1 hour" */
$semver("1.2.3-beta+build").prerelease     /* "beta" */
```

### Collections

| Function | Signature | Description |
|----------|-----------|-------------|
| `$has` | `(object, key)` | Check if key exists in map/struct |
| `$dict` | `(key1, val1, ...)` | Create a map from key/value pairs |
| `$prepend` | `(value, list)` | Prepend element to list |
| `$uniq` | `(list)` | Remove duplicates (gomplate impl) |
| `$sortBy` | `(key, list)` | Sort list of maps by a key |
| `$pick` | `(key1, key2, ..., map)` | Select specific keys from map |
| `$omit` | `(key1, key2, ..., map)` | Exclude specific keys from map |
| `$coalesce` | `(val1, val2, ...)` | Return first non-nil, non-empty value |
| `$first` | `(list)` | First element |
| `$last` | `(list)` | Last element |
| `$matchLabel` | `(labels, key, patterns)` | Match Kubernetes-style label selectors |
| `$keyValToMap` | `(str)` | Parse `"key=val,key2=val2"` into a map |
| `$mapToKeyVal` | `(map)` | Convert map to `"key=val,key2=val2"` string |

#### Examples

```jsonata
$pick("name", "age", {"name": "Alice", "age": 30, "email": "a@b.c"})
/* {"name": "Alice", "age": 30} */

$coalesce(null, "", "hello")    /* "hello" */
$sortBy("name", [{"name": "B"}, {"name": "A"}])
/* [{"name": "A"}, {"name": "B"}] */
```

### Query Languages

| Function | Signature | Description |
|----------|-----------|-------------|
| `$jq` | `(expr, data)` | Execute a [jq](https://stedolan.github.io/jq/) query |
| `$jmespath` | `(expr, data)` | Execute a [JMESPath](https://jmespath.org/) query |
| `$jsonpath` | `(expr, data)` | Execute a [JSONPath](https://goessner.net/articles/JsonPath/) query |

#### Examples

```jsonata
$jq(".items[].name", data)
$jmespath("items[?status=='active'].name", data)
$jsonpath("$.items[*].name", data)
```

### Time / Duration

| Function | Signature | Description |
|----------|-----------|-------------|
| `$timeNow` | `()` | Current time as RFC 3339 string |
| `$timeZoneName` | `()` | Local timezone name |
| `$timeZoneOffset` | `()` | Local timezone UTC offset (seconds) |
| `$timeParse` | `(layout, value)` | Parse time using Go layout |
| `$timeParseLocal` | `(layout, value)` | Parse time in local timezone |
| `$timeParseInLocation` | `(layout, location, value)` | Parse time in named timezone |
| `$timeUnix` | `(seconds)` | Convert Unix timestamp to RFC 3339 |
| `$parseDuration` | `(str)` | Parse duration string (e.g. `"1h30m"`, `"2d"`) |
| `$parseDateTime` | `(str)` | Parse various datetime formats including datemath (`"now-1h"`) |
| `$inTimeRange` | `(timestamp, start, end)` | Check if time-of-day falls within range |
| `$inBusinessHours` | `(timestamp)` | Check if time is within configured business hours |

#### Examples

```jsonata
$timeNow()                                           /* "2024-01-15T14:30:00Z" */
$timeParse("2006-01-02", "2024-01-15")              /* "2024-01-15T00:00:00Z" */
$parseDuration("1h30m")                              /* "1h30m0s" */
$inTimeRange("2024-01-15T12:00:00Z", "09:00", "17:00")  /* true */
$timeUnix("1705312200")                              /* "2024-01-15T10:30:00Z" */
```

### Math

> Built-in JSONata math functions (`$abs`, `$floor`, `$ceil`, `$round`, `$power`, `$sqrt`, `$sum`, `$max`, `$min`) always take precedence.
> The `$math*` prefixed functions use gomplate's implementation which handles type coercion differently.

| Function | Signature | Description |
|----------|-----------|-------------|
| `$mathIsInt` | `(value)` | Check if value is an integer (use string input) |
| `$mathIsFloat` | `(value)` | Check if value is a float |
| `$mathIsNum` | `(value)` | Check if value is numeric |
| `$mathAdd` | `(n1, n2, ...)` | Addition (variadic) |
| `$mathSub` | `(a, b)` | Subtraction |
| `$mathMul` | `(n1, n2, ...)` | Multiplication (variadic) |
| `$mathDiv` | `(a, b)` | Division |
| `$mathRem` | `(a, b)` | Remainder / modulo |
| `$mathPow` | `(base, exp)` | Power |
| `$mathSeq` | `(end)` or `(start, end)` or `(start, end, step)` | Generate integer sequence |

#### Examples

```jsonata
$mathAdd(1, 2, 3)       /* 6 */
$mathDiv(10, 4)          /* 2.5 */
$mathSeq(1, 5)           /* [1, 2, 3, 4, 5] */
$mathRem(10, 3)          /* 1 */
```

### Data Formats

| Function | Signature | Description |
|----------|-----------|-------------|
| `$toJSON` | `(value)` | Serialize to JSON string |
| `$fromJSON` | `(str)` | Parse JSON string to object |
| `$toJSONPretty` | `(indent, value)` | Serialize to pretty-printed JSON |
| `$fromJSONArray` | `(str)` | Parse JSON array string |
| `$toYAML` | `(value)` | Serialize to YAML string |
| `$fromYAML` | `(str)` | Parse YAML string to object |
| `$fromYAMLArray` | `(str)` | Parse YAML array string |
| `$toTOML` | `(value)` | Serialize to TOML string |
| `$fromTOML` | `(str)` | Parse TOML string |
| `$toCSV` | `(value)` | Serialize to CSV string |
| `$fromCSV` | `(str)` | Parse CSV string |
| `$csvByRow` | `(str)` | Parse CSV as array of row maps |
| `$csvByColumn` | `(str)` | Parse CSV as map of column arrays |

#### Examples

```jsonata
$toJSON({"name": "Alice"})              /* '{"name":"Alice"}' */
$fromJSON('{"name":"Alice"}').name      /* "Alice" */
$toJSONPretty("  ", {"a": 1})           /* pretty-printed JSON with 2-space indent */
```

### Cryptographic Hashing

| Function | Signature | Description |
|----------|-----------|-------------|
| `$sha1` | `(input)` | SHA-1 hash (hex) |
| `$sha224` | `(input)` | SHA-224 hash (hex) |
| `$sha256` | `(input)` | SHA-256 hash (hex) |
| `$sha384` | `(input)` | SHA-384 hash (hex) |
| `$sha512` | `(input)` | SHA-512 hash (hex) |
| `$sha512_224` | `(input)` | SHA-512/224 hash (hex) |
| `$sha512_256` | `(input)` | SHA-512/256 hash (hex) |

#### Examples

```jsonata
$sha256("hello")    /* "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" */
```

### Regular Expressions

| Function | Signature | Description |
|----------|-----------|-------------|
| `$regexpFind` | `(pattern, input)` | Find first match |
| `$regexpFindAll` | `(pattern, input[, limit])` | Find all matches |
| `$regexpMatch` | `(pattern, input)` | Test if pattern matches |
| `$regexpReplace` | `(pattern, replacement, input)` | Replace all matches |
| `$regexpReplaceLiteral` | `(pattern, replacement, input)` | Replace with literal string |
| `$regexpSplit` | `(pattern, input[, limit])` | Split by pattern |
| `$regexpQuoteMeta` | `(str)` | Escape regex metacharacters |

#### Examples

```jsonata
$regexpMatch("[0-9]+", "abc123")                /* true */
$regexpFind("[0-9]+", "abc123def")              /* "123" */
$regexpReplace("[0-9]+", "NUM", "abc123def456") /* "abcNUMdefNUM" */
$regexpQuoteMeta("a.b")                         /* "a\.b" */
```

### UUID

| Function | Signature | Description |
|----------|-----------|-------------|
| `$uuidV1` | `()` | Generate UUID v1 (MAC + time) |
| `$uuidV4` | `()` | Generate UUID v4 (random) |
| `$uuidNil` | `()` | Nil UUID (`00000000-0000-0000-0000-000000000000`) |
| `$uuidIsValid` | `(str)` | Check if string is a valid UUID |
| `$uuidParse` | `(str)` | Parse and normalize a UUID string |
| `$uuidHash` | `(val1, val2, ...)` | Deterministic UUID from SHA-256 of inputs |

#### Examples

```jsonata
$uuidV4()                                                /* "550e8400-e29b-41d4-a716-446655440000" */
$uuidIsValid("550e8400-e29b-41d4-a716-446655440000")     /* true */
$uuidHash("my-resource", "namespace")                     /* same UUID every time for same inputs */
```

### Random

| Function | Signature | Description |
|----------|-----------|-------------|
| `$randomASCII` | `(count)` | Random printable ASCII string |
| `$randomAlpha` | `(count)` | Random alphabetic string |
| `$randomAlphaNum` | `(count)` | Random alphanumeric string |
| `$randomString` | `(count[, pattern])` or `(count, lower, upper)` | Random string from pattern or range |
| `$randomItem` | `(list)` | Random element from list |
| `$randomNumber` | `([min, max])` | Random integer (default 0--100) |
| `$randomFloat` | `([min, max])` | Random float (default 0.0--1.0) |

#### Examples

```jsonata
$randomAlpha(10)              /* "aBcDeFgHiJ" */
$randomNumber(1, 100)         /* 42 */
$randomItem(["a", "b", "c"]) /* "b" */
```

### File Path

| Function | Signature | Description |
|----------|-----------|-------------|
| `$filepathBase` | `(path)` | Filename from path |
| `$filepathDir` | `(path)` | Directory from path |
| `$filepathExt` | `(path)` | File extension |
| `$filepathClean` | `(path)` | Clean path (resolve `..`, extra slashes) |
| `$filepathJoin` | `(elem1, elem2, ...)` | Join path elements |
| `$filepathIsAbs` | `(path)` | Check if absolute path |
| `$filepathMatch` | `(pattern, name)` | Glob-match a path |
| `$filepathRel` | `(basepath, targpath)` | Relative path between two paths |
| `$filepathSplit` | `(path)` | Split into `[dir, file]` |
| `$filepathFromSlash` | `(path)` | Convert `/` to OS separator |
| `$filepathToSlash` | `(path)` | Convert OS separator to `/` |

#### Examples

```jsonata
$filepathBase("/foo/bar/baz.txt")        /* "baz.txt" */
$filepathDir("/foo/bar/baz.txt")         /* "/foo/bar" */
$filepathExt("file.tar.gz")             /* ".gz" */
$filepathJoin("/foo", "bar", "baz.txt") /* "/foo/bar/baz.txt" */
$filepathClean("/foo/../bar")            /* "/bar" */
```

### Network

| Function | Signature | Description |
|----------|-----------|-------------|
| `$netIsValidIP` | `(ip)` | Check if string is a valid IPv4 or IPv6 address |
| `$netContainsCIDR` | `(cidr, ip)` | Check if IP is within CIDR block |

#### Examples

```jsonata
$netIsValidIP("192.168.1.1")                  /* true */
$netContainsCIDR("10.0.0.0/8", "10.1.2.3")   /* true */
$netContainsCIDR("10.0.0.0/8", "192.168.1.1") /* false */
```

### Kubernetes

| Function | Signature | Description |
|----------|-----------|-------------|
| `$isHealthy` | `(resource)` | Check if K8s resource is healthy |
| `$isReady` | `(resource)` | Check if K8s resource is ready |
| `$getStatus` | `(resource)` | Get status string |
| `$getHealth` | `(resource)` | Get health as map `{status, health, ready, message}` |
| `$neat` | `(yamlStr)` | Pretty-print K8s YAML (remove managed fields, etc.) |
| `$k8sIsHealthy` | `(resource)` | Alias for `$isHealthy` |
| `$k8sIsReady` | `(resource)` | Alias for `$isReady` |
| `$k8sGetStatus` | `(resource)` | Alias for `$getStatus` |
| `$k8sGetHealth` | `(resource)` | Alias for `$getHealth` |
| `$k8sNeat` | `(yamlStr)` | Alias for `$neat` |
| `$k8sCPUAsMillicores` | `(str)` | Convert CPU string to millicores (`"500m"` -> `500`, `"1"` -> `1000`) |
| `$k8sMemoryAsBytes` | `(str)` | Convert memory string to bytes (`"1Gi"` -> `1073741824`) |

#### Examples

```jsonata
$isHealthy(pod)                   /* true */
$getStatus(deployment)            /* "Running" */
$k8sCPUAsMillicores("500m")       /* 500 */
$k8sMemoryAsBytes("2Gi")          /* 2147483648 */
```

### AWS

| Function | Signature | Description |
|----------|-----------|-------------|
| `$arnToMap` | `(arnStr)` | Parse ARN into `{service, region, account, resource}` |
| `$fromAWSMap` | `(tagList)` | Convert `[{Name: "k", Value: "v"}, ...]` to `{k: v}` |

#### Examples

```jsonata
$arnToMap("arn:aws:s3:us-east-1:123456789:my-bucket")
/* {"service": "s3", "region": "us-east-1", "account": "123456789", "resource": "my-bucket"} */

$fromAWSMap([{"Name": "env", "Value": "prod"}, {"Name": "team", "Value": "platform"}])
/* {"env": "prod", "team": "platform"} */
```

### Type Conversion

| Function | Signature | Description |
|----------|-----------|-------------|
| `$toBool` | `(value)` | Convert to boolean |
| `$toInt` | `(value)` | Convert to integer |
| `$toFloat64` | `(value)` | Convert to float64 |
| `$toString` | `(value)` | Convert to string |
| `$default` | `(fallback, value)` | Return `value` if truthy, else `fallback` |
| `$join` | `(list, separator)` | Join list elements with separator |

#### Examples

```jsonata
$toBool("true")          /* true */
$toInt(3.7)              /* 3 */
$default("N/A", "")      /* "N/A" */
$default("N/A", "hello") /* "hello" */
$join(["a", "b"], ",")   /* "a,b" */
```

### Test / Assert

| Function | Signature | Description |
|----------|-----------|-------------|
| `$testFail` | `([message])` | Fail with optional message |
| `$testRequired` | `([message,] value)` | Fail if value is nil/empty |
| `$testTernary` | `(trueVal, falseVal, condition)` | Ternary operator |
| `$testKind` | `(value)` | Return Go type kind (e.g. `"string"`, `"float64"`) |
| `$testIsKind` | `(kind, value)` | Check if value is of given kind |

#### Examples

```jsonata
$testTernary("yes", "no", true)    /* "yes" */
$testKind("hello")                 /* "string" */
$testIsKind("string", "hello")     /* true */
```

### Go Template Formatting

| Function | Signature | Description |
|----------|-----------|-------------|
| `$f` | `(template, data)` | Evaluate a Go template string against data |

Supports both `{{...}}` and `$(...)` delimiters.

#### Examples

```jsonata
$f("Hello {{.name}}", {"name": "World"})   /* "Hello World" */
$f("ID: $(.id)", {"id": 123})              /* "ID: 123" */
```

---

## Builtin Precedence

When a custom function has the same name as a JSONata built-in, the **built-in always wins**. This affects the following names where gomplate also provides an implementation:

`append`, `assert`, `boolean`, `contains`, `error`, `exists`, `flatten`, `join`, `keys`, `lookup`, `merge`, `not`, `reverse`, `sort`, `split`, `type`, `values`

For these, use the JSONata syntax. Gomplate-specific alternatives are available with different names where needed (e.g. `$sortBy` for sorting by key, `$uniq` for deduplication).
