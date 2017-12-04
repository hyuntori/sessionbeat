package main

import (
	"os"

	"github.com/hyuntori/sessionbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
