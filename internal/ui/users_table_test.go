package ui

import (
	"bytes"
	"strings"
	"testing"

	"rollbar-cli/internal/rollbar"
)

func TestRenderUsersPlain(t *testing.T) {
	var buf bytes.Buffer
	err := renderUsersPlain(&buf, []rollbar.User{
		{ID: 7, Username: "alice", Email: "alice@example.com"},
	}, UserRenderOptions{})
	if err != nil {
		t.Fatalf("renderUsersPlain() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "USERNAME") || !strings.Contains(out, "alice@example.com") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRenderUser(t *testing.T) {
	var buf bytes.Buffer
	err := renderUser(&buf, rollbar.User{ID: 7, Username: "alice", Email: "alice@example.com"})
	if err != nil {
		t.Fatalf("renderUser() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ID: 7") || !strings.Contains(out, "Username: alice") || !strings.Contains(out, "Email: alice@example.com") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRenderUsersEmpty(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RenderUsers(nil); err != nil {
			t.Fatalf("RenderUsers() error = %v", err)
		}
	})
	if !strings.Contains(out, "No users found") {
		t.Fatalf("unexpected output: %q", out)
	}
}
