package gomplate

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gomplate/v3/coll"
	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/gomplate/v3/funcs"
	"github.com/flanksource/gomplate/v3/kubernetes"
	"github.com/recolabs/gnata"
)

// jsonataCustomFuncs returns all gomplate-specific functions for JSONata.
// Functions whose names collide with gnata built-ins are excluded in getJSONataCustomFuncs.
func jsonataCustomFuncs() map[string]gnata.CustomFunc {
	sf := &funcs.StringFuncs{}
	collFuncs := &funcs.CollFuncs{}
	timeFuncs := funcs.TimeNS()
	mathFuncs := &funcs.MathFuncs{}
	dataFuncs := &funcs.DataFuncs{}
	cryptoFuncs := &funcs.CryptoFuncs{}
	reFuncs := &funcs.ReFuncs{}
	randomFuncs := &funcs.RandomFuncs{}
	uuidFuncs := &funcs.UUIDFuncs{}
	fpFuncs := &funcs.FilePathFuncs{}
	netFuncs := &funcs.NetFuncs{}
	testFuncs := &funcs.TestFuncs{}
	k8sFuncs := &funcs.KubernetesFuncs{}
	convFuncs := &funcs.ConvFuncs{}

	m := make(map[string]gnata.CustomFunc)

	addStringFuncs(m, sf)
	addCollFuncs(m, collFuncs)
	addTimeFuncs(m, timeFuncs)
	addMathFuncs(m, mathFuncs)
	addDataFuncs(m, dataFuncs)
	addCryptoFuncs(m, cryptoFuncs)
	addRegexpFuncs(m, reFuncs)
	addRandomFuncs(m, randomFuncs)
	addUUIDFuncs(m, uuidFuncs)
	addFilepathFuncs(m, fpFuncs)
	addNetFuncs(m, netFuncs)
	addTestFuncs(m, testFuncs)
	addK8sFuncs(m, k8sFuncs)
	addAWSFuncs(m)
	addConvFuncs(m, convFuncs)
	addTemplateFuncs(m)

	return m
}

func addStringFuncs(m map[string]gnata.CustomFunc, sf *funcs.StringFuncs) {
	m["humanDuration"] = jsonata1(func(a any) (any, error) { return sf.HumanDuration(a) })
	m["humanSize"] = jsonata1(func(a any) (any, error) { return sf.HumanSize(a) })
	m["semver"] = jsonata1(func(a any) (any, error) {
		v, err := sf.SemverMap(conv.ToString(a))
		if err != nil {
			return nil, err
		}
		return v, nil
	})
	m["semverCompare"] = jsonata2(func(a, b any) (any, error) {
		return sf.SemverCompare(conv.ToString(a), conv.ToString(b))
	})
	m["abbrev"] = jsonataV(func(args []any) (any, error) { return sf.Abbrev(args...) })
	m["replaceAll"] = jsonata3(func(old, new_, s any) (any, error) {
		return sf.ReplaceAll(conv.ToString(old), conv.ToString(new_), s), nil
	})
	m["contains"] = jsonata2(func(substr, s any) (any, error) {
		return sf.Contains(conv.ToString(substr), s), nil
	})
	m["hasPrefix"] = jsonata2(func(prefix, s any) (any, error) {
		return sf.HasPrefix(conv.ToString(prefix), s), nil
	})
	m["hasSuffix"] = jsonata2(func(suffix, s any) (any, error) {
		return sf.HasSuffix(conv.ToString(suffix), s), nil
	})
	m["repeat"] = jsonata2(func(count, s any) (any, error) {
		return sf.Repeat(conv.ToInt(count), s)
	})
	m["sortStrings"] = jsonata1(func(a any) (any, error) { return sf.Sort(a) })
	m["splitN"] = jsonata3(func(sep, n, s any) (any, error) {
		return sf.SplitN(conv.ToString(sep), conv.ToInt(n), s), nil
	})
	m["trimPrefix"] = jsonata2(func(cutset, s any) (any, error) {
		return sf.TrimPrefix(conv.ToString(cutset), s), nil
	})
	m["trimSuffix"] = jsonata2(func(cutset, s any) (any, error) {
		return sf.TrimSuffix(conv.ToString(cutset), s), nil
	})
	m["title"] = jsonata1(func(a any) (any, error) { return sf.Title(a), nil })
	m["toUpper"] = jsonata1(func(a any) (any, error) { return sf.ToUpper(a), nil })
	m["toLower"] = jsonata1(func(a any) (any, error) { return sf.ToLower(a), nil })
	m["trimSpace"] = jsonata1(func(a any) (any, error) { return sf.TrimSpace(a), nil })
	m["trunc"] = jsonata2(func(length, s any) (any, error) {
		return sf.Trunc(conv.ToInt(length), s), nil
	})
	m["indent"] = jsonataV(func(args []any) (any, error) { return sf.Indent(args...) })
	m["slug"] = jsonata1(func(a any) (any, error) { return sf.Slug(a), nil })
	m["quote"] = jsonata1(func(a any) (any, error) { return sf.Quote(a), nil })
	m["shellQuote"] = jsonata1(func(a any) (any, error) { return sf.ShellQuote(a), nil })
	m["squote"] = jsonata1(func(a any) (any, error) { return sf.Squote(a), nil })
	m["snakeCase"] = jsonata1(func(a any) (any, error) { return sf.SnakeCase(a) })
	m["camelCase"] = jsonata1(func(a any) (any, error) { return sf.CamelCase(a) })
	m["kebabCase"] = jsonata1(func(a any) (any, error) { return sf.KebabCase(a) })
	m["wordWrap"] = jsonataV(func(args []any) (any, error) { return sf.WordWrap(args...) })
	m["runeCount"] = jsonataV(func(args []any) (any, error) { return sf.RuneCount(args...) })
}

