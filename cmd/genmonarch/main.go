// Command genmonarch generates the Monaco language definitions for the
// expression languages gomplate evaluates.
//
// Everything it writes is derived: the lexical rules come from cel-go's own
// ANTLR grammar and from text/template's lexer, and the function catalogue is
// read out of a live cel.Env and the FuncMap gomplate installs. Nothing here is
// a list maintained by hand alongside the code it describes.
//
//	go run ./cmd/genmonarch -out web/packages/lang/src/generated
//	go run ./cmd/genmonarch -out web/packages/lang/src/generated -check
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/flanksource/gomplate/v3/genmonarch"
)

func main() {
	out := flag.String("out", "web/packages/lang/src/generated", "directory to write the generated definitions into")
	check := flag.Bool("check", false, "verify the checked-in files are up to date instead of writing them; exits non-zero on drift")
	flag.Parse()

	if err := run(*out, *check); err != nil {
		fmt.Fprintln(os.Stderr, "genmonarch:", err)
		os.Exit(1)
	}
}

// referenceDocs are the Markdown references the conformance corpus draws its
// snippets from, relative to the repository root.
var referenceDocs = []string{"CEL.md", "GO_TEMPLATE.md", "README.md"}

func run(dir string, check bool) error {
	docs, err := readDocs(referenceDocs)
	if err != nil {
		return err
	}

	bundle, err := genmonarch.Build(docs)
	if err != nil {
		return err
	}
	files, err := genmonarch.Render(bundle)
	if err != nil {
		return err
	}

	if check {
		return verify(dir, files)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := removeStale(dir, files); err != nil {
		return err
	}
	for _, name := range sortedNames(files) {
		if err := os.WriteFile(filepath.Join(dir, name), files[name], 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}
	fmt.Printf("genmonarch: wrote %d files to %s (%d CEL functions, %d go-template functions, %d conformance cases)\n",
		len(files), dir, len(bundle.Spec.CEL.Functions), len(bundle.Spec.GoTemplate.Functions), len(bundle.Conformance))
	return nil
}

// readDocs loads the reference Markdown. A missing file is fatal: silently
// generating a thinner corpus would weaken the gate without saying so.
func readDocs(names []string) (map[string]string, error) {
	docs := map[string]string{}
	for _, name := range names {
		content, err := os.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("reading %s (run from the repository root): %w", name, err)
		}
		docs[name] = string(content)
	}
	return docs, nil
}

// verify reports the first file that differs, so CI fails loudly on drift
// rather than shipping a highlighter that disagrees with the evaluator.
func verify(dir string, files map[string][]byte) error {
	for _, name := range sortedNames(files) {
		existing, err := os.ReadFile(filepath.Join(dir, name))
		if os.IsNotExist(err) {
			return fmt.Errorf("%s has not been generated; run `make monarch`", name)
		}
		if err != nil {
			return err
		}
		if string(existing) != string(files[name]) {
			return fmt.Errorf("%s is out of date; run `make monarch`", name)
		}
	}

	stale, err := staleFiles(dir, files)
	if err != nil {
		return err
	}
	if len(stale) > 0 {
		return fmt.Errorf("%v are no longer generated; run `make monarch`", stale)
	}
	return nil
}

// removeStale deletes previously generated files that are no longer produced,
// so a renamed language does not leave a stale definition behind.
func removeStale(dir string, files map[string][]byte) error {
	stale, err := staleFiles(dir, files)
	if err != nil {
		return err
	}
	for _, name := range stale {
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

func staleFiles(dir string, files map[string][]byte) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var stale []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if _, generated := files[entry.Name()]; !generated {
			stale = append(stale, entry.Name())
		}
	}
	sort.Strings(stale)
	return stale, nil
}

func sortedNames(files map[string][]byte) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
