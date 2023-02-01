package main

import (
	"flag"
	"log"
	"time"

	"github.com/flanksource/gomplate/v3/gencel"
)

var (
	wait = flag.Bool("wait", false, "should wait")
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}
	log.Printf("Generating cel functions for %s\n", args)

	// For debugging reason
	if *wait {
		time.Sleep(time.Second * 5)
	}

	g := gencel.Generator{}
	g.ParsePkg(args...)
	g.Generate()
}
