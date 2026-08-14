package playground

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flanksource/gomplate/v3/genmonarch"
)

// Handler serves the playground API.
//
// It carries no authentication of its own, deliberately. /api/eval runs
// arbitrary expressions with whatever Options grants them: in a host whose
// functions reach a database or a repository, that is arbitrary execution
// against real data. Mount it behind the same authorization as any other query
// endpoint. Handler.Mux satisfies http.Handler, so echo hosts wrap it with
// echo.WrapHandler inside an already-authenticated group.
type Handler struct {
	spec    genmonarch.Spec
	options Options
}

// NewHandler builds the API over a freshly extracted spec, so a running
// playground reflects the current binary rather than a stale generated file.
//
// The spec is extracted once, against context.Background(): Options.CelEnvs is
// a per-request factory because a function's *binding* closes over a request's
// context, but the declarations it registers -- the names, overloads and types
// the editor completes from -- are the same for every request.
func NewHandler(options Options) (*Handler, error) {
	celSpec, err := genmonarch.ExtractCEL(options.celEnvs(context.Background())...)
	if err != nil {
		return nil, err
	}
	goSpec, err := genmonarch.ExtractGoTemplate()
	if err != nil {
		return nil, err
	}
	return &Handler{
		spec:    genmonarch.Spec{CEL: celSpec, GoTemplate: goSpec},
		options: options,
	}, nil
}

// Mux returns the routes, ready to serve.
func (h *Handler) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/eval", h.handleEval)
	mux.HandleFunc("GET /api/spec", h.handleSpec)
	mux.HandleFunc("GET /api/examples", h.handleExamples)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

func (h *Handler) handleEval(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{
			Error: &EvalError{Message: fmt.Sprintf("malformed request: %s", err)},
		})
		return
	}

	resp, err := h.Evaluate(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Error: &EvalError{Message: err.Error()}})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleSpec(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.spec)
}

// handleExamples always writes an array, never null: the playground renders the
// response directly and a null would be an empty picker with no explanation.
func (h *Handler) handleExamples(w http.ResponseWriter, _ *http.Request) {
	examples := h.options.Examples
	if examples == nil {
		examples = []Example{}
	}
	writeJSON(w, http.StatusOK, examples)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// The status line is already written, so the only useful signal left is
		// the truncated body the client will fail to parse.
		fmt.Fprintf(w, "\n{\"error\":{\"message\":%q}}\n", err.Error())
	}
}