func addCollFuncs(m map[string]gnata.CustomFunc, cf *funcs.CollFuncs) {
	m["has"] = jsonata2(func(in, key any) (any, error) {
		return cf.Has(in, conv.ToString(key)), nil
	})
	m["dict"] = jsonataV(func(args []any) (any, error) { return cf.Dict(args...) })
	m["prepend"] = jsonata2(func(v, list any) (any, error) { return cf.Prepend(v, list) })
	m["uniq"] = jsonata1(func(a any) (any, error) { return cf.Uniq(a) })
	m["sortBy"] = jsonata2(func(key, list any) (any, error) {
		return cf.Sort(key, list)
	})
	m["pick"] = jsonataV(func(args []any) (any, error) { return cf.Pick(args...) })
	m["omit"] = jsonataV(func(args []any) (any, error) { return cf.Omit(args...) })
	m["coalesce"] = jsonataV(func(args []any) (any, error) { return cf.Coalesce(args...), nil })
	m["first"] = jsonata1(func(a any) (any, error) { return cf.First(a), nil })
	m["last"] = jsonata1(func(a any) (any, error) { return cf.Last(a), nil })
	m["matchLabel"] = jsonata3(func(labels, key, patterns any) (any, error) {
		labelsMap, ok := labels.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("$matchLabel: expected map[string]any, got %T", labels)
		}
		valuePatterns := strings.Split(conv.ToString(patterns), ",")
		return coll.MatchLabel(labelsMap, conv.ToString(key), valuePatterns...), nil
	})
	m["keyValToMap"] = jsonata1(func(a any) (any, error) {
		return coll.KeyValToMap(conv.ToString(a))
	})
	m["mapToKeyVal"] = jsonata1(func(a any) (any, error) {
		m, ok := a.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("$mapToKeyVal: expected map[string]any, got %T", a)
		}
		return coll.MapToKeyVal(m), nil
	})
	m["jq"] = jsonata2(func(expr, data any) (any, error) {
		return coll.JQ(context.Background(), conv.ToString(expr), data)
	})
	m["jmespath"] = jsonata2(func(expr, data any) (any, error) {
		return coll.JMESPath(conv.ToString(expr), data)
	})
	m["jsonpath"] = jsonata2(func(expr, data any) (any, error) {
		return coll.JSONPath(conv.ToString(expr), data)
	})
}

