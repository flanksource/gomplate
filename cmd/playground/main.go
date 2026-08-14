// Command playground serves the evaluation API behind the language playground.
//
// The Vite dev server proxies /api to it, so the playground evaluates against
// the real gomplate engine rather than a reimplementation in the browser.
//
//	go run ./cmd/playground -addr :8321
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/flanksource/gomplate/v3/playground"
)

func main() {
	addr := flag.String("addr", ":8321", "address to listen on")
	timeout := flag.Duration("timeout", 5*time.Second, "ceiling on one evaluation")
	flag.Parse()

	// No CelEnvs or Functions: this binary serves gomplate's own language. A
	// host embeds the same package and supplies its own.
	handler, err := playground.NewHandler(playground.Options{Timeout: *timeout})
	if err != nil {
		fmt.Fprintln(os.Stderr, "playground:", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           handler.Mux(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("playground: listening on %s\n", *addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "playground:", err)
		os.Exit(1)
	}
}
