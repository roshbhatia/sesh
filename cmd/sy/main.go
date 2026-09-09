package main

import (
	"os"

	"github.com/roshbhatia/seshy/cmd"
	"github.com/roshbhatia/seshy/internal/exitcode"
)

func main() {
	err := cmd.Execute()
	cmd.Report(os.Stderr, err)
	os.Exit(exitcode.Of(err))
}
