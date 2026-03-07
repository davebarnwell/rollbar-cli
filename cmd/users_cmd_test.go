package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUsersListCommandJSON(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"err":0,"result":{"users":[{"id":7,"username":"alice","email":"alice@example.com"}]}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"users", "list",
		"--json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotPath != "/api/1/users" {
		t.Fatalf("unexpected request path: %s", gotPath)
	}
	if !strings.Contains(out, "\"users\"") || !strings.Contains(out, "\"alice\"") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestUsersListCommandValidationErrors(t *testing.T) {
	_, err := runCLIWithCapturedStdout(t,
		"users", "list",
		"--token", "tok",
		"--output", "bogus",
	)
	if err == nil || !strings.Contains(err.Error(), "invalid --output") {
		t.Fatalf("expected invalid output error, got %v", err)
	}
}

func TestUsersGetCommandJSON(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"err":0,"result":{"id":7,"username":"alice","email":"alice@example.com"}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"users", "get", "7",
		"--json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotPath != "/api/1/user/7" {
		t.Fatalf("unexpected request path: %s", gotPath)
	}
	if !strings.Contains(out, "\"user\"") || !strings.Contains(out, "\"alice\"") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestUsersGetCommandNDJSON(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"err":0,"result":{"id":7,"username":"alice","email":"alice@example.com"}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"users", "get", "7",
		"--ndjson",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotPath != "/api/1/user/7" {
		t.Fatalf("unexpected request path: %s", gotPath)
	}
	if strings.Contains(out, "\"user\"") || !strings.Contains(out, "\"Username\":\"alice\"") {
		t.Fatalf("unexpected ndjson output: %q", out)
	}
}

func TestUsersGetCommandValidationErrors(t *testing.T) {
	_, err := runCLIWithCapturedStdout(t,
		"users", "get",
		"--token", "tok",
	)
	if err == nil || !strings.Contains(err.Error(), "missing user identifier") {
		t.Fatalf("expected missing identifier error, got %v", err)
	}

	_, err = runCLIWithCapturedStdout(t,
		"users", "get", "abc",
		"--token", "tok",
	)
	if err == nil || !strings.Contains(err.Error(), "invalid user id") {
		t.Fatalf("expected invalid id error, got %v", err)
	}
}
