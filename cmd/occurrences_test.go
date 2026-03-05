package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveOccurrenceIdentifier(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	cmd.Flags().Int64("id", 0, "")
	cmd.Flags().String("uuid", "", "")

	gotID, gotUUID, err := resolveOccurrenceIdentifier(cmd, []string{"123"}, occurrencesGetOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != 123 || gotUUID != "" {
		t.Fatalf("unexpected result: (%d, %q)", gotID, gotUUID)
	}

	if _, _, err := resolveOccurrenceIdentifier(cmd, []string{"-1"}, occurrencesGetOptions{}); err == nil {
		t.Fatalf("expected invalid negative id error")
	}
}

func TestResolveOccurrenceListItemIdentifier(t *testing.T) {
	cmd := &cobra.Command{Use: "list"}
	cmd.Flags().Int64("item-id", 0, "")
	cmd.Flags().String("item-uuid", "", "")

	gotID, gotUUID, err := resolveOccurrenceListItemIdentifier(cmd, []string{"item-uuid"}, occurrencesListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != 0 || gotUUID != "item-uuid" {
		t.Fatalf("unexpected result: (%d, %q)", gotID, gotUUID)
	}
}
