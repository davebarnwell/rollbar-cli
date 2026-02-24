package cmd

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

type updateGlobalsSnapshot struct {
	status                 string
	resolvedInVersion      string
	title                  string
	level                  string
	assignedUserID         int64
	clearAssignedUser      bool
	assignedTeamID         int64
	clearAssignedTeam      bool
	snoozeEnabled          bool
	snoozeExpirationSecond int
}

func snapshotUpdateGlobals() updateGlobalsSnapshot {
	return updateGlobalsSnapshot{
		status:                 itemsUpdateStatus,
		resolvedInVersion:      itemsUpdateResolvedInVersion,
		title:                  itemsUpdateTitle,
		level:                  itemsUpdateLevel,
		assignedUserID:         itemsUpdateAssignedUserID,
		clearAssignedUser:      itemsUpdateClearAssignedUser,
		assignedTeamID:         itemsUpdateAssignedTeamID,
		clearAssignedTeam:      itemsUpdateClearAssignedTeam,
		snoozeEnabled:          itemsUpdateSnoozeEnabled,
		snoozeExpirationSecond: itemsUpdateSnoozeExpirationSecond,
	}
}

func restoreUpdateGlobals(s updateGlobalsSnapshot) {
	itemsUpdateStatus = s.status
	itemsUpdateResolvedInVersion = s.resolvedInVersion
	itemsUpdateTitle = s.title
	itemsUpdateLevel = s.level
	itemsUpdateAssignedUserID = s.assignedUserID
	itemsUpdateClearAssignedUser = s.clearAssignedUser
	itemsUpdateAssignedTeamID = s.assignedTeamID
	itemsUpdateClearAssignedTeam = s.clearAssignedTeam
	itemsUpdateSnoozeEnabled = s.snoozeEnabled
	itemsUpdateSnoozeExpirationSecond = s.snoozeExpirationSecond
}

func newUpdateCmdForTest() *cobra.Command {
	cmd := &cobra.Command{Use: "update"}
	f := cmd.Flags()
	f.StringVar(&itemsUpdateStatus, "status", "", "")
	f.StringVar(&itemsUpdateResolvedInVersion, "resolved-in-version", "", "")
	f.StringVar(&itemsUpdateTitle, "title", "", "")
	f.StringVar(&itemsUpdateLevel, "level", "", "")
	f.Int64Var(&itemsUpdateAssignedUserID, "assigned-user-id", 0, "")
	f.BoolVar(&itemsUpdateClearAssignedUser, "clear-assigned-user", false, "")
	f.Int64Var(&itemsUpdateAssignedTeamID, "assigned-team-id", 0, "")
	f.BoolVar(&itemsUpdateClearAssignedTeam, "clear-assigned-team", false, "")
	f.BoolVar(&itemsUpdateSnoozeEnabled, "snooze-enabled", false, "")
	f.IntVar(&itemsUpdateSnoozeExpirationSecond, "snooze-expiration-seconds", 0, "")
	return cmd
}

func mustSetFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set flag %s: %v", name, err)
	}
}

func TestResolveItemIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		id      int64
		uuid    string
		wantID  int64
		wantUID string
		wantErr bool
	}{
		{name: "missing", wantErr: true},
		{name: "multiple sources", args: []string{"123"}, id: 456, wantErr: true},
		{name: "positional id", args: []string{"123"}, wantID: 123},
		{name: "positional uuid", args: []string{"abcd-efgh"}, wantUID: "abcd-efgh"},
		{name: "flag id", id: 999, wantID: 999},
		{name: "flag uuid", uuid: "u-1", wantUID: "u-1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotID, gotUUID, err := resolveItemIdentifier(tc.args, tc.id, tc.uuid)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotID != tc.wantID || gotUUID != tc.wantUID {
				t.Fatalf("got (%d, %q), want (%d, %q)", gotID, gotUUID, tc.wantID, tc.wantUID)
			}
		})
	}
}

func TestBuildUpdateBodySuccess(t *testing.T) {
	snap := snapshotUpdateGlobals()
	defer restoreUpdateGlobals(snap)

	cmd := newUpdateCmdForTest()
	mustSetFlag(t, cmd, "status", "resolved")
	mustSetFlag(t, cmd, "resolved-in-version", "aabbcc1")
	mustSetFlag(t, cmd, "title", "Checkout failure")
	mustSetFlag(t, cmd, "level", "error")
	mustSetFlag(t, cmd, "assigned-user-id", "321")
	mustSetFlag(t, cmd, "snooze-enabled", "true")
	mustSetFlag(t, cmd, "snooze-expiration-seconds", "3600")

	got, err := buildUpdateBody(cmd)
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

func TestBuildUpdateBodyClearAssignments(t *testing.T) {
	snap := snapshotUpdateGlobals()
	defer restoreUpdateGlobals(snap)

	cmd := newUpdateCmdForTest()
	mustSetFlag(t, cmd, "status", "active")
	mustSetFlag(t, cmd, "clear-assigned-user", "true")
	mustSetFlag(t, cmd, "clear-assigned-team", "true")

	got, err := buildUpdateBody(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["assigned_user_id"] != nil || got["assigned_team_id"] != nil {
		t.Fatalf("expected cleared assignment fields, got %#v", got)
	}
}

func TestBuildUpdateBodyValidationErrors(t *testing.T) {
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
			snap := snapshotUpdateGlobals()
			defer restoreUpdateGlobals(snap)

			cmd := newUpdateCmdForTest()
			for k, v := range tc.setFlags {
				mustSetFlag(t, cmd, k, v)
			}

			if _, err := buildUpdateBody(cmd); err == nil {
				t.Fatalf("expected validation error, got nil")
			}
		})
	}
}

func TestBuildUpdateBodyAssignmentConflictErrors(t *testing.T) {
	tests := []struct {
		name       string
		clearFlag  string
		idFlagName string
		idValue    string
	}{
		{name: "user conflict", clearFlag: "clear-assigned-user", idFlagName: "assigned-user-id", idValue: "42"},
		{name: "team conflict", clearFlag: "clear-assigned-team", idFlagName: "assigned-team-id", idValue: "99"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snap := snapshotUpdateGlobals()
			defer restoreUpdateGlobals(snap)

			cmd := newUpdateCmdForTest()
			mustSetFlag(t, cmd, "status", "active")
			mustSetFlag(t, cmd, tc.clearFlag, "true")
			mustSetFlag(t, cmd, tc.idFlagName, tc.idValue)

			if _, err := buildUpdateBody(cmd); err == nil {
				t.Fatalf("expected conflict error, got nil")
			}
		})
	}
}
