package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/davebarnwell/rollbar-cli/internal/rollbar"
)

func TestRenderEnvironmentsPlain(t *testing.T) {
	var buf bytes.Buffer
	err := renderEnvironmentsPlain(&buf, []rollbar.Environment{
		{ID: 3, ProjectID: 99, Name: "production"},
	}, EnvironmentRenderOptions{})
	if err != nil {
		t.Fatalf("renderEnvironmentsPlain() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ENVIRONMENT") || !strings.Contains(out, "production") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRenderEnvironmentsEmpty(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RenderEnvironments(nil); err != nil {
			t.Fatalf("RenderEnvironments() error = %v", err)
		}
	})
	if !strings.Contains(out, "No environments found") {
		t.Fatalf("unexpected output: %q", out)
	}
}
