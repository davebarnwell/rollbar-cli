package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnvironmentsListCommandJSONPaginates(t *testing.T) {
	var gotPages []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPages = append(gotPages, r.URL.Query().Get("page"))
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"err":0,"result":{"environments":[
				{"id":1,"project_id":42,"environment":"production"},
				{"id":2,"project_id":42,"environment":"staging"},
				{"id":3,"project_id":42,"environment":"dev"},
				{"id":4,"project_id":42,"environment":"qa"},
				{"id":5,"project_id":42,"environment":"preview-1"},
				{"id":6,"project_id":42,"environment":"preview-2"},
				{"id":7,"project_id":42,"environment":"preview-3"},
				{"id":8,"project_id":42,"environment":"preview-4"},
				{"id":9,"project_id":42,"environment":"preview-5"},
				{"id":10,"project_id":42,"environment":"preview-6"},
				{"id":11,"project_id":42,"environment":"preview-7"},
				{"id":12,"project_id":42,"environment":"preview-8"},
				{"id":13,"project_id":42,"environment":"preview-9"},
				{"id":14,"project_id":42,"environment":"preview-10"},
				{"id":15,"project_id":42,"environment":"preview-11"},
				{"id":16,"project_id":42,"environment":"preview-12"},
				{"id":17,"project_id":42,"environment":"preview-13"},
				{"id":18,"project_id":42,"environment":"preview-14"},
				{"id":19,"project_id":42,"environment":"preview-15"},
				{"id":20,"project_id":42,"environment":"preview-16"}
			]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"err":0,"result":{"environments":[{"id":21,"project_id":42,"environment":"sandbox"}]}}`))
		case "3":
			_, _ = w.Write([]byte(`{"err":0,"result":{"environments":[]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"environments", "list",
		"--json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if len(gotPages) != 3 || gotPages[0] != "1" || gotPages[1] != "2" || gotPages[2] != "3" {
		t.Fatalf("unexpected requested pages: %#v", gotPages)
	}
	if !strings.Contains(out, "\"environments\"") || !strings.Contains(out, "\"Name\": \"sandbox\"") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestEnvironmentsListCommandRawJSONAggregatesPages(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = w.Write([]byte(`{"err":0,"result":{"environments":[{"id":1,"project_id":42,"environment":"production"}]}}`))
		case "2":
			_, _ = w.Write([]byte(`{"err":0,"result":{"environments":[]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"environments", "list",
		"--raw-json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if !strings.Contains(out, "\"pages\"") || !strings.Contains(out, "\"environment\": \"production\"") || !strings.Contains(out, "\"environments\": []") {
		t.Fatalf("unexpected raw-json output: %q", out)
	}
}
