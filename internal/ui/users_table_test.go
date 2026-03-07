package ui

import (
	"bytes"
	"os"
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
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.String()
	}()

	if err := RenderUsers(nil); err != nil {
		t.Fatalf("RenderUsers() error = %v", err)
	}
	_ = w.Close()
	out := <-done
	_ = r.Close()

	if !strings.Contains(out, "No users found") {
		t.Fatalf("unexpected output: %q", out)
	}
}
