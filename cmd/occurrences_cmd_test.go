package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type occurrencesGlobalsSnapshot struct {
	cfg cliConfig

	listItemID   int64
	listItemUUID string
	listPage     int
	listOutput   string
	listJSON     bool

	getID     int64
	getUUID   string
	getOutput string
	getJSON   bool
}

func snapshotOccurrencesGlobals() occurrencesGlobalsSnapshot {
	return occurrencesGlobalsSnapshot{
		cfg: cfg,

		listItemID:   occurrencesListItemID,
		listItemUUID: occurrencesListItemUUID,
		listPage:     occurrencesListPage,
		listOutput:   occurrencesListOutput,
		listJSON:     occurrencesListJSON,

		getID:     occurrencesGetID,
		getUUID:   occurrencesGetUUID,
		getOutput: occurrencesGetOutput,
		getJSON:   occurrencesGetJSON,
	}
}

func restoreOccurrencesGlobals(s occurrencesGlobalsSnapshot) {
	cfg = s.cfg

	occurrencesListItemID = s.listItemID
	occurrencesListItemUUID = s.listItemUUID
	occurrencesListPage = s.listPage
	occurrencesListOutput = s.listOutput
	occurrencesListJSON = s.listJSON

	occurrencesGetID = s.getID
	occurrencesGetUUID = s.getUUID
	occurrencesGetOutput = s.getOutput
	occurrencesGetJSON = s.getJSON
}

func runCLIWithCapturedStdout(t *testing.T, args ...string) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	rootCmd.SetArgs(args)
	runErr := Execute()

	_ = w.Close()
	os.Stdout = oldStdout
	out := <-done
	_ = r.Close()
	return out, runErr
}

func TestOccurrencesListCommandJSON(t *testing.T) {
	snap := snapshotOccurrencesGlobals()
	defer restoreOccurrencesGlobals(snap)

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
		"--timeout", (2 * time.Second).String(),
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotPath != "/api/1/item/42/instances" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotPage != "2" {
		t.Fatalf("unexpected page query: %q", gotPage)
	}
	if !strings.Contains(out, "\"instances\"") || !strings.Contains(out, "\"inst-1\"") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestOccurrencesListCommandTextPositionalIdentifier(t *testing.T) {
	snap := snapshotOccurrencesGlobals()
	defer restoreOccurrencesGlobals(snap)

	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"err":0,"result":{"instances":[{"id":501,"uuid":"inst-1","level":"error","environment":"production","timestamp":1700001000}]}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"occurrences", "list", "item-uuid-1",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotPath != "/api/1/item/item-uuid-1/instances" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "inst-1") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestOccurrencesGetCommandJSONByUUID(t *testing.T) {
	snap := snapshotOccurrencesGlobals()
	defer restoreOccurrencesGlobals(snap)

	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"err":0,"result":{"instance":{"id":501,"uuid":"inst-1","timestamp":1700001000}}}`))
	}))
	defer ts.Close()

	out, err := runCLIWithCapturedStdout(t,
		"occurrences", "get",
		"--uuid", "inst-1",
		"--json",
		"--token", "tok",
		"--base-url", ts.URL,
	)
	if err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}
	if gotPath != "/api/1/instance/inst-1" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if !strings.Contains(out, "\"instance\"") || !strings.Contains(out, "\"inst-1\"") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestOccurrencesGetCommandTextByIDUsingAlias(t *testing.T) {
	snap := snapshotOccurrencesGlobals()
	defer restoreOccurrencesGlobals(snap)

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
	t.Run("invalid output", func(t *testing.T) {
		snap := snapshotOccurrencesGlobals()
		defer restoreOccurrencesGlobals(snap)

		_, err := runCLIWithCapturedStdout(t,
			"occurrences", "list",
			"--output", "xml",
			"--token", "tok",
		)
		if err == nil || !strings.Contains(err.Error(), "invalid --output") {
			t.Fatalf("expected invalid output error, got %v", err)
		}
	})

	t.Run("missing identifier", func(t *testing.T) {
		snap := snapshotOccurrencesGlobals()
		defer restoreOccurrencesGlobals(snap)

		_, err := runCLIWithCapturedStdout(t,
			"occurrences", "list",
			"--token", "tok",
		)
		if err == nil || !strings.Contains(err.Error(), "missing item identifier") {
			t.Fatalf("expected missing identifier error, got %v", err)
		}
	})
}