func addTimeFuncs(m map[string]gnata.CustomFunc, tf *funcs.TimeFuncs) {
	m["timeNow"] = jsonata0(func() (any, error) { return tf.Now().Format("2006-01-02T15:04:05Z07:00"), nil })
	m["timeZoneName"] = jsonata0(func() (any, error) { return tf.ZoneName(), nil })
	m["timeZoneOffset"] = jsonata0(func() (any, error) { return tf.ZoneOffset(), nil })
	m["timeParse"] = jsonata2(func(layout, value any) (any, error) {
		t, err := tf.Parse(conv.ToString(layout), value)
		if err != nil {
			return nil, err
		}
		return t.Format("2006-01-02T15:04:05Z07:00"), nil
	})
	m["timeParseLocal"] = jsonata2(func(layout, value any) (any, error) {
		t, err := tf.ParseLocal(conv.ToString(layout), value)
		if err != nil {
			return nil, err
		}
		return t.Format("2006-01-02T15:04:05Z07:00"), nil
	})
	m["timeParseInLocation"] = jsonata3(func(layout, location, value any) (any, error) {
		t, err := tf.ParseInLocation(conv.ToString(layout), conv.ToString(location), value)
		if err != nil {
			return nil, err
		}
		return t.Format("2006-01-02T15:04:05Z07:00"), nil
	})
	m["timeUnix"] = jsonata1(func(a any) (any, error) {
		t, err := tf.Unix(a)
		if err != nil {
			return nil, err
		}
		return t.Format("2006-01-02T15:04:05Z07:00"), nil
	})
	m["parseDuration"] = jsonata1(func(a any) (any, error) {
		d, err := tf.ParseDuration(a)
		if err != nil {
			return nil, err
		}
		return d.String(), nil
	})
	m["parseDateTime"] = jsonata1(func(a any) (any, error) {
		t := funcs.ParseDateTime(conv.ToString(a))
		if t == nil {
			return nil, nil
		}
		return t.Format("2006-01-02T15:04:05Z07:00"), nil
	})
	m["inTimeRange"] = jsonata3(func(t, start, end any) (any, error) {
		return tf.InTimeRange(t, conv.ToString(start), conv.ToString(end))
	})
	m["inBusinessHours"] = jsonata1(func(a any) (any, error) {
		return tf.InBusinessHour(conv.ToString(a))
	})
}

func addMathFuncs(m map[string]gnata.CustomFunc, mf *funcs.MathFuncs) {
	m["mathIsInt"] = jsonata1(func(a any) (any, error) { return mf.IsInt(a), nil })
	m["mathIsFloat"] = jsonata1(func(a any) (any, error) { return mf.IsFloat(a), nil })
	m["mathIsNum"] = jsonata1(func(a any) (any, error) { return mf.IsNum(a), nil })
	m["mathAdd"] = jsonataV(func(args []any) (any, error) { return mf.Add(args...), nil })
	m["mathMul"] = jsonataV(func(args []any) (any, error) { return mf.Mul(args...), nil })
	m["mathSub"] = jsonata2(func(a, b any) (any, error) { return mf.Sub(a, b), nil })
	m["mathDiv"] = jsonata2(func(a, b any) (any, error) { return mf.Div(a, b) })
	m["mathRem"] = jsonata2(func(a, b any) (any, error) { return mf.Rem(a, b), nil })
	m["mathPow"] = jsonata2(func(a, b any) (any, error) { return mf.Pow(a, b), nil })
	m["mathSeq"] = jsonataV(func(args []any) (any, error) { return mf.Seq(args...) })
}

