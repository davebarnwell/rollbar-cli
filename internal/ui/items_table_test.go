package ui

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"rollbar-cli/internal/rollbar"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
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

	fn()
	_ = w.Close()
	os.Stdout = old
	out := <-done
	_ = r.Close()
	return out
}

func TestRenderItemsPlain(t *testing.T) {
	var buf bytes.Buffer
	err := renderItemsPlain(&buf, []rollbar.Item{{
		Counter:                 10,
		Level:                   "error",
		Status:                  "active",
		Environment:             "production",
		LastOccurrenceTimestamp: 1700000000,
		Title:                   "something broke",
	}}, ItemListRenderOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "COUNTER") || !strings.Contains(out, "something broke") {
		t.Fatalf("unexpected plain output: %q", out)
	}
}

func TestRenderItemsEmpty(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RenderItems(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No items found") {
		t.Fatalf("expected no-items message, got %q", out)
	}
}

func TestRenderItem(t *testing.T) {
	out := captureStdout(t, func() {
		_ = RenderItem(rollbar.Item{ID: 12, Counter: 34})
	})
	if !strings.Contains(out, "ID: 12") {
		t.Fatalf("missing ID in output: %q", out)
	}
	if !strings.Contains(out, "Title: -") || !strings.Contains(out, "Last Seen: -") {
		t.Fatalf("missing fallback values in output: %q", out)
	}
}

func TestRenderItemWithInstances(t *testing.T) {
	out := captureStdout(t, func() {
		_ = RenderItemWithInstancesOptions(
			rollbar.Item{ID: 77, Counter: 1, Title: "oops"},
			[]rollbar.ItemInstance{
				{
					ID:          88,
					UUID:        "instance-uuid",
					Level:       "error",
					Environment: "production",
					Timestamp:   1700000000,
					StackFrames: []rollbar.StackFrame{
						{Filename: "app/main.go", Line: 42, Method: "handler"},
					},
					Payload: map[string]any{
						"request": map[string]any{"url": "https://example.com/checkout"},
					},
				},
			},
			ItemDetailsRenderOptions{
				Payload: PayloadRenderOptions{Mode: "full"},
			},
		)
	})

	if !strings.Contains(out, "Instances: 1") {
		t.Fatalf("missing instances count: %q", out)
	}
	if !strings.Contains(out, "app/main.go:42 (handler)") {
		t.Fatalf("missing stack frame file/line: %q", out)
	}
	if !strings.Contains(out, "\"request\"") || !strings.Contains(out, "https://example.com/checkout") {
		t.Fatalf("missing payload details: %q", out)
	}
}

func TestHelpers(t *testing.T) {
	if got := formatUnix(0); got != "-" {
		t.Fatalf("formatUnix(0) = %q, want -", got)
	}
	if got := fallback(""); got != "-" {
		t.Fatalf("fallback(\"\") = %q, want -", got)
	}
	if got := min(2, 5); got != 2 {
		t.Fatalf("min(2,5) = %d, want 2", got)
	}
}

type failWriter struct{}

