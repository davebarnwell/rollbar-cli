package cmd

import (
	"reflect"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func newUpdateCmdForTest(opts *itemsUpdateOptions) *cobra.Command {
	cmd := &cobra.Command{Use: "update"}
	f := cmd.Flags()
	f.StringVar(&opts.Status, "status", "", "")
	f.StringVar(&opts.ResolvedInVersion, "resolved-in-version", "", "")
	f.StringVar(&opts.Title, "title", "", "")
	f.StringVar(&opts.Level, "level", "", "")
	f.Int64Var(&opts.AssignedUserID, "assigned-user-id", 0, "")
	f.BoolVar(&opts.ClearAssignedUser, "clear-assigned-user", false, "")
	f.Int64Var(&opts.AssignedTeamID, "assigned-team-id", 0, "")
	f.BoolVar(&opts.ClearAssignedTeam, "clear-assigned-team", false, "")
	f.BoolVar(&opts.SnoozeEnabled, "snooze-enabled", false, "")
	f.IntVar(&opts.SnoozeExpirationSeconds, "snooze-expiration-seconds", 0, "")
	return cmd
}

func mustSetFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set flag %s: %v", name, err)
	}
}

func TestBuildItemUpdateBodySuccess(t *testing.T) {
	opts := &itemsUpdateOptions{}
	cmd := newUpdateCmdForTest(opts)
	mustSetFlag(t, cmd, "status", "resolved")
	mustSetFlag(t, cmd, "resolved-in-version", "aabbcc1")
	mustSetFlag(t, cmd, "title", "Checkout failure")
	mustSetFlag(t, cmd, "level", "error")
	mustSetFlag(t, cmd, "assigned-user-id", "321")
	mustSetFlag(t, cmd, "snooze-enabled", "true")
	mustSetFlag(t, cmd, "snooze-expiration-seconds", "3600")

	got, err := buildItemUpdateBody(cmd, *opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]any{
		"status":                       "resolved",
		"resolved_in_version":          "aabbcc1",
		"title":                        "Checkout failure",
		"level":                        "error",
		"assigned_user_id":             int64(321),
		"snooze_enabled":               true,
		"snooze_expiration_in_seconds": 3600,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected body\n got: %#v\nwant: %#v", got, want)
	}
}

func TestBuildItemUpdateBodyValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		setFlags map[string]string
	}{
		{name: "no updates"},
		{name: "bad status", setFlags: map[string]string{"status": "bad"}},
		{name: "bad level", setFlags: map[string]string{"level": "bad"}},
		{name: "title too short", setFlags: map[string]string{"title": ""}},
		{name: "resolved too long", setFlags: map[string]string{"resolved-in-version": "12345678901234567890123456789012345678901"}},
		{name: "assigned user invalid", setFlags: map[string]string{"assigned-user-id": "0"}},
		{name: "assigned team invalid", setFlags: map[string]string{"assigned-team-id": "0"}},
		{name: "snooze expiration invalid", setFlags: map[string]string{"snooze-expiration-seconds": "0"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := &itemsUpdateOptions{}
			cmd := newUpdateCmdForTest(opts)
			for k, v := range tc.setFlags {
				mustSetFlag(t, cmd, k, v)
			}
			if _, err := buildItemUpdateBody(cmd, *opts); err == nil {
				t.Fatalf("expected validation error, got nil")
			}
		})
	}
}

func TestParseItemTimeRange(t *testing.T) {
	opts := itemsListOptions{Last: 2 * time.Hour}
	since, until, err := parseItemTimeRange(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if since.IsZero() {
		t.Fatalf("expected since to be set")
	}
	if !until.IsZero() {
		t.Fatalf("expected until to be empty")
	}
}
