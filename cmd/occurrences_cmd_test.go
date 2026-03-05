package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOccurrencesListCommandJSON(t *testing.T) {
	var gotPath string
	var gotPage string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotPage = r.URL.Query().Get("page")
		_, _ = w.Write([]byte(`{"err":0,"result":{"instances":[{"id":501,"uuid":"inst-1","timestamp":1700001000}]}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"occurrences", "list",
		"--item-id", "42",
		"--page", "2",
		"--json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotPath != "/api/1/item/42/instances" || gotPage != "2" {
		t.Fatalf("unexpected request: path=%s page=%s", gotPath, gotPage)
	}
	if !strings.Contains(out, "\"occurrences\"") || !strings.Contains(out, "\"inst-1\"") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestOccurrencesGetCommandTextByIDUsingAlias(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"err":0,"result":{"instance":{"id":501,"uuid":"inst-1","level":"error","environment":"production","timestamp":1700001000}}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"occurences", "get", "501",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotPath != "/api/1/instance/501" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if !strings.Contains(out, "Instances: 1") || !strings.Contains(out, "UUID: inst-1") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestOccurrencesListCommandValidationErrors(t *testing.T) {
	_, err := runCLIWithCapturedStdout(t,
		"occurrences", "list",
		"--token", "tok",
	)
	if err == nil || !strings.Contains(err.Error(), "missing item identifier") {
		t.Fatalf("expected missing identifier error, got %v", err)
	}
}
