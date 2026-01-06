package main

import (
	"os"

	"github.com/selimozten/walgo/cmd"
)

func main() {
	if err := run(os.Args); err != nil {
		os.Exit(1)
	}
}

// run is a testable version of main
func run(args []string) error {
	// Store original args and restore after
	oldArgs := os.Args
	os.Args = args
	defer func() { os.Args = oldArgs }()

	return cmd.Execute()
}
