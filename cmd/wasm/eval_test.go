package main

import (
	"encoding/json"
	"testing"

	"github.com/flanksource/gomplate/v4/playground"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWASM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WASM bridge")
}

var _ = Describe("JSON evaluation bridge", func() {
	DescribeTable("preserves native evaluation results",
		func(request playground.Request, expected string) {
			input, err := json.Marshal(request)
			Expect(err).NotTo(HaveOccurred())
			var response playground.Response
			Expect(json.Unmarshal([]byte(evaluateJSON(string(input))), &response)).To(Succeed())
			Expect(response.Error).To(BeNil())
			Expect(response.Result).To(Equal(expected))
		},
		Entry("CEL", playground.Request{Language: "cel", Source: "count * 2", Input: "count: 3"}, "6"),
		Entry("template", playground.Request{Language: "gotemplate", Source: "[[ .name ]]", Input: "name: acme", LeftDelim: "[[", RightDelim: "]]"}, "acme"),
		Entry("JSONPath", playground.Request{Language: "jsonpath", Source: "$.name", Input: "name: acme"}, "acme"),
		Entry("JavaScript", playground.Request{Language: "javascript", Source: "1 + 2"}, "3"),
	)

	It("returns malformed requests as structured errors", func() {
		var response playground.Response
		Expect(json.Unmarshal([]byte(evaluateJSON("{")), &response)).To(Succeed())
		Expect(response.Error).NotTo(BeNil())
		Expect(response.Error.Message).To(ContainSubstring("unexpected end"))
	})

	It("preserves compiler positions", func() {
		var response playground.Response
		Expect(json.Unmarshal([]byte(evaluateJSON(`{"language":"cel","source":"missing"}`)), &response)).To(Succeed())
		Expect(response.Error).NotTo(BeNil())
		Expect(response.Error.Line).To(Equal(1))
		Expect(response.Error.Column).To(Equal(1))
	})
})
