// Package playground evaluates expressions for the language playground, using
// the same entry points a gomplate caller uses so what the playground shows is
// what production does.
//
// A host embeds this to give its own authors a playground over its own
// language: Options carries the CEL options and template functions the host
// registers, so the catalogue, the highlighting and the evaluator all agree
// with what that binary can actually run.
package playground

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/cel-go/cel"
	"github.com/robertkrimen/otto"
	ottoparser "github.com/robertkrimen/otto/parser"
	"gopkg.in/yaml.v3"

	gomplate "github.com/flanksource/gomplate/v3"
	"github.com/flanksource/gomplate/v3/coll"
)

// Options configure a playground for one host.
//
// The two function fields are shaped as factories rather than plain slices to
// match how hosts already register: duty keeps
// `map[string]func(Context) cel.EnvOption`, because a function like
// `catalog.query` closes over the database handle it queries through.
type Options struct {
	// CelEnvs are layered onto gomplate's own CEL options, per evaluation.
	CelEnvs func(context.Context) []cel.EnvOption
	// Functions are exposed to both CEL and go templates. Note gomplate's
	// constraint: a CEL-visible entry must be a `func() any`; anything else
	// belongs in CelEnvs.
	Functions func(context.Context) map[string]any
	// Examples are the samples the playground offers to load.
	Examples []Example
	// Timeout bounds one evaluation. Zero means no bound.
	//
	// It bounds the *response*, not the work: gomplate honours no context
	// deadline while evaluating -- there is no cel.ContextEval and no deadline
	// check in RunTemplateContext -- so a runaway expression keeps its
	// goroutine after the caller has been answered. That is still the right
	// trade for a shared endpoint, where a hung request is the worse failure,
	// but it is not cancellation and should not be mistaken for it.
	Timeout time.Duration
}

// Example is one sample an author can load into the playground.
type Example struct {
	Name     string   `json:"name"`
	Language Language `json:"language"`
	Source   string   `json:"source"`
	Input    string   `json:"input"`
}

func (o Options) celEnvs(ctx context.Context) []cel.EnvOption {
	if o.CelEnvs == nil {
		return nil
	}
	return o.CelEnvs(ctx)
}

func (o Options) functions(ctx context.Context) map[string]any {
	if o.Functions == nil {
		return nil
	}
	return o.Functions(ctx)
}

// template returns the prototype every evaluation starts from: the host's
// extensions, with the language-specific source left to the caller.
func (o Options) template(ctx context.Context) gomplate.Template {
	return gomplate.Template{
		CelEnvs:   o.celEnvs(ctx),
		Functions: o.functions(ctx),
	}
}

// Language selects which evaluator to run.
type Language string

const (
	LanguageCEL        Language = "cel"
	LanguageGoTemplate Language = "gotemplate"
	LanguageJSONPath   Language = "jsonpath"
	LanguageJavaScript Language = "javascript"
)

// Request is one evaluation.
type Request struct {
	Language Language `json:"language"`
	Source   string   `json:"source"`
	// Input is the evaluation environment, as YAML or JSON. JSON is valid YAML,
	// so one parser covers both.
	Input string `json:"input,omitempty"`
	// LeftDelim and RightDelim override the go-template delimiters. Both must
	// be set together.
	LeftDelim  string `json:"leftDelim,omitempty"`
	RightDelim string `json:"rightDelim,omitempty"`
}

// Response is the result of an evaluation.
type Response struct {
	// Result is the value rendered as a string, as a gomplate caller sees it.
	Result string `json:"result"`
	// Value is the native result, so the playground can show typed JSON rather
	// than a stringified value.
	Value any `json:"value,omitempty"`
	// Type names the Go type of Value, which is what makes CEL's int/uint/
	// double distinction visible.
	Type string `json:"type,omitempty"`
	// Error is set when evaluation failed. Result is empty in that case.
	Error *EvalError `json:"error,omitempty"`
	// DurationMs is wall-clock evaluation time.
	DurationMs float64 `json:"durationMs"`
}

