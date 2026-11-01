package releasepolicy

import (
	"fmt"
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
			prepare := strings.ReplaceAll(benchmarkWorkflowStep("Prepare base worktree"), "${{ github.event.pull_request.base.sha }}", "base")
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

	It("compares matching benchmarks across module major versions", func() {
		root, err := filepath.Abs(filepath.Join("..", ".."))
		Expect(err).NotTo(HaveOccurred())
		Expect(os.MkdirAll(filepath.Join(root, ".tmp"), 0o755)).To(Succeed())
		fixture, err := os.MkdirTemp(filepath.Join(root, ".tmp"), "bench-compare-")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, fixture)
		Expect(os.MkdirAll(filepath.Join(fixture, ".github/scripts"), 0o755)).To(Succeed())
		summary, err := os.ReadFile(filepath.Join(root, ".github/scripts/benchstat-summary.py"))
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(fixture, ".github/scripts/benchstat-summary.py"), summary, 0o644)).To(Succeed())
		for _, sample := range []struct {
			file, module string
			nanoseconds  int
		}{{"bench-base.txt", "v3", 100}, {"bench-head.txt", "v4", 125}} {
			results := fmt.Sprintf("goos: linux\ngoarch: amd64\npkg: github.com/flanksource/gomplate/%s\n", sample.module)
			results += strings.Repeat(fmt.Sprintf("BenchmarkUpgrade 100 %d ns/op 100 B/op 1 allocs/op\n", sample.nanoseconds), 10)
			Expect(os.WriteFile(filepath.Join(fixture, sample.file), []byte(results), 0o644)).To(Succeed())
		}
		install := exec.Command("go", "install", "golang.org/x/perf/cmd/benchstat@82a0b07e230d76fa1b3036c383d7a98172f87334")
		install.Env = append(os.Environ(), "GOWORK=off", "GOBIN="+filepath.Join(fixture, "bin"))
		output, err := install.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		compare := strings.SplitN(benchmarkWorkflowStep("Compare"), "\n\n", 2)[0]
		command := exec.Command("/bin/bash", "-e", "-c", compare)
		command.Dir = fixture
		command.Env = append(os.Environ(), "PATH="+filepath.Join(fixture, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
		output, err = command.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		report, err := os.ReadFile(filepath.Join(fixture, "bench-summary.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(report)).To(ContainSubstring("| `Upgrade` | sec/op | 100 ns/op | 125 ns/op | **+25.00%** |"))
	})
})

func benchmarkWorkflowStep(name string) string {
	GinkgoHelper()
	content, err := os.ReadFile(filepath.Join("..", "..", ".github/workflows/benchmark.yml"))
	Expect(err).NotTo(HaveOccurred())
	var workflow struct {
		Jobs map[string]struct {
			Steps []struct{ Name, Run string }
		}
	}
	Expect(yaml.Unmarshal(content, &workflow)).To(Succeed())
	for _, step := range workflow.Jobs["bench"].Steps {
		if step.Name == name {
			return step.Run
		}
	}
	Fail("Missing benchmark workflow step: " + name)
	return ""
}