func (failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestRenderItemWriteError(t *testing.T) {
	if err := renderItem(failWriter{}, rollbar.Item{}); err == nil {
		t.Fatalf("expected write error, got nil")
	}
}

func TestRenderItemsPlainWriteError(t *testing.T) {
	if err := renderItemsPlain(failWriter{}, []rollbar.Item{{Counter: 1}}, ItemListRenderOptions{}); err == nil {
		t.Fatalf("expected write error, got nil")
	}
}

func TestModelToggleDetails(t *testing.T) {
	m := newTestModel()
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	updated := next.(model)
	if !updated.showDetails {
		t.Fatalf("expected details to be visible")
	}
}

func TestModelFetchOccurrences(t *testing.T) {
	m := newTestModel()
	m.interactions = &ItemListInteractions{
		FetchOccurrences: func(item rollbar.Item) ([]rollbar.ItemInstance, error) {
			return []rollbar.ItemInstance{{ID: 88, UUID: "occ-1"}}, nil
		},
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(model)
	if !updated.showOccurrences {
		t.Fatalf("expected occurrences panel to be visible")
	}
	if len(updated.occurrences) != 1 || updated.occurrences[0].UUID != "occ-1" {
		t.Fatalf("unexpected occurrences: %#v", updated.occurrences)
	}
}

func TestModelCopyAndResolveActions(t *testing.T) {
	copied := false
	m := newTestModel()
	m.interactions = &ItemListInteractions{
		CopyItemID: func(item rollbar.Item) error {
			copied = true
			return nil
		},
		ResolveItem: func(item rollbar.Item) (rollbar.Item, error) {
			item.Status = "resolved"
			return item, nil
		},
		MuteItem: func(item rollbar.Item) (rollbar.Item, error) {
			item.Status = "muted"
			return item, nil
		},
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	updated := next.(model)
	if !copied {
		t.Fatalf("expected copy callback to run")
	}
	if !strings.Contains(updated.statusMessage, "copied item id") {
		t.Fatalf("unexpected copy status: %q", updated.statusMessage)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated = next.(model)
	if updated.items[0].Status != "resolved" {
		t.Fatalf("expected resolved status, got %#v", updated.items[0])
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	updated = next.(model)
	if updated.items[0].Status != "muted" {
		t.Fatalf("expected muted status, got %#v", updated.items[0])
	}
}

func TestClipboardCommands(t *testing.T) {
	if got := clipboardCommands("darwin"); len(got) != 1 || got[0].name != "pbcopy" {
		t.Fatalf("unexpected darwin clipboard commands: %#v", got)
	}
	if got := clipboardCommands("windows"); len(got) != 1 || got[0].name != "clip" {
		t.Fatalf("unexpected windows clipboard commands: %#v", got)
	}
	got := clipboardCommands("linux")
	if len(got) < 3 || got[0].name != "wl-copy" || got[1].name != "xclip" || got[2].name != "xsel" {
		t.Fatalf("unexpected linux clipboard commands: %#v", got)
	}
}

func TestModelCopyItemIDUsesAvailableClipboardCommand(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "clipboard.txt")
	candidates := clipboardCommands(runtime.GOOS)
	if len(candidates) == 0 {
		t.Fatal("expected at least one clipboard command candidate")
	}
	candidateName := candidates[0].name

	m := newTestModel()
	m.lookPath = func(file string) (string, error) {
		if file == candidateName {
			return file, nil
		}
		return "", exec.ErrNotFound
	}
	m.command = func(name string, args ...string) *exec.Cmd {
		if name != candidateName {
			t.Fatalf("unexpected clipboard command %q", name)
		}
		return newClipboardHelperCommand(t, outputPath)
	}

	if err := m.copyItemID(m.items[0]); err != nil {
		t.Fatalf("copyItemID() error = %v", err)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := strings.TrimSpace(string(raw)); got != "12" {
		t.Fatalf("clipboard contents = %q, want %q", got, "12")
	}
}

func TestModelCopyItemIDErrorListsAttemptedCommands(t *testing.T) {
	m := newTestModel()
	m.lookPath = func(string) (string, error) {
		return "", exec.ErrNotFound
	}

	err := m.copyItemID(m.items[0])
	if err == nil {
		t.Fatal("expected copyItemID() error")
	}
	if !strings.Contains(err.Error(), runtime.GOOS) {
		t.Fatalf("expected OS in error, got %q", err)
	}
	for _, name := range clipboardCommandNames(clipboardCommands(runtime.GOOS)) {
		if !strings.Contains(err.Error(), name) {
			t.Fatalf("expected %q in error, got %q", name, err)
		}
	}
}

func TestHelperProcessClipboard(t *testing.T) {
	if os.Getenv("GO_WANT_CLIPBOARD_HELPER_PROCESS") != "1" {
		return
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "read stdin: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(os.Getenv("ROLLBAR_CLIPBOARD_OUTPUT"), data, 0o600); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "write output: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func newClipboardHelperCommand(t *testing.T, outputPath string) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcessClipboard", "--")
	cmd.Env = append(os.Environ(),
		"GO_WANT_CLIPBOARD_HELPER_PROCESS=1",
		"ROLLBAR_CLIPBOARD_OUTPUT="+outputPath,
	)
	return cmd
}

func newTestModel() model {
	items := []rollbar.Item{{
		ID:          12,
		Counter:     34,
		Status:      "active",
		Level:       "error",
		Environment: "production",
		Title:       "something broke",
	}}
	tbl := table.New(
		table.WithColumns([]table.Column{
			{Title: "ID", Width: 10},
			{Title: "Counter", Width: 9},
			{Title: "Level", Width: 8},
			{Title: "Status", Width: 10},
			{Title: "Environment", Width: 14},
			{Title: "Last Seen", Width: 19},
			{Title: "Title", Width: 56},
		}),
		table.WithRows([]table.Row{{
			"12", "34", "error", "active", "production", "-", "something broke",
		}}),
	)
	return model{
		table:    tbl,
		items:    items,
		lookPath: exec.LookPath,
		command:  exec.Command,
	}
}
