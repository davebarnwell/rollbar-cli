package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/davebarnwell/rollbar-cli/internal/rollbar"
)

func TestRenderDeploysPlain(t *testing.T) {
	var buf bytes.Buffer
	err := renderDeploysPlain(&buf, []rollbar.Deploy{
		{ID: 123, Status: "succeeded", Environment: "production", Revision: "aabbcc1", Comment: "done"},
	}, DeployRenderOptions{})
	if err != nil {
		t.Fatalf("renderDeploysPlain() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "STATUS") || !strings.Contains(out, "production") || !strings.Contains(out, "aabbcc1") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRenderDeploy(t *testing.T) {
	var buf bytes.Buffer
	err := renderDeploy(&buf, rollbar.Deploy{
		ID:              123,
		ProjectID:       42,
		Environment:     "production",
		Revision:        "aabbcc1",
		Status:          "succeeded",
		Comment:         "done",
		LocalUsername:   "ci-bot",
		RollbarUsername: "alice",
		StartTime:       1700000000,
		FinishTime:      1700003600,
	})
	if err != nil {
		t.Fatalf("renderDeploy() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ID: 123") || !strings.Contains(out, "Project ID: 42") || !strings.Contains(out, "Rollbar Username: alice") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRenderDeploysEmpty(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RenderDeploys(nil); err != nil {
			t.Fatalf("RenderDeploys() error = %v", err)
		}
	})
	if !strings.Contains(out, "No deploys found") {
		t.Fatalf("unexpected output: %q", out)
	}
}
