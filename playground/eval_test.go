package playground

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlayground(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "playground")
}

// run evaluates against a playground with no host extensions, which is what
// most specs are about.
func run(req Request) (*Response, error) {
	return runWith(Options{}, req)
}

func runWith(options Options, req Request) (*Response, error) {
	handler, err := NewHandler(options)
	if err != nil {
		return nil, err
	}
	return handler.Evaluate(context.Background(), req)
}

var _ = ginkgo.Describe("evaluating a playground request", func() {
	const podInput = `
pod:
  metadata:
    name: web
  status:
    phase: Running
count: 3
`

	ginkgo.DescribeTable("returns the same result the language itself would",
		func(req Request, want string) {
			resp, err := run(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Error).To(BeNil(), "unexpected error: %+v", resp.Error)
			Expect(resp.Result).To(Equal(want))
		},
		ginkgo.Entry("cel field access",
			Request{Language: LanguageCEL, Source: `pod.metadata.name`, Input: podInput}, "web"),
		ginkgo.Entry("cel arithmetic",
			Request{Language: LanguageCEL, Source: `count * 2`, Input: podInput}, "6"),
		ginkgo.Entry("cel member function",
			Request{Language: LanguageCEL, Source: `pod.status.phase.lowerAscii()`, Input: podInput}, "running"),
		ginkgo.Entry("cel k8s helper",
			Request{Language: LanguageCEL, Source: `k8s.cpuAsMillicores("500m")`}, "500"),
		ginkgo.Entry("go template",
			Request{Language: LanguageGoTemplate, Source: `{{ .pod.metadata.name }}`, Input: podInput}, "web"),
		ginkgo.Entry("go template pipeline",
			Request{Language: LanguageGoTemplate, Source: `{{ .pod.metadata.name | strings.ToUpper }}`, Input: podInput}, "WEB"),
		ginkgo.Entry("jsonpath",
			Request{Language: LanguageJSONPath, Source: `$.pod.metadata.name`, Input: podInput}, "web"),
		ginkgo.Entry("jsonpath built-in function",
			Request{
				Language: LanguageJSONPath,
				Source:   `$.items[?(length(@.name) == 3)].name`,
				Input:    "items:\n  - name: web\n  - name: worker\n",
			}, "web"),
		ginkgo.Entry("javascript",
			Request{Language: LanguageJavaScript, Source: `count + 1`, Input: podInput}, "4"),
		ginkgo.Entry("javascript registered function",
			Request{Language: LanguageJavaScript, Source: `startsWith(pod.metadata.name, "we")`, Input: podInput}, "true"),
	)

	ginkgo.It("honours custom go-template delimiters", func() {
		resp, err := run(Request{
			Language:   LanguageGoTemplate,
			Source:     `$[[ .count ]]`,
			Input:      podInput,
			LeftDelim:  "$[[",
			RightDelim: "]]",
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).To(BeNil())
		Expect(resp.Result).To(Equal("3"))
	})

	ginkgo.It("evaluates jsonpath, which RunTemplateContext itself ignores", func() {
		// Template.JSONPath is declared but never dispatched on, so routing had
		// to be explicit. Guard against it silently returning nothing again.
		resp, err := run(Request{Language: LanguageJSONPath, Source: `$.count`, Input: podInput})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).To(BeNil())
		Expect(resp.Result).ToNot(BeEmpty())
	})

	ginkgo.It("reports a CEL compile error at the position of the offending token", func() {
		resp, err := run(Request{Language: LanguageCEL, Source: "1 +"})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).ToNot(BeNil())
		Expect(resp.Error.Message).ToNot(BeEmpty())
		Expect(resp.Error.Line).To(Equal(1))
		// A marker needs a 1-based column; 0 would silently land at the start.
		Expect(resp.Error.Column).To(BeNumerically(">", 0))
	})

	ginkgo.It("places the marker on the right line of a multi-line expression", func() {
		resp, err := run(Request{Language: LanguageCEL, Source: "1 +\n2 +\n%%%"})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).ToNot(BeNil())
		Expect(resp.Error.Line).To(Equal(3))
	})

	ginkgo.DescribeTable("reports parser validation at the offending source position",
		func(language Language, source string, line, column int) {
			resp, err := run(Request{Language: language, Source: source})
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Error).ToNot(BeNil())
			Expect(resp.Error.Message).ToNot(BeEmpty())
			Expect(resp.Error.Line).To(Equal(line))
			Expect(resp.Error.Column).To(Equal(column))
		},
		ginkgo.Entry("jsonpath", LanguageJSONPath, `$.items[?(`, 1, 11),
		ginkgo.Entry("javascript", LanguageJavaScript, "var value = 1;\nvalue + ;", 2, 9),
		ginkgo.Entry("go template parse", LanguageGoTemplate, "hello\n{{ nope .a }}\n", 2, 1),
		ginkgo.Entry("go template unclosed action", LanguageGoTemplate, "one\ntwo\n{{ if true }}", 3, 1),
	)

	ginkgo.It("reports a go-template execution failure at the offending action", func() {
		// The failure surfaces while executing, not while parsing, and text/template
		// reports the position only inside the message text.
		resp, err := run(Request{
			Language: LanguageGoTemplate,
			Source:   "x\n{{ .pod.metadata.name.nope }}\n",
			Input:    podInput,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).ToNot(BeNil())
		Expect(resp.Error.Line).To(Equal(2))
		// text/template picks the column inside the failing action itself; the
		// contract here is that it lands on that line, 1-based, not at 0.
		Expect(resp.Error.Column).To(BeNumerically(">", 0))
		Expect(resp.Error.Column).To(BeNumerically("<=", len("{{ .pod.metadata.name.nope }}")))
		// The position belongs in the marker, not repeated in its message, and
		// the empty template name reads as noise.
		Expect(resp.Error.Message).ToNot(ContainSubstring("template: "))
		Expect(resp.Error.Message).ToNot(ContainSubstring(`executing ""`))
		Expect(resp.Error.Message).To(ContainSubstring("can't evaluate field"))
	})

	ginkgo.It("reports a javascript runtime failure at the frame that raised it", func() {
		// otto keeps the position in the stack trace rather than the message, and
		// the innermost frame is where the editor should point.
		resp, err := run(Request{
			Language: LanguageJavaScript,
			Source:   "function fail() {\n  throw new Error('boom')\n}\nfail()",
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).ToNot(BeNil())
		Expect(resp.Error.Message).To(ContainSubstring("boom"))
		Expect(resp.Error.Line).To(Equal(2))
		Expect(resp.Error.Column).To(BeNumerically(">", 0))
	})

	ginkgo.It("reports an unknown identifier rather than evaluating to empty", func() {
		resp, err := run(Request{Language: LanguageCEL, Source: `nope.field`})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).ToNot(BeNil())
	})

	ginkgo.It("surfaces malformed input as an error instead of an empty environment", func() {
		resp, err := run(Request{Language: LanguageCEL, Source: "a", Input: "{not: [valid"})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).ToNot(BeNil())
		Expect(resp.Error.Message).To(ContainSubstring("input:"))
	})

	ginkgo.It("rejects an unknown language", func() {
		resp, err := run(Request{Language: "klingon", Source: "x"})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error.Message).To(ContainSubstring("unknown language"))
	})

	ginkgo.It("keeps the native value so the playground can show its type", func() {
		resp, err := run(Request{Language: LanguageCEL, Source: `[1, 2, 3]`})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).To(BeNil())
		Expect(resp.Result).To(Equal("[1,2,3]"))
		Expect(resp.Type).ToNot(BeEmpty())
	})
})

