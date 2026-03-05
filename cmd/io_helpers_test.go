package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWriteStdoutf(t *testing.T) {
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

	if err := writeStdoutf("value=%d", 42); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	_ = w.Close()
	os.Stdout = oldStdout
	out := <-done
	_ = r.Close()

	if !strings.Contains(out, "value=42") {
		t.Fatalf("unexpected output: %q", out)
	}
}
