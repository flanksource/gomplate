// Command hostplayground stands in for a host like mission-control: it serves
// the playground API with a function of its own registered through
// playground.Options, so the whole path -- catalogue, highlighting, completion,
// hover, evaluation -- can be exercised the way a host will exercise it.
//
//	go run ./hack/hostplayground -addr :8321
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	"github.com/flanksource/gomplate/v3/playground"
)

// catalogQuery imitates duty's `catalog.query`: namespaced, one typed overload,
// a binding that in the real thing would reach a database.
func catalogQuery() cel.EnvOption {
	return cel.Function("catalog.query",
		cel.Overload("catalog.query_string",
			[]*cel.Type{cel.StringType},
			cel.StringType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				return types.String("matched:" + fmt.Sprint(args[0].Value()))
			}),
		),
	)
}

func main() {
	addr := flag.String("addr", ":8321", "address to listen on")
	flag.Parse()

	handler, err := playground.NewHandler(playground.Options{
		Timeout: 5 * time.Second,
		CelEnvs: func(context.Context) []cel.EnvOption {
			return []cel.EnvOption{catalogQuery()}
		},
		Functions: func(context.Context) map[string]any {
			return map[string]any{"hostName": func() any { return "mission-control" }}
		},
		Examples: []playground.Example{{
			Name:     "Catalogue query",
			Language: playground.LanguageCEL,
			Source:   `catalog.query("health=unhealthy")`,
			Input:    "pod:\n  status:\n    phase: Running\n",
		}},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "hostplayground:", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           handler.Mux(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Printf("hostplayground: listening on %s\n", *addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "hostplayground:", err)
		os.Exit(1)
	}
}
