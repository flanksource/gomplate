package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gomplate/v3/gencel"
)

var (
	input = flag.String("input", "", "input file")
	wait  = flag.Bool("wait", false, "should wait")
)

func main() {
	flag.Parse()
	log.Printf("Generating cel functions for %s\n", *input)

	// For debugging reason
	if *wait {
		time.Sleep(time.Second * 5)
	}

	fileName := filepath.Base(*input)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))

	g := gencel.Generator{}
	g.Execute(*input)
	g.Generate()

	src := g.Format()

	// Write to file.
	var dir = filepath.Dir(*input)
	outputName := filepath.Join(dir, fmt.Sprintf("%s_gen.go", fileName))

	log.Printf("Writing to [%s]", outputName)
	err := os.WriteFile(outputName, src, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