var _ = ginkgo.Describe("the playground API", func() {
	var mux *http.ServeMux

	ginkgo.BeforeEach(func() {
		handler, err := NewHandler(Options{})
		Expect(err).ToNot(HaveOccurred())
		mux = handler.Mux()
	})

	post := func(path string, body any) *httptest.ResponseRecorder {
		encoded, err := json.Marshal(body)
		Expect(err).ToNot(HaveOccurred())
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, bytes.NewReader(encoded)))
		return rec
	}

	ginkgo.It("evaluates over HTTP", func() {
		rec := post("/api/eval", Request{Language: LanguageCEL, Source: `"a" + "b"`})
		Expect(rec.Code).To(Equal(http.StatusOK))

		var resp Response
		Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).To(Succeed())
		Expect(resp.Result).To(Equal("ab"))
	})

	ginkgo.It("returns an evaluation failure as a 200 with an error body", func() {
		// The request was well formed; the expression was not. The playground
		// needs the message and position, not a transport-level failure.
		rec := post("/api/eval", Request{Language: LanguageCEL, Source: "1 +"})
		Expect(rec.Code).To(Equal(http.StatusOK))

		var resp Response
		Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).To(Succeed())
		Expect(resp.Error).ToNot(BeNil())
	})

	ginkgo.It("rejects a malformed request body", func() {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/eval", bytes.NewReader([]byte("{"))))
		Expect(rec.Code).To(Equal(http.StatusBadRequest))
	})

	ginkgo.It("serves the spec the editor completes from", func() {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/spec", nil))
		Expect(rec.Code).To(Equal(http.StatusOK))

		var body struct {
			CEL struct {
				Functions []struct {
					Name string `json:"name"`
				} `json:"functions"`
			} `json:"cel"`
		}
		Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(Succeed())
		Expect(len(body.CEL.Functions)).To(BeNumerically(">", 100))
	})
})

