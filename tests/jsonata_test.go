package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flanksource/gomplate/v3"
)

// runJSONataTests is a helper that runs a slice of Test cases using the Jsonata field.
func runJSONataTests(t *testing.T, tests []Test) {
	t.Helper()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.expression, func(t *testing.T) {
			out, err := gomplate.RunTemplate(tc.env, gomplate.Template{
				Jsonata: tc.expression,
			})
			assert.NoError(t, err)
			assert.Equal(t, tc.out, out)
		})
	}
}

// TestJSONataBasic covers basic field-access, arithmetic, and comparisons.
func TestJSONataBasic(t *testing.T) {
	m := map[string]any{
		"x": map[string]any{
			"a": "b",
			"c": 1,
			"d": true,
		},
	}

	runJSONataTests(t, []Test{
		{m, "x.a", "b"},
		{m, "x.d", "true"},
		{nil, "1 + 2", "3"},
		{nil, "10 - 3", "7"},
		{nil, "3 * 4", "12"},
		{nil, "10 / 4", "2.5"},
		{nil, "1 = 1", "true"},
		{nil, "1 != 2", "true"},
		{nil, "1 < 2", "true"},
		{nil, `"hello" & " " & "world"`, "hello world"},
	})
}

// TestJSONataStrings exercises built-in JSONata string functions.
func TestJSONataStrings(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$length("hello")`, "5"},
		{nil, `$substring("hello world", 6)`, "world"},
		{nil, `$substringBefore("hello world", " ")`, "hello"},
		{nil, `$substringAfter("hello world", " ")`, "world"},
		{nil, `$trim("  hello  ")`, "hello"},
		{nil, `$uppercase("hello")`, "HELLO"},
		{nil, `$lowercase("HELLO")`, "hello"},
		{nil, `$join(["a", "b", "c"], ",")`, "a,b,c"},
		{nil, `$contains("hello world", "world")`, "true"},
		{nil, `$split("a,b,c", ",")`, "[a b c]"},
	})
}

// TestJSONataStringsFuncs tests gomplate-specific string functions.
func TestJSONataStringsFuncs(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$replaceAll("o", "0", "foo bar")`, "f00 bar"},
		{nil, `$trimPrefix("hello-", "hello-world")`, "world"},
		{nil, `$trimSuffix("-world", "hello-world")`, "hello"},
		{nil, `$title("hello world")`, "Hello World"},
		{nil, `$toUpper("hello")`, "HELLO"},
		{nil, `$toLower("HELLO")`, "hello"},
		{nil, `$trimSpace("  hello  ")`, "hello"},
		{nil, `$trunc(5, "hello world")`, "hello"},
		{nil, `$slug("Hello World!")`, "hello-world"},
		{nil, `$quote("hello")`, `"hello"`},
		{nil, `$squote("hello")`, "'hello'"},
		{nil, `$snakeCase("Hello World")`, "Hello_world"},
		{nil, `$camelCase("hello world")`, "helloWorld"},
		{nil, `$kebabCase("Hello World")`, "Hello-world"},
		{nil, `$hasPrefix("hello", "hello world")`, "true"},
		{nil, `$hasSuffix("world", "hello world")`, "true"},
		{nil, `$repeat(3, "ab")`, "ababab"},
	})
}

// TestJSONataArrays exercises built-in JSONata array functions.
func TestJSONataArrays(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$count([1, 2, 3])`, "3"},
		{nil, `$sum([1, 2, 3, 4, 5])`, "15"},
		{nil, `$max([1, 2, 3])`, "3"},
		{nil, `$min([1, 2, 3])`, "1"},
		{nil, `$reverse([1, 2, 3])`, "[3 2 1]"},
		{nil, `$sort([3, 1, 2])`, "[1 2 3]"},
		{nil, `$distinct([1, 2, 2, 3, 3])`, "[1 2 3]"},
		{nil, `$flatten([1, [2, 3], [4, [5]]])`, "[1 2 3 4 5]"},
	})
}

// TestJSONataCollFuncs tests gomplate-specific collection functions.
func TestJSONataCollFuncs(t *testing.T) {
	env := map[string]any{
		"m": map[string]any{"a": 1, "b": 2, "c": 3},
		"l": []any{3, 1, 2},
	}

	runJSONataTests(t, []Test{
		{env, `$has(m, "a")`, "true"},
		{env, `$has(m, "z")`, "false"},
		{env, `$first(l)`, "3"},
		{env, `$last(l)`, "2"},
		{env, `$coalesce(null, "", "hello")`, "hello"},
	})
}

// TestJSONataMath exercises built-in JSONata numeric functions.
func TestJSONataMath(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$abs(-5)`, "5"},
		{nil, `$floor(5.6)`, "5"},
		{nil, `$ceil(5.4)`, "6"},
		{nil, `$round(5.5)`, "6"},
		{nil, `$power(2, 8)`, "256"},
		{nil, `$sqrt(9)`, "3"},
	})
}

