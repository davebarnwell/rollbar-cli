package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestItemsListCommandJSON(t *testing.T) {
	var gotPath string
	var gotPage string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotPage = r.URL.Query().Get("page")
		_, _ = w.Write([]byte(`{"err":0,"result":{"items":[{"id":42,"counter":10,"title":"boom","level":"error","status":"active","environment":"production","last_occurrence_timestamp":1700000000}]}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"items", "list",
		"--page", "2",
		"--json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotPath != "/api/1/items" || gotPage != "2" {
		t.Fatalf("unexpected request: path=%s page=%s", gotPath, gotPage)
	}
	if !strings.Contains(out, "\"items\"") || !strings.Contains(out, "\"boom\"") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestItemsGetCommandJSONWithInstancesHasStableShape(t *testing.T) {
	var paths []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/api/1/item/42":
			_, _ = w.Write([]byte(`{"err":0,"result":{"item":{"id":42,"counter":10,"title":"boom","level":"error"}}}`))
		case "/api/1/item/42/instances":
			_, _ = w.Write([]byte(`{"err":0,"result":{"instances":[{"id":501,"uuid":"inst-1","timestamp":1700001000}]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"items", "get", "42",
		"--instances",
		"--json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected two requests, got %#v", paths)
	}
	if !strings.Contains(out, "\"item\"") || !strings.Contains(out, "\"instances\"") || !strings.Contains(out, "\"inst-1\"") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestItemsCommandValidationErrors(t *testing.T) {
	_, err := runCLIWithCapturedStdout(t,
		"items", "get",
		"--id", "-1",
		"--token", "tok",
	)
	if err == nil || !strings.Contains(err.Error(), "invalid item id") {
		t.Fatalf("expected invalid id error, got %v", err)
	}
}
