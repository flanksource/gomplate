package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/flanksource/gomplate/v4/playground"
	playgroundui "github.com/flanksource/gomplate/v4/web/apps/playground"
	"github.com/flanksource/gomplate/v4/web/packages/lang"
)

func serveHandler(options playground.Options) (http.Handler, error) {
	spec, err := lang.Spec()
	if err != nil {
		return nil, err
	}
	options.Spec = spec
	api, err := playground.NewHandler(options)
	if err != nil {
		return nil, err
	}
	ui, err := playgroundui.Handler()
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.Handle("/api/", api.Mux())
	mux.Handle("/", ui)
	return mux, nil
}

func runServe(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("ceval serve", flag.ContinueOnError)
	flags.SetOutput(stderr)
	addr := flags.String("addr", "127.0.0.1:8321", "address to listen on")
	timeout := flags.Duration("timeout", 5*time.Second, "ceiling on one server evaluation")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *timeout <= 0 {
		fmt.Fprintln(stderr, "serve requires no positional arguments and a positive timeout")
		return 2
	}
	handler, err := serveHandler(playground.Options{Timeout: *timeout})
	if err != nil {
		fmt.Fprintln(stderr, "playground:", err)
		return 1
	}
	server := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	fmt.Fprintf(stdout, "playground: http://%s\n", *addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(stderr, "playground:", err)
		return 1
	}
	return 0
}
