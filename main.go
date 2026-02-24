package main

import (
	"fmt"
	"os"

	"rollbar-cli/cmd"
)

func run() int {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
