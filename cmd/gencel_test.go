package main

import (
	"os"
	"testing"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gomplate/v3/gencel"
)

func TestGencel(t *testing.T) {
	wd, _ := os.Getwd()
	logger.Infof("WD: %s", wd)

	args := []string{"../funcs"}
	g := gencel.Generator{}
	g.ParsePkg(args...)
	g.Generate()
}
