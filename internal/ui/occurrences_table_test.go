package ui

import (
	"bytes"
	"strings"
	"testing"

	"rollbar-cli/internal/rollbar"
)

func TestRenderOccurrencesPlain(t *testing.T) {
	var buf bytes.Buffer
	err := renderOccurrencesPlain(&buf, []rollbar.ItemInstance{{
		ID:          501,
		UUID:        "inst-1",
		Level:       "error",
		Environment: "production",
		Timestamp:   1700000000,
		StackFrames: []rollbar.StackFrame{{Filename: "app/main.go"}},
	}}, OccurrenceRenderOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "inst-1") {
		t.Fatalf("unexpected plain output: %q", out)
	}
}

func TestRenderOccurrencesEmpty(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RenderOccurrences(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No occurrences found") {
		t.Fatalf("expected no-occurrences message, got %q", out)
	}
}

func TestRenderOccurrence(t *testing.T) {
	out := captureStdout(t, func() {
		_ = RenderOccurrence(rollbar.ItemInstance{ID: 501, UUID: "inst-1"})
	})
	if !strings.Contains(out, "Instances: 1") {
		t.Fatalf("missing instance count in output: %q", out)
	}
	if !strings.Contains(out, "UUID: inst-1") {
		t.Fatalf("missing uuid in output: %q", out)
	}
}

func TestRenderOccurrencesPlainWriteError(t *testing.T) {
	if err := renderOccurrencesPlain(failWriter{}, []rollbar.ItemInstance{{ID: 1}}, OccurrenceRenderOptions{}); err == nil {
		t.Fatalf("expected write error, got nil")
	}
}
