package releasepolicy

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("benchmark preparation", func() {
	DescribeTable("adapts copied benchmarks without requiring ripgrep",
		func(module, expectedImport string) {
			root, err := filepath.Abs(filepath.Join("..", ".."))
			Expect(err).NotTo(HaveOccurred())
			content, err := os.ReadFile(filepath.Join(root, ".github/workflows/benchmark.yml"))
			Expect(err).NotTo(HaveOccurred())
			var workflow struct {
				Jobs map[string]struct {
					Steps []struct{ Name, Run string }
				}
			}
			Expect(yaml.Unmarshal(content, &workflow)).To(Succeed())
			var prepare string
			for _, step := range workflow.Jobs["bench"].Steps {
				if step.Name == "Prepare base worktree" {
					prepare = strings.ReplaceAll(step.Run, "${{ github.event.pull_request.base.sha }}", "base")
				}
			}
			Expect(prepare).NotTo(BeEmpty())
			Expect(os.MkdirAll(filepath.Join(root, ".tmp"), 0o755)).To(Succeed())
			fixture, err := os.MkdirTemp(filepath.Join(root, ".tmp"), "bench-prepare-")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(os.RemoveAll, fixture)
			for _, directory := range []string{"bin", ".tmp/bench/base", ".github"} {
				Expect(os.MkdirAll(filepath.Join(fixture, directory), 0o755)).To(Succeed())
			}
			patch, err := os.ReadFile(filepath.Join(root, ".github/bench-v3.patch"))
			Expect(err).NotTo(HaveOccurred())
			files := map[string][]byte{
				"bin/git":                      []byte("#!/bin/sh\nexit 0\n"),
				"bin/rg":                       []byte("#!/bin/sh\nexit 127\n"),
				".tmp/bench/base/go.mod":       []byte("module " + module + "\n"),
				"serialize_bench_test.go":      []byte("package gomplate\n"),
				"run_expression_bench_test.go": []byte("package gomplate\nimport (\n\"cel.dev/cel-go/cel\"\n\"cel.dev/cel-go/common/types/ref\"\n)\n"),
				".github/bench-v3.patch":       patch,
			}
			for name, data := range files {
				Expect(os.WriteFile(filepath.Join(fixture, name), data, 0o755)).To(Succeed())
			}
			command := exec.Command("/bin/bash", "-e", "-c", prepare)
			command.Dir = fixture
			command.Env = append(os.Environ(), "GOWORK=off", "PATH="+strings.Join([]string{
				filepath.Join(fixture, "bin"), filepath.Join(runtime.GOROOT(), "bin"), "/usr/bin", "/bin",
			}, string(os.PathListSeparator)))
			output, err := command.CombinedOutput()
			if expectedImport == "" {
				Expect(err).To(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("Unsupported benchmark base module: " + module))
				return
			}
			Expect(err).NotTo(HaveOccurred(), string(output))
			benchmark, err := os.ReadFile(filepath.Join(fixture, ".tmp/bench/base/run_expression_bench_test.go"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(benchmark)).To(ContainSubstring("\"" + expectedImport + "/cel\""))
			Expect(string(benchmark)).To(ContainSubstring("\"" + expectedImport + "/common/types/ref\""))
		},
		Entry("v3 base", "github.com/flanksource/gomplate/v3", "github.com/google/cel-go"),
		Entry("v4 base", "github.com/flanksource/gomplate/v4", "cel.dev/cel-go"),
		Entry("unsupported base", "example.com/unexpected", ""),
	)
})
