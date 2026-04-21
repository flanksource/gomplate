package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/flanksource/gomplate/v3"
	"gopkg.in/yaml.v3"
)

func loadEnvironment(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read env file %q: %w", path, err)
	}

	var env map[string]any
	if err := json.Unmarshal(data, &env); err == nil && env != nil {
		return env, nil
	}
	if err := yaml.Unmarshal(data, &env); err == nil && env != nil {
		return env, nil
	}

	var generic any
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return nil, fmt.Errorf("parse env file %q as yaml/json: %w", path, err)
	}

	return map[string]any{"data": generic}, nil
}

func printValue(w io.Writer, value any) error {
	switch v := value.(type) {
	case string:
		_, err := fmt.Fprintln(w, v)
		return err
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			_, err = fmt.Fprintln(w, value)
			return err
		}
		_, err = fmt.Fprintln(w, string(b))
		return err
	}
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("ceval", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var file string
	var expr string
	fs.StringVar(&file, "f", "", "Path to env yaml/json file")
	fs.StringVar(&file, "file", "", "Path to env yaml/json file")
	fs.StringVar(&expr, "e", "", "CEL expression to evaluate")
	fs.StringVar(&expr, "expr", "", "CEL expression to evaluate")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if file == "" || expr == "" {
		_, _ = fmt.Fprintln(stderr, "usage: ceval -f env.yaml -e \"labels['a/b/c']\"")
		fs.PrintDefaults()
		return 2
	}

	env, err := loadEnvironment(file)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	out, err := gomplate.RunExpression(env, gomplate.Template{Expression: expr})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	if err := printValue(stdout, out); err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