func addDataFuncs(m map[string]gnata.CustomFunc, df *funcs.DataFuncs) {
	m["toJSON"] = jsonata1(func(a any) (any, error) { return df.ToJSON(a) })
	m["fromJSON"] = jsonata1(func(a any) (any, error) { return df.JSON(a) })
	m["toYAML"] = jsonata1(func(a any) (any, error) { return df.ToYAML(a) })
	m["fromYAML"] = jsonata1(func(a any) (any, error) { return df.YAML(a) })
	m["toJSONPretty"] = jsonata2(func(indent, in any) (any, error) {
		return df.ToJSONPretty(conv.ToString(indent), in)
	})
	m["fromJSONArray"] = jsonata1(func(a any) (any, error) { return df.JSONArray(a) })
	m["fromYAMLArray"] = jsonata1(func(a any) (any, error) { return df.YAMLArray(a) })
	m["toTOML"] = jsonata1(func(a any) (any, error) { return df.ToTOML(a) })
	m["fromTOML"] = jsonata1(func(a any) (any, error) { return df.TOML(a) })
	m["toCSV"] = jsonataV(func(args []any) (any, error) { return df.ToCSV(args...) })
	m["fromCSV"] = jsonataV(func(args []any) (any, error) {
		strs := conv.ToStrings(args...)
		return df.CSV(strs...)
	})
	m["csvByRow"] = jsonataV(func(args []any) (any, error) {
		strs := conv.ToStrings(args...)
		return df.CSVByRow(strs...)
	})
	m["csvByColumn"] = jsonataV(func(args []any) (any, error) {
		strs := conv.ToStrings(args...)
		return df.CSVByColumn(strs...)
	})
}

func addCryptoFuncs(m map[string]gnata.CustomFunc, cf *funcs.CryptoFuncs) {
	m["sha1"] = jsonata1(func(a any) (any, error) { return cf.SHA1(a), nil })
	m["sha224"] = jsonata1(func(a any) (any, error) { return cf.SHA224(a), nil })
	m["sha256"] = jsonata1(func(a any) (any, error) { return cf.SHA256(a), nil })
	m["sha384"] = jsonata1(func(a any) (any, error) { return cf.SHA384(a), nil })
	m["sha512"] = jsonata1(func(a any) (any, error) { return cf.SHA512(a), nil })
	m["sha512_224"] = jsonata1(func(a any) (any, error) { return cf.SHA512_224(a), nil })
	m["sha512_256"] = jsonata1(func(a any) (any, error) { return cf.SHA512_256(a), nil })
}

func addRegexpFuncs(m map[string]gnata.CustomFunc, rf *funcs.ReFuncs) {
	m["regexpFind"] = jsonata2(func(re, input any) (any, error) { return rf.Find(re, input) })
	m["regexpFindAll"] = jsonataV(func(args []any) (any, error) { return rf.FindAll(args...) })
	m["regexpMatch"] = jsonata2(func(re, input any) (any, error) { return rf.Match(re, input), nil })
	m["regexpReplace"] = jsonata3(func(re, repl, input any) (any, error) {
		return rf.Replace(re, repl, input), nil
	})
	m["regexpReplaceLiteral"] = jsonata3(func(re, repl, input any) (any, error) {
		return rf.ReplaceLiteral(re, repl, input)
	})
	m["regexpSplit"] = jsonataV(func(args []any) (any, error) { return rf.Split(args...) })
	m["regexpQuoteMeta"] = jsonata1(func(a any) (any, error) { return rf.QuoteMeta(a), nil })
}

func addRandomFuncs(m map[string]gnata.CustomFunc, rf *funcs.RandomFuncs) {
	m["randomASCII"] = jsonata1(func(a any) (any, error) { return rf.ASCII(a) })
	m["randomAlpha"] = jsonata1(func(a any) (any, error) { return rf.Alpha(a) })
	m["randomAlphaNum"] = jsonata1(func(a any) (any, error) { return rf.AlphaNum(a) })
	m["randomString"] = jsonataV(func(args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("$randomString: expected at least 1 argument")
		}
		return rf.String(args[0], args[1:]...)
	})
	m["randomItem"] = jsonata1(func(a any) (any, error) { return rf.Item(a) })
	m["randomNumber"] = jsonataV(func(args []any) (any, error) { return rf.Number(args...) })
	m["randomFloat"] = jsonataV(func(args []any) (any, error) { return rf.Float(args...) })
}

