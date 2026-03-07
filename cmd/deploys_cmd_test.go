package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeploysListCommandJSON(t *testing.T) {
	var gotPages []string
	var gotLimits []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPages = append(gotPages, r.URL.Query().Get("page"))
		gotLimits = append(gotLimits, r.URL.Query().Get("limit"))
		_, _ = w.Write([]byte(`{"err":0,"result":{"deploys":[{"id":123,"environment":"production","revision":"aabbcc1","status":"started"},{"id":124,"environment":"production","revision":"ddbbcc2","status":"succeeded"}]}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"deploys", "list",
		"--page", "2",
		"--limit", "1",
		"--json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if len(gotPages) != 1 || gotPages[0] != "2" {
		t.Fatalf("unexpected requested pages: %#v", gotPages)
	}
	if len(gotLimits) != 1 || gotLimits[0] != "1" {
		t.Fatalf("unexpected limits: %#v", gotLimits)
	}
	if !strings.Contains(out, "\"deploys\"") || strings.Contains(out, "\"ID\": 124") || !strings.Contains(out, "\"ID\": 123") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestDeploysGetCommandNDJSON(t *testing.T) {
	var gotPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"err":0,"result":{"deploy":{"id":123,"environment":"production","revision":"aabbcc1","status":"started"}}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"deploys", "get", "123",
		"--ndjson",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotPath != "/api/1/deploy/123" {
		t.Fatalf("unexpected request path: %s", gotPath)
	}
	if strings.Contains(out, "\"deploy\"") || !strings.Contains(out, "\"Status\":\"started\"") {
		t.Fatalf("unexpected ndjson output: %q", out)
	}
}

func TestDeploysCreateCommandJSON(t *testing.T) {
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/1/deploy" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"err":0,"result":{"deploy":{"id":123,"environment":"production","revision":"aabbcc1","status":"started"}}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"deploys", "create",
		"--environment", "production",
		"--revision", "aabbcc1",
		"--status", "started",
		"--comment", "Deploy started from CI",
		"--local-username", "ci-bot",
		"--rollbar-username", "alice",
		"--json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotBody["environment"] != "production" || gotBody["revision"] != "aabbcc1" || gotBody["rollbar_username"] != "alice" {
		t.Fatalf("unexpected request body: %#v", gotBody)
	}
	if !strings.Contains(out, "\"deploy\"") || !strings.Contains(out, "\"ID\": 123") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestDeploysUpdateCommandJSON(t *testing.T) {
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/1/deploy/123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"err":0,"result":{"deploy":{"id":123,"environment":"production","revision":"aabbcc1","status":"succeeded"}}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"deploys", "update", "123",
		"--status", "succeeded",
		"--json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if len(gotBody) != 1 || gotBody["status"] != "succeeded" {
		t.Fatalf("unexpected request body: %#v", gotBody)
	}
	if !strings.Contains(out, "\"Status\": \"succeeded\"") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestDeploysCommandValidationErrors(t *testing.T) {
	_, err := runCLIWithCapturedStdout(t,
		"deploys", "get",
		"--token", "tok",
	)
	if err == nil || !strings.Contains(err.Error(), "missing deploy identifier") {
		t.Fatalf("expected missing identifier error, got %v", err)
	}

	_, err = runCLIWithCapturedStdout(t,
		"deploys", "create",
		"--environment", "production",
		"--token", "tok",
	)
	if err == nil || !strings.Contains(err.Error(), "missing required flag: --revision") {
		t.Fatalf("expected missing revision error, got %v", err)
	}

	_, err = runCLIWithCapturedStdout(t,
		"deploys", "update", "123",
		"--status", "bogus",
		"--token", "tok",
	)
	if err == nil || !strings.Contains(err.Error(), "invalid deploy status") {
		t.Fatalf("expected invalid status error, got %v", err)
	}

	_, err = runCLIWithCapturedStdout(t,
		"deploys", "update", "123",
		"--token", "tok",
	)
	if err == nil || !strings.Contains(err.Error(), "missing required flag: --status") {
		t.Fatalf("expected missing status error, got %v", err)
	}
}