// TestJSONataMathFuncs tests gomplate-specific math functions.
func TestJSONataMathFuncs(t *testing.T) {
	runJSONataTests(t, []Test{
		// Note: JSONata passes all numbers as float64, so mathIsInt sees float64
		{nil, `$mathIsInt("42")`, "true"},
		{nil, `$mathIsFloat(3.14)`, "true"},
		{nil, `$mathIsNum(42)`, "true"},
		{nil, `$mathAdd(1, 2, 3)`, "6"},
		{nil, `$mathSub(10, 3)`, "7"},
		{nil, `$mathMul(2, 3)`, "6"},
		{nil, `$mathDiv(10, 4)`, "2.5"},
		{nil, `$mathRem(10, 3)`, "1"},
		{nil, `$mathPow(2, 8)`, "256"},
	})
}

// TestJSONataObjects exercises built-in JSONata object functions.
func TestJSONataObjects(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$count($keys({"a": 1, "b": 2}))`, "2"},
		{nil, `$sum($values({"a": 1, "b": 2}))`, "3"},
		{nil, `$merge([{"a": 1}, {"b": 2}]).a`, "1"},
	})
}

// TestJSONataCrypto covers the gomplate-specific crypto functions.
func TestJSONataCrypto(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$sha1("hello")`, "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{nil, `$sha224("hello")`, "ea09ae9cc6768c50fcee903ed054556e5bfc8347907f12598aa24193"},
		{nil, `$sha256("hello")`, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{nil, `$sha384("hello")`, "59e1748777448c69de6b800d7a33bbfb9ff1b463e44354c3553bcdb9c666fa90125a3c79f90397bdf5f6a13de828684f"},
		{nil, `$sha512("hello")`, "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
	})
}

// TestJSONataData covers data serialisation functions.
func TestJSONataData(t *testing.T) {
	runJSONataTests(t, []Test{
		{map[string]any{"name": "Alice"}, `$toJSON({"key": "val"})`, `{"key":"val"}`},
		{nil, `$fromJSON("{\"name\":\"Alice\"}").name`, "Alice"},
		{nil, `$toYAML({"key": "val"})`, "key: val\n"},
	})
}

// TestJSONataDataExtra covers additional data format functions.
func TestJSONataDataExtra(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$fromJSONArray("[1,2,3]")`, "[1 2 3]"},
	})
}

// TestJSONataRegexp covers the gomplate-specific regexp functions.
func TestJSONataRegexp(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$regexpMatch("[0-9]+", "abc123")`, "true"},
		{nil, `$regexpMatch("[0-9]+", "abc")`, "false"},
		{nil, `$regexpFind("[0-9]+", "abc123def")`, "123"},
		{nil, `$regexpReplace("[0-9]+", "NUM", "abc123def456")`, "abcNUMdefNUM"},
		{nil, `$regexpQuoteMeta("a.b")`, `a\.b`},
	})
}

// TestJSONataJQ covers the $jq custom function.
func TestJSONataJQ(t *testing.T) {
	env := map[string]any{
		"i": map[string]any{
			"name": "Aditya",
			"Address": map[string]any{
				"city_name": "Kathmandu",
			},
		},
	}
	runJSONataTests(t, []Test{
		{env, `$jq(".Address.city_name", i)`, "Kathmandu"},
	})
}

