package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flanksource/gomplate/v4/playground"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ceval serve")
}

var _ = Describe("embedded playground server", func() {
	DescribeTable("routes API, SPA and assets independently",
		func(path string, status int, contentType string) {
			handler, err := serveHandler(playground.Options{})
			Expect(err).NotTo(HaveOccurred())
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			Expect(response.Code).To(Equal(status))
			Expect(response.Header().Get("Content-Type")).To(ContainSubstring(contentType))
		},
		Entry("index", "/", http.StatusOK, "text/html"),
		Entry("SPA route", "/playground/cel", http.StatusOK, "text/html"),
		Entry("health API", "/api/health", http.StatusOK, "application/json"),
		Entry("missing API", "/api/missing", http.StatusNotFound, "text/plain"),
		Entry("missing asset", "/assets/missing.js", http.StatusNotFound, "text/plain"),
		Entry("WASM module", "/wasm/eval.wasm.gz", http.StatusOK, "application/gzip"),
		Entry("Go runtime", "/wasm/wasm_exec.js", http.StatusOK, "javascript"),
	)
})
