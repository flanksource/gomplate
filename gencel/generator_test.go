package gencel

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CEL Generator Suite")
}

var _ = Describe("CEL imports", func() {
	It("uses the canonical module path in generated bindings and exports", func() {
		dir := GinkgoT().TempDir()
		pkg := &Package{name: "example", baseDir: dir}
		file := &File{
			pkg: pkg, file: &ast.File{}, name: "example.go", path: filepath.Join(dir, "example.go"),
			decls: []FuncDecl{{
				Name: "Echo", RecvType: "Example",
				Args: []ast.Expr{ast.NewIdent("string")}, ReturnTypes: []ast.Expr{ast.NewIdent("string")},
			}},
		}
		generator := &Generator{pkg: pkg}
		generator.generateFile(file)
		generator.generateExports()

		for filename, want := range map[string][]string{
			"example_gen.go":     {"cel.dev/cel-go/cel", "cel.dev/cel-go/common/types", "cel.dev/cel-go/common/types/ref"},
			"cel_gen_exports.go": {"log", "cel.dev/cel-go/cel", "cel.dev/cel-go/common/types/ref"},
		} {
			generated, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, filename), nil, parser.ImportsOnly)
			Expect(err).NotTo(HaveOccurred())
			var imports []string
			for _, spec := range generated.Imports {
				path, err := strconv.Unquote(spec.Path.Value)
				Expect(err).NotTo(HaveOccurred())
				imports = append(imports, path)
			}
			Expect(imports).To(ConsistOf(want), filename)
		}
	})
})