// hostCELFunction is a stand-in for the shape a host actually registers --
// duty's `catalog.query` is a namespaced function taking a string -- without
// dragging a database into the spec.
func hostCELFunction() cel.EnvOption {
	return cel.Function("catalog.query",
		cel.Overload("catalog.query_string",
			[]*cel.Type{cel.StringType},
			cel.StringType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				return types.String("queried:" + args[0].Value().(string))
			}),
		),
	)
}

func hostOptions() Options {
	return Options{
		CelEnvs: func(context.Context) []cel.EnvOption {
			return []cel.EnvOption{hostCELFunction()}
		},
	}
}

var _ = ginkgo.Describe("a host's own language", func() {
	const query = `catalog.query("name=web")`

	ginkgo.It("evaluates a function supplied through CelEnvs", func() {
		resp, err := runWith(hostOptions(), Request{Language: LanguageCEL, Source: query})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).To(BeNil(), "unexpected error: %+v", resp.Error)
		Expect(resp.Result).To(Equal("queried:name=web"))
	})

	ginkgo.It("rejects the same expression without the option", func() {
		// The negative half matters: without it the spec above would still pass
		// if the function were somehow reaching the evaluator by another route.
		resp, err := run(Request{Language: LanguageCEL, Source: query})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).ToNot(BeNil())
	})

	ginkgo.It("compiles against the host's options, so its own function is not undeclared", func() {
		// The compile pass exists to give errors a source position, and it
		// builds its own environment. Miss the host's options there and its
		// functions fail this check before the evaluator ever sees them.
		resp, err := runWith(hostOptions(), Request{
			Language: LanguageCEL,
			Source:   query + ` + "!"`,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).To(BeNil(), "unexpected error: %+v", resp.Error)
	})

	ginkgo.It("exposes a template function to go templates", func() {
		resp, err := runWith(Options{
			Functions: func(context.Context) map[string]any {
				return map[string]any{"hostName": func() any { return "mission-control" }}
			},
		}, Request{Language: LanguageGoTemplate, Source: `{{ hostName }}`})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).To(BeNil(), "unexpected error: %+v", resp.Error)
		Expect(resp.Result).To(Equal("mission-control"))
	})

	ginkgo.It("catalogues the host's function, with its overload, in the spec", func() {
		handler, err := NewHandler(hostOptions())
		Expect(err).ToNot(HaveOccurred())

		rec := httptest.NewRecorder()
		handler.Mux().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/spec", nil))
		Expect(rec.Code).To(Equal(http.StatusOK))

		var body struct {
			CEL struct {
				Namespaces []string `json:"namespaces"`
				Functions  []struct {
					Name      string `json:"name"`
					Namespace string `json:"namespace"`
					Overloads []struct {
						Args   []string `json:"args"`
						Result string   `json:"result"`
					} `json:"overloads"`
				} `json:"functions"`
			} `json:"cel"`
		}
		Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(Succeed())

		var found bool
		for _, fn := range body.CEL.Functions {
			if fn.Name != "catalog.query" {
				continue
			}
			found = true
			Expect(fn.Namespace).To(Equal("catalog"))
			Expect(fn.Overloads).To(HaveLen(1))
			Expect(fn.Overloads[0].Args).To(Equal([]string{"string"}))
			Expect(fn.Overloads[0].Result).To(Equal("string"))
		}
		Expect(found).To(BeTrue(), "catalog.query missing from the spec")

		// The namespace is what the tokenizer colours `catalog.` by.
		Expect(body.CEL.Namespaces).To(ContainElement("catalog"))
	})
})

