package main

import (
	"os"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"rollbar-cli", "--help"}
	if code := run(); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestRunFailure(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"rollbar-cli", "definitely-not-a-command"}
	if code := run(); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}