func addUUIDFuncs(m map[string]gnata.CustomFunc, uf *funcs.UUIDFuncs) {
	m["uuidV1"] = jsonata0(func() (any, error) { return uf.V1() })
	m["uuidV4"] = jsonata0(func() (any, error) { return uf.V4() })
	m["uuidNil"] = jsonata0(func() (any, error) { return uf.Nil() })
	m["uuidIsValid"] = jsonata1(func(a any) (any, error) { return uf.IsValid(a) })
	m["uuidParse"] = jsonata1(func(a any) (any, error) { return uf.Parse(a) })
	m["uuidHash"] = jsonataV(func(args []any) (any, error) { return uf.HashUUID(args...) })
}

func addFilepathFuncs(m map[string]gnata.CustomFunc, fp *funcs.FilePathFuncs) {
	m["filepathBase"] = jsonata1(func(a any) (any, error) { return fp.Base(a), nil })
	m["filepathClean"] = jsonata1(func(a any) (any, error) { return fp.Clean(a), nil })
	m["filepathDir"] = jsonata1(func(a any) (any, error) { return fp.Dir(a), nil })
	m["filepathExt"] = jsonata1(func(a any) (any, error) { return fp.Ext(a), nil })
	m["filepathFromSlash"] = jsonata1(func(a any) (any, error) { return fp.FromSlash(a), nil })
	m["filepathToSlash"] = jsonata1(func(a any) (any, error) { return fp.ToSlash(a), nil })
	m["filepathIsAbs"] = jsonata1(func(a any) (any, error) { return fp.IsAbs(a), nil })
	m["filepathJoin"] = jsonataV(func(args []any) (any, error) { return fp.Join(args...), nil })
	m["filepathMatch"] = jsonata2(func(pattern, name any) (any, error) { return fp.Match(pattern, name) })
	m["filepathRel"] = jsonata2(func(base, targ any) (any, error) { return fp.Rel(base, targ) })
	m["filepathSplit"] = jsonata1(func(a any) (any, error) { return fp.Split(a), nil })
}

func addNetFuncs(m map[string]gnata.CustomFunc, nf *funcs.NetFuncs) {
	m["netContainsCIDR"] = jsonata2(func(cidr, ip any) (any, error) {
		return nf.ContainsCIDR(conv.ToString(cidr), conv.ToString(ip)), nil
	})
	m["netIsValidIP"] = jsonata1(func(a any) (any, error) {
		return nf.IsValidIP(conv.ToString(a)), nil
	})
}

func addTestFuncs(m map[string]gnata.CustomFunc, tf *funcs.TestFuncs) {
	m["testFail"] = jsonataV(func(args []any) (any, error) { return tf.Fail(args...) })
	m["testRequired"] = jsonataV(func(args []any) (any, error) { return tf.Required(args...) })
	m["testTernary"] = jsonata3(func(tval, fval, b any) (any, error) { return tf.Ternary(tval, fval, b), nil })
	m["testKind"] = jsonata1(func(a any) (any, error) { return tf.Kind(a), nil })
	m["testIsKind"] = jsonata2(func(kind, arg any) (any, error) {
		return tf.IsKind(conv.ToString(kind), arg), nil
	})
}

func addK8sFuncs(m map[string]gnata.CustomFunc, kf *funcs.KubernetesFuncs) {
	m["isHealthy"] = jsonata1(func(a any) (any, error) { return kf.IsHealthy(a), nil })
	m["isReady"] = jsonata1(func(a any) (any, error) { return kf.IsReady(a), nil })
	m["getStatus"] = jsonata1(func(a any) (any, error) { return kf.GetStatus(a), nil })
	m["getHealth"] = jsonata1(func(a any) (any, error) { return kf.GetHealthMap(a), nil })
	m["neat"] = jsonata1(func(a any) (any, error) { return kf.Neat(conv.ToString(a)) })
	m["k8sIsHealthy"] = m["isHealthy"]
	m["k8sIsReady"] = m["isReady"]
	m["k8sGetStatus"] = m["getStatus"]
	m["k8sGetHealth"] = m["getHealth"]
	m["k8sNeat"] = m["neat"]
	m["k8sCPUAsMillicores"] = jsonata1(func(a any) (any, error) {
		return kubernetes.CPUAsMillicores(conv.ToString(a)), nil
	})
	m["k8sMemoryAsBytes"] = jsonata1(func(a any) (any, error) {
		return kubernetes.MemoryAsBytes(conv.ToString(a)), nil
	})
}