// TestJSONataGoTemplateFunc covers the $f function.
func TestJSONataGoTemplateFunc(t *testing.T) {
	runJSONataTests(t, []Test{
		{
			map[string]any{"row": map[string]any{"id": 123, "name": "test"}},
			`$f("{{id}}", row)`,
			"123",
		},
		{
			map[string]any{"row": map[string]any{"user": map[string]string{"name": "john"}}},
			`$f("Hello $(.user.name)", row)`,
			"Hello john",
		},
	})
}

// TestJSONataUUID tests UUID functions.
func TestJSONataUUID(t *testing.T) {
	t.Run("uuidV4 returns valid UUID", func(t *testing.T) {
		out, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$uuidV4()`})
		assert.NoError(t, err)
		assert.Len(t, out, 36)
		assert.Contains(t, out, "-")
	})

	t.Run("uuidNil returns nil UUID", func(t *testing.T) {
		out, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$uuidNil()`})
		assert.NoError(t, err)
		assert.Equal(t, "00000000-0000-0000-0000-000000000000", out)
	})

	t.Run("uuidIsValid", func(t *testing.T) {
		runJSONataTests(t, []Test{
			{nil, `$uuidIsValid("550e8400-e29b-41d4-a716-446655440000")`, "true"},
			{nil, `$uuidIsValid("not-a-uuid")`, "false"},
		})
	})

	t.Run("uuidHash is deterministic", func(t *testing.T) {
		out1, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$uuidHash("hello")`})
		assert.NoError(t, err)
		out2, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$uuidHash("hello")`})
		assert.NoError(t, err)
		assert.Equal(t, out1, out2)
		assert.Len(t, out1, 36)
	})
}

// TestJSONataFilepath tests filepath functions.
func TestJSONataFilepath(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$filepathBase("/foo/bar/baz.txt")`, "baz.txt"},
		{nil, `$filepathDir("/foo/bar/baz.txt")`, "/foo/bar"},
		{nil, `$filepathExt("/foo/bar/baz.txt")`, ".txt"},
		{nil, `$filepathJoin("/foo", "bar", "baz.txt")`, "/foo/bar/baz.txt"},
		{nil, `$filepathIsAbs("/foo/bar")`, "true"},
		{nil, `$filepathIsAbs("foo/bar")`, "false"},
		{nil, `$filepathClean("/foo/../bar")`, "/bar"},
	})
}

// TestJSONataNet tests network functions.
func TestJSONataNet(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$netIsValidIP("192.168.1.1")`, "true"},
		{nil, `$netIsValidIP("not-an-ip")`, "false"},
		{nil, `$netContainsCIDR("192.168.1.0/24", "192.168.1.100")`, "true"},
		{nil, `$netContainsCIDR("192.168.1.0/24", "192.168.2.100")`, "false"},
	})
}

// TestJSONataConv tests conversion functions.
func TestJSONataConv(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$toBool("true")`, "true"},
		{nil, `$toBool("false")`, "false"},
		{nil, `$toInt(3.7)`, "3"},
		{nil, `$toFloat64("3.14")`, "3.14"},
		{nil, `$toString(42)`, "42"},
		{nil, `$default("fallback", "")`, "fallback"},
		{nil, `$default("fallback", "value")`, "value"},
		{nil, `$join(["a", "b", "c"], ",")`, "a,b,c"},
	})
}

// TestJSONataTest tests test/assert functions.
func TestJSONataTestFuncs(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$testTernary("yes", "no", true)`, "yes"},
		{nil, `$testTernary("yes", "no", false)`, "no"},
		// Note: JSONata passes all numbers as float64
		{nil, `$testKind(42)`, "float64"},
		{nil, `$testKind("hello")`, "string"},
		{nil, `$testIsKind("string", "hello")`, "true"},
		{nil, `$testIsKind("int", "hello")`, "false"},
	})
}

