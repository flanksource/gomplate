package gencel

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

type Package struct {
	name  string
	defs  map[*ast.Ident]types.Object
	files []*File
}

type Generator struct {
	pkg *Package
	buf bytes.Buffer // Accumulated output.
}

// ParsePkg loads and adds the Go packages named by the given patterns.
func (t *Generator) ParsePkg(patterns ...string) {
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}

	t.addPackage(pkgs[0])
}

func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &File{
			file: file,
			pkg:  g.pkg,
			path: pkg.GoFiles[i],
		}
	}
}

func (g *Generator) Generate() {
	for _, file := range g.pkg.files {
		g.generateFile(file)
	}
}

// generateFile will generate the cel functions of the given file
// and write it out to a new file alongside the given file.
// The new file will be suffixed with "_gen.go"
func (g *Generator) generateFile(file *File) {
	if strings.HasSuffix(file.path, "_gen.go") {
		log.Println("Ignoring generation of gen file")
		return
	}

	g.clearBuf()

	// Print the header and package clause.
	g.printf("// Code generated by \"gencel\";\n// DO NOT EDIT.\n")
	g.printf("\n")
	g.printf("package %s", g.pkg.name)
	g.printf("\n")
	g.printf("import %s\n", `"github.com/google/cel-go/common/types"`)
	// g.printf("import %s\n", `"time"`)
	g.printf("import %s\n", `"github.com/google/cel-go/cel"`)
	g.printf("import %s\n", `"github.com/google/cel-go/common/types/ref"`)
	g.printf("\n")

	ast.Inspect(file.file, file.visitor)

	fileName := filepath.Base(file.path)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))

	if len(file.decls) == 0 {
		return
	}

	for _, decl := range file.decls {
		v := funcDefTemplateView{
			ParentFileName: fileName,
			FnName:         decl.Name,
			Args:           getCelArgs(decl.Args),
			ReturnTypes:    getCelArgs(decl.ReturnTypes),
			RecvType:       decl.RecvType,
		}

		g.render(v)
	}

	// Write to file.
	outputName := filepath.Join(filepath.Dir(file.path), fmt.Sprintf("%s_gen.go", fileName))

	log.Printf("Writing to [%s]", outputName)
	err := os.WriteFile(outputName, g.format(), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Generator) clearBuf() {
	g.buf = bytes.Buffer{}
}

func (g *Generator) printf(format string, args ...interface{}) {
	_, err := fmt.Fprintf(&g.buf, format, args...)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Generator) render(model interface{}) {
	t := template.New("main").Funcs(tplFuncs)

	t, err := t.Parse(funcDefTemplate)
	if err != nil {
		log.Fatal("instance template parse: ", err)
	}

	t, err = t.Parse(funcBodyTemplate)
	if err != nil {
		log.Fatal("instance template parse: ", err)
	}

	err = t.Execute(&g.buf, model)
	if err != nil {
		log.Fatal("Execute: ", err)
		return
	}
}

func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}