func addAWSFuncs(m map[string]gnata.CustomFunc) {
	m["arnToMap"] = jsonata1(func(a any) (any, error) {
		arn := conv.ToString(a)
		parts := strings.Split(arn, ":")
		if len(parts) < 6 {
			return nil, fmt.Errorf("$arnToMap: invalid ARN %q: expected at least 6 colon-separated parts", arn)
		}
		return map[string]any{
			"service":  parts[2],
			"region":   parts[3],
			"account":  parts[4],
			"resource": parts[5],
		}, nil
	})
	m["fromAWSMap"] = jsonata1(func(a any) (any, error) {
		list, ok := a.([]any)
		if !ok {
			return nil, fmt.Errorf("$fromAWSMap: expected []any, got %T", a)
		}
		out := make(map[string]any)
		for _, item := range list {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			out[conv.ToString(m["Name"])] = m["Value"]
		}
		return out, nil
	})
}

func addConvFuncs(m map[string]gnata.CustomFunc, cf *funcs.ConvFuncs) {
	m["toBool"] = jsonata1(func(a any) (any, error) { return cf.ToBool(a), nil })
	m["toInt"] = jsonata1(func(a any) (any, error) { return cf.ToInt(a), nil })
	m["toFloat64"] = jsonata1(func(a any) (any, error) { return cf.ToFloat64(a), nil })
	m["toString"] = jsonata1(func(a any) (any, error) { return cf.ToString(a), nil })
	m["default"] = jsonata2(func(def, in any) (any, error) { return cf.Default(def, in), nil })
	m["join"] = jsonata2(func(in, sep any) (any, error) { return cf.Join(in, conv.ToString(sep)) })
}

func addTemplateFuncs(m map[string]gnata.CustomFunc) {
	m["f"] = func(args []any, _ any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("$f: expected 2 arguments (format, data), got %d", len(args))
		}
		format := conv.ToString(args[0])
		data := args[1]
		env := map[string]any{}
		switch v := data.(type) {
		case map[string]any:
			env = v
		case map[string]string:
			for k, val := range v {
				env[k] = val
			}
		default:
			env["data"] = v
		}
		st := StructTemplater{
			Context:        newContext(),
			Values:         env,
			ValueFunctions: true,
			DelimSets: []Delims{
				{Left: "$(", Right: ")"},
				{Left: "{{", Right: "}}"},
			},
		}
		return st.Template(format)
	}
}

// jsonata0 wraps a 0-arg function as a gnata.CustomFunc.
func jsonata0(fn func() (any, error)) gnata.CustomFunc {
	return func(args []any, _ any) (any, error) {
		return fn()
	}
}

// jsonata1 wraps a 1-arg function as a gnata.CustomFunc.
func jsonata1(fn func(any) (any, error)) gnata.CustomFunc {
	return func(args []any, _ any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
		}
		return fn(args[0])
	}
}

// jsonata2 wraps a 2-arg function as a gnata.CustomFunc.
func jsonata2(fn func(any, any) (any, error)) gnata.CustomFunc {
	return func(args []any, _ any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
		}
		return fn(args[0], args[1])
	}
}

// jsonata3 wraps a 3-arg function as a gnata.CustomFunc.
func jsonata3(fn func(any, any, any) (any, error)) gnata.CustomFunc {
	return func(args []any, _ any) (any, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("expected 3 arguments, got %d", len(args))
		}
		return fn(args[0], args[1], args[2])
	}
}

// jsonataV wraps a variadic function as a gnata.CustomFunc.
func jsonataV(fn func([]any) (any, error)) gnata.CustomFunc {
	return func(args []any, _ any) (any, error) {
		return fn(args)
	}
}