// TestJSONataRandom tests random functions (just verifying they don't error).
func TestJSONataRandom(t *testing.T) {
	t.Run("randomAlpha returns string of correct length", func(t *testing.T) {
		out, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$randomAlpha(10)`})
		assert.NoError(t, err)
		assert.Len(t, out, 10)
	})

	t.Run("randomAlphaNum returns string of correct length", func(t *testing.T) {
		out, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$randomAlphaNum(8)`})
		assert.NoError(t, err)
		assert.Len(t, out, 8)
	})

	t.Run("randomNumber returns a number", func(t *testing.T) {
		out, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$randomNumber(1, 100)`})
		assert.NoError(t, err)
		assert.NotEmpty(t, out)
	})
}

// TestJSONataK8s tests Kubernetes functions.
func TestJSONataK8s(t *testing.T) {
	env := map[string]any{
		"pod": map[string]any{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]any{
				"name":      "test-pod",
				"namespace": "default",
			},
			"status": map[string]any{
				"phase": "Running",
				"conditions": []any{
					map[string]any{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}

	runJSONataTests(t, []Test{
		{env, `$isHealthy(pod)`, "true"},
		{env, `$isReady(pod)`, "true"},
		{env, `$getStatus(pod)`, "Running:  "},
	})

	t.Run("k8sCPUAsMillicores", func(t *testing.T) {
		runJSONataTests(t, []Test{
			{nil, `$k8sCPUAsMillicores("500m")`, "500"},
			{nil, `$k8sCPUAsMillicores("1")`, "1000"},
		})
	})

	t.Run("k8sMemoryAsBytes", func(t *testing.T) {
		out, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$k8sMemoryAsBytes("1Gi")`})
		assert.NoError(t, err)
		assert.Equal(t, "1073741824", out)
	})
}

// TestJSONataAWS tests AWS functions.
func TestJSONataAWS(t *testing.T) {
	runJSONataTests(t, []Test{
		{nil, `$arnToMap("arn:aws:s3:us-east-1:123456789:my-bucket").service`, "s3"},
		{nil, `$arnToMap("arn:aws:s3:us-east-1:123456789:my-bucket").region`, "us-east-1"},
		{nil, `$arnToMap("arn:aws:s3:us-east-1:123456789:my-bucket").account`, "123456789"},
		{nil, `$arnToMap("arn:aws:s3:us-east-1:123456789:my-bucket").resource`, "my-bucket"},
	})
}

// TestJSONataTemplate verifies that the Jsonata field on Template is honoured.
func TestJSONataTemplate(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]any
		jsonata  string
		expected string
	}{
		{name: "simple arithmetic", jsonata: "1 + 2", expected: "3"},
		{name: "field access", env: map[string]any{"greeting": "hello"}, jsonata: "greeting", expected: "hello"},
		{name: "boolean result", jsonata: "1 = 1", expected: "true"},
		{name: "string concat", jsonata: `"foo" & "bar"`, expected: "foobar"},
		{name: "null / missing returns empty string", jsonata: "missing_field", expected: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out, err := gomplate.RunTemplate(tc.env, gomplate.Template{Jsonata: tc.jsonata})
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, out)
		})
	}
}

// TestJSONataBuiltinPrecedence verifies gnata built-ins take precedence.
func TestJSONataBuiltinPrecedence(t *testing.T) {
	out, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$string(42)`})
	assert.NoError(t, err)
	assert.Equal(t, "42", out)

	out, err = gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$count([1, 2, 3])`})
	assert.NoError(t, err)
	assert.Equal(t, "3", out)
}

// TestJSONataTime tests time-related functions.
func TestJSONataTime(t *testing.T) {
	t.Run("timeNow returns non-empty", func(t *testing.T) {
		out, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$timeNow()`})
		assert.NoError(t, err)
		assert.NotEmpty(t, out)
		assert.True(t, strings.Contains(out, "T"))
	})

	t.Run("inTimeRange", func(t *testing.T) {
		out, err := gomplate.RunTemplate(nil, gomplate.Template{
			Jsonata: `$inTimeRange("2024-01-15T12:00:00Z", "09:00", "17:00")`,
		})
		assert.NoError(t, err)
		assert.Equal(t, "true", out)

		out, err = gomplate.RunTemplate(nil, gomplate.Template{
			Jsonata: `$inTimeRange("2024-01-15T22:00:00Z", "09:00", "17:00")`,
		})
		assert.NoError(t, err)
		assert.Equal(t, "false", out)
	})

	t.Run("parseDuration", func(t *testing.T) {
		out, err := gomplate.RunTemplate(nil, gomplate.Template{Jsonata: `$parseDuration("1h30m")`})
		assert.NoError(t, err)
		assert.Equal(t, "1h30m0s", out)
	})
}
