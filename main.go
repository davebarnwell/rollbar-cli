package main

import (
	"fmt"
	"os"

	"rollbar-cli/cmd"
)

func run() int {
	if err := cmd.Execute(); err != nil {
		if _, writeErr := fmt.Fprintln(os.Stderr, err); writeErr != nil {
			return 1
		}
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
