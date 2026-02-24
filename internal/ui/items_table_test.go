package ui

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"rollbar-cli/internal/rollbar"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
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

	fn()
	_ = w.Close()
	os.Stdout = old
	out := <-done
	_ = r.Close()
	return out
}

func TestRenderItemsPlain(t *testing.T) {
	var buf bytes.Buffer
	err := renderItemsPlain(&buf, []rollbar.Item{{
		Counter:                 10,
		Level:                   "error",
		Status:                  "active",
		Environment:             "production",
		LastOccurrenceTimestamp: 1700000000,
		Title:                   "something broke",
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "COUNTER") || !strings.Contains(out, "something broke") {
		t.Fatalf("unexpected plain output: %q", out)
	}
}

func TestRenderItemsEmpty(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RenderItems(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No items found") {
		t.Fatalf("expected no-items message, got %q", out)
	}
}

func TestRenderItem(t *testing.T) {
	out := captureStdout(t, func() {
		_ = RenderItem(rollbar.Item{ID: 12, Counter: 34})
	})
	if !strings.Contains(out, "ID: 12") {
		t.Fatalf("missing ID in output: %q", out)
	}
	if !strings.Contains(out, "Title: -") || !strings.Contains(out, "Last Seen: -") {
		t.Fatalf("missing fallback values in output: %q", out)
	}
}

func TestHelpers(t *testing.T) {
	if got := formatUnix(0); got != "-" {
		t.Fatalf("formatUnix(0) = %q, want -", got)
	}
	if got := fallback(""); got != "-" {
		t.Fatalf("fallback(\"\") = %q, want -", got)
	}
	if got := min(2, 5); got != 2 {
		t.Fatalf("min(2,5) = %d, want 2", got)
	}
}

type failWriter struct{}

func (failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestRenderItemWriteError(t *testing.T) {
	if err := renderItem(failWriter{}, rollbar.Item{}); err == nil {
		t.Fatalf("expected write error, got nil")
	}
}

func TestRenderItemsPlainWriteError(t *testing.T) {
	if err := renderItemsPlain(failWriter{}, []rollbar.Item{{Counter: 1}}); err == nil {
		t.Fatalf("expected write error, got nil")
	}
}
