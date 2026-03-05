package cmd

import (
	"strings"
	"testing"
)

func TestCompletionRequiresShellArgument(t *testing.T) {
	_, err := runCLIWithCapturedStdout(t, "completion")
	if err == nil || !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Fatalf("expected arg validation error, got %v", err)
	}
}