// EvalError carries a message and, where the compiler reports one, a source
// position so the editor can place a marker on the offending token.
type EvalError struct {
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// Evaluate runs one request. A failed evaluation is a populated Error in the
// response, not a Go error: the playground always has something to render. A
// Go error means the request itself was malformed.
func (h *Handler) Evaluate(ctx context.Context, req Request) (*Response, error) {
	environment, err := parseInput(req.Input)
	if err != nil {
		return &Response{Error: &EvalError{Message: fmt.Sprintf("input: %s", err)}}, nil
	}

	if strings.TrimSpace(req.Source) == "" {
		return &Response{}, nil
	}

	base := h.options.template(ctx)

	started := time.Now()
	value, evalErr := h.runBounded(req, environment, base)
	elapsed := float64(time.Since(started).Microseconds()) / 1000

	resp := &Response{DurationMs: elapsed}
	if evalErr != nil {
		resp.Error = evalErr
		return resp, nil
	}
	resp.Value = value
	resp.Type = fmt.Sprintf("%T", value)
	resp.Result = renderResult(value)
	return resp, nil
}

type evalOutcome struct {
	value any
	err   *EvalError
}

// runBounded answers within Options.Timeout. See the field's documentation for
// why the abandoned evaluation keeps running: gomplate has no cancellation to
// call, so the choice is between answering late and answering at all.
func (h *Handler) runBounded(
	req Request,
	environment map[string]any,
	base gomplate.Template,
) (any, *EvalError) {
	if h.options.Timeout <= 0 {
		return evaluate(req, environment, base)
	}

	done := make(chan evalOutcome, 1)
	go func() {
		value, err := evaluate(req, environment, base)
		done <- evalOutcome{value: value, err: err}
	}()

	timer := time.NewTimer(h.options.Timeout)
	defer timer.Stop()
	select {
	case outcome := <-done:
		return outcome.value, outcome.err
	case <-timer.C:
		return nil, &EvalError{
			Message: fmt.Sprintf("evaluation exceeded %s", h.options.Timeout),
		}
	}
}

func evaluate(req Request, environment map[string]any, base gomplate.Template) (any, *EvalError) {
	switch req.Language {
	case LanguageCEL:
		return evaluateCEL(req.Source, environment, base)

	case LanguageJSONPath:
		// RunTemplateContext declares a JSONPath field but never evaluates it,
		// so route to the implementation the `jsonpath` function uses rather
		// than silently returning nothing.
		out, err := coll.JSONPath(req.Source, environment)
		if err != nil {
			return nil, jsonPathParseError(req.Source, err)
		}
		return out, nil

	case LanguageGoTemplate:
		tpl := base
		tpl.Template = req.Source
		tpl.LeftDelim = req.LeftDelim
		tpl.RightDelim = req.RightDelim
		out, err := gomplate.RunTemplate(environment, tpl)
		if err != nil {
			return nil, goTemplateError(err)
		}
		return out, nil

	case LanguageJavaScript:
		if validationErr := validateJavaScript(req.Source); validationErr != nil {
			return nil, validationErr
		}
		tpl := base
		tpl.Javascript = req.Source
		out, err := gomplate.RunTemplate(environment, tpl)
		if err != nil {
			return nil, javaScriptRuntimeError(err)
		}
		return out, nil

	default:
		return nil, &EvalError{Message: fmt.Sprintf("unknown language %q", req.Language)}
	}
}

// text/template reports where it failed only inside the message text:
// `template: <name>:<line>: <what>` while parsing, and
// `template: <name>:<line>:<column>: executing "<name>" at <action>: <what>`
// while executing. The template is unnamed here, which is why the name group
// is allowed to be empty.
var goTemplatePosition = regexp.MustCompile(`^template: .*?:(\d+)(?::(\d+))?: `)

// The template name repeated inside an execution failure is empty here, so it
// reads as `executing "" at <.a.b>` -- noise in front of the part that says
// which action failed.
var goTemplateExecuting = regexp.MustCompile(`^executing "[^"]*" at `)

// otto keeps the position in the stack trace rather than the message, one
// `at [name (]<anonymous>:<line>:<column>[)]` frame per line, innermost first.
var javaScriptFrame = regexp.MustCompile(`\n\s*at (?:[^\s(]+ \()?<anonymous>:(\d+):(\d+)\)?`)

// goTemplateError lifts the position text/template buried in its message into
// a field the editor can place a marker from. An error raised before parsing --
// gomplate's own, or a function's -- carries no position and keeps its message
// whole.
func goTemplateError(err error) *EvalError {
	message := err.Error()
	match := goTemplatePosition.FindStringSubmatch(message)
	if match == nil {
		return &EvalError{Message: message}
	}
	line, convErr := strconv.Atoi(match[1])
	if convErr != nil {
		return &EvalError{Message: message}
	}
	// Parse failures report a line but no column; the start of the line is as
	// close as the marker can honestly get.
	column := 1
	if match[2] != "" {
		if column, convErr = strconv.Atoi(match[2]); convErr != nil {
			return &EvalError{Message: message}
		}
	}
	rest := goTemplateExecuting.ReplaceAllString(message[len(match[0]):], "at ")
	return &EvalError{Message: rest, Line: line, Column: column}
}

// javaScriptRuntimeError reads the innermost stack frame, which is where the
// throw happened rather than where the call chain started.
func javaScriptRuntimeError(err error) *EvalError {
	var runtimeErr *otto.Error
	if !errors.As(err, &runtimeErr) {
		return &EvalError{Message: err.Error()}
	}
	out := &EvalError{Message: runtimeErr.Error()}
	if frame := javaScriptFrame.FindStringSubmatch(runtimeErr.String()); frame != nil {
		line, lineErr := strconv.Atoi(frame[1])
		column, columnErr := strconv.Atoi(frame[2])
		if lineErr == nil && columnErr == nil {
			out.Line, out.Column = line, column
		}
	}
	return out
}

func jsonPathParseError(source string, parseErr error) *EvalError {
	message := parseErr.Error()
	withoutSource := strings.TrimSuffix(message, " in "+source)
	separator := strings.LastIndex(withoutSource, " at ")
	if separator < 0 {
		return &EvalError{Message: "jsonpath parser returned an unpositioned error: " + message}
	}
	offset, err := strconv.Atoi(withoutSource[separator+4:])
	if err != nil || offset < 1 {
		return &EvalError{Message: "jsonpath parser returned an invalid error position: " + message}
	}
	line, column := positionAtByteOffset(source, offset)
	return &EvalError{Message: message, Line: line, Column: column}
}

func positionAtByteOffset(source string, offset int) (int, int) {
	before := source
	if offset-1 < len(source) {
		before = source[:offset-1]
	}
	line := strings.Count(before, "\n") + 1
	if newline := strings.LastIndexByte(before, '\n'); newline >= 0 {
		before = before[newline+1:]
	}
	return line, utf8.RuneCountInString(before) + 1
}

func validateJavaScript(source string) *EvalError {
	if _, err := ottoparser.ParseFile(nil, "", source, 0); err != nil {
		var parseErrors *ottoparser.ErrorList
		if errors.As(err, &parseErrors) && len(*parseErrors) > 0 {
			first := (*parseErrors)[0]
			return &EvalError{
				Message: first.Message,
				Line:    first.Position.Line,
				Column:  first.Position.Column,
			}
		}
		return &EvalError{Message: "javascript parser returned an unpositioned error: " + err.Error()}
	}
	return nil
}

// evaluateCEL compiles before evaluating so compile errors carry a source
// position. RunExpression alone reports the message without one, and a marker
// without a position lands on line 1.
//
// The compile-check environment comes from CompileEnvOptions rather than
// GetCelEnv so it matches what the evaluator will build. Without the host's
// CelEnvs and Functions, its own functions report "undeclared reference" here
// and never reach the evaluator that would have run them perfectly well.
func evaluateCEL(source string, environment map[string]any, base gomplate.Template) (any, *EvalError) {
	tpl := base
	tpl.Expression = source

	env, err := cel.NewEnv(gomplate.CompileEnvOptions(environment, tpl)...)
	if err != nil {
		return nil, &EvalError{Message: err.Error()}
	}
	if _, issues := env.Compile(source); issues != nil && issues.Err() != nil {
		return nil, celIssueError(issues)
	}

	out, err := gomplate.RunExpression(environment, tpl)
	if err != nil {
		return nil, &EvalError{Message: err.Error()}
	}
	return out, nil
}

// celIssueError takes the first issue's position, which is where the editor
// should point.
func celIssueError(issues *cel.Issues) *EvalError {
	out := &EvalError{Message: issues.Err().Error()}
	if errs := issues.Errors(); len(errs) > 0 {
		out.Message = errs[0].Message
		out.Line = errs[0].Location.Line()
		// CEL columns are 0-based; Monaco's are 1-based.
		out.Column = errs[0].Location.Column() + 1
	}
	return out
}

// parseInput reads the environment. JSON is valid YAML, so one parser serves
// both, and an empty input is an empty environment rather than an error.
func parseInput(input string) (map[string]any, error) {
	if strings.TrimSpace(input) == "" {
		return map[string]any{}, nil
	}
	var environment map[string]any
	if err := yaml.Unmarshal([]byte(input), &environment); err != nil {
		return nil, err
	}
	if environment == nil {
		return map[string]any{}, nil
	}
	return environment, nil
}

// renderResult stringifies the way a gomplate caller sees a result: strings
// verbatim, structures as JSON.
func renderResult(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	}
	if encoded, err := json.Marshal(value); err == nil {
		return string(encoded)
	}
	return fmt.Sprintf("%v", value)
}
