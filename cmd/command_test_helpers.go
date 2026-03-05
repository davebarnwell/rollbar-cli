package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func runCLIWithCapturedStdout(t *testing.T, args ...string) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	cmd := newRootCmd()
	cmd.SetArgs(args)
	runErr := cmd.Execute()

	_ = w.Close()
	os.Stdout = oldStdout
	out := <-done
	_ = r.Close()
	return out, runErr
}
