//go:build ignore

package main

import (
	"log/slog"

	"github.com/phillezi/gob"
)

func init() {
	logger := slog.New(gob.NewPrettyHandler(nil))
	slog.SetDefault(logger)
}

func main() {
	// takes in options
	gob.New(gob.WithDefaultTarget("all")).Add(
		"all",
		// takes in options to customize
		// can also be chaned with .For(os, arch)
		// or .Matrix([]string{"linux"}, []string{"amd64", "arm64"})
		gob.Static(),
	).Add(
		"clean",
		gob.Clean(),
	).Run()
}
