package cmd

import "testing"

func TestResolveOccurrenceIdentifier(t *testing.T) {
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
			gotID, gotUUID, err := resolveOccurrenceIdentifier(tc.args, tc.id, tc.uuid)
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

func TestResolveOccurrenceListItemIdentifier(t *testing.T) {
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
		{name: "positional uuid", args: []string{"item-uuid"}, wantUID: "item-uuid"},
		{name: "flag id", id: 999, wantID: 999},
		{name: "flag uuid", uuid: "item-1", wantUID: "item-1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotID, gotUUID, err := resolveOccurrenceListItemIdentifier(tc.args, tc.id, tc.uuid)
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