var _ = ginkgo.Describe("serving examples", func() {
	serve := func(options Options) *httptest.ResponseRecorder {
		handler, err := NewHandler(options)
		Expect(err).ToNot(HaveOccurred())
		rec := httptest.NewRecorder()
		handler.Mux().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/examples", nil))
		return rec
	}

	ginkgo.It("serves what the host configured", func() {
		rec := serve(Options{Examples: []Example{{
			Name:     "Healthy pods",
			Language: LanguageCEL,
			Source:   `pod.status.phase == "Running"`,
			Input:    "pod:\n  status:\n    phase: Running\n",
		}}})
		Expect(rec.Code).To(Equal(http.StatusOK))

		var examples []Example
		Expect(json.Unmarshal(rec.Body.Bytes(), &examples)).To(Succeed())
		Expect(examples).To(HaveLen(1))
		Expect(examples[0].Name).To(Equal("Healthy pods"))
		Expect(examples[0].Language).To(Equal(LanguageCEL))
	})

	ginkgo.It("serves an empty array rather than null when none are configured", func() {
		// The playground renders the response directly; a null would show an
		// empty picker with nothing to explain it.
		rec := serve(Options{})
		Expect(rec.Body.String()).To(ContainSubstring("[]"))

		var examples []Example
		Expect(json.Unmarshal(rec.Body.Bytes(), &examples)).To(Succeed())
		Expect(examples).To(BeEmpty())
	})
})

var _ = ginkgo.Describe("bounding an evaluation", func() {
	// A slow function rather than a pathological expression: text/template
	// refuses deep recursion on its own after ~0.1s, so a runaway template
	// would exercise that limit instead of this one.
	slowOptions := func(d time.Duration) Options {
		return Options{
			Timeout: 100 * time.Millisecond,
			Functions: func(context.Context) map[string]any {
				return map[string]any{"slow": func() any { time.Sleep(d); return "done" }}
			},
		}
	}

	ginkgo.It("answers with an error instead of waiting for the evaluation", func() {
		started := time.Now()
		resp, err := runWith(slowOptions(10*time.Second), Request{
			Language: LanguageGoTemplate,
			Source:   `{{ slow }}`,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).ToNot(BeNil())
		Expect(resp.Error.Message).To(ContainSubstring("evaluation exceeded 100ms"))

		// The point of the bound: the caller is answered long before the
		// evaluation finishes. It is still running -- gomplate has no
		// cancellation to call -- but nobody is waiting on it.
		Expect(time.Since(started)).To(BeNumerically("<", 5*time.Second))
	})

	ginkgo.It("returns the real result when the evaluation finishes in time", func() {
		resp, err := runWith(slowOptions(time.Millisecond), Request{
			Language: LanguageGoTemplate,
			Source:   `{{ slow }}`,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).To(BeNil(), "unexpected error: %+v", resp.Error)
		Expect(resp.Result).To(Equal("done"))
	})

	ginkgo.It("leaves a fast evaluation alone", func() {
		resp, err := runWith(Options{Timeout: 30 * time.Second}, Request{
			Language: LanguageCEL,
			Source:   `1 + 1`,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error).To(BeNil())
		Expect(resp.Result).To(Equal("2"))
	})
})
