package main

import (
	"os"
	"walgo/cmd"
)

func main() {
	run(os.Args)
}

// run is a testable version of main
func run(args []string) {
	// Store original args and restore after
	oldArgs := os.Args
	os.Args = args
	defer func() { os.Args = oldArgs }()

	cmd.Execute()
}
