package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"rollbar-cli/internal/rollbar"
	"rollbar-cli/internal/ui"
)

var (
	occurrencesListItemID   int64
	occurrencesListItemUUID string
	occurrencesListPage     int
	occurrencesListOutput   string
	occurrencesListJSON     bool

	occurrencesGetID     int64
	occurrencesGetUUID   string
	occurrencesGetOutput string
	occurrencesGetJSON   bool
)

var occurrencesCmd = &cobra.Command{
	Use:     "occurrences",
	Aliases: []string{"occurences"},
	Short:   "Query Rollbar occurrences",
}

var occurrencesListCmd = &cobra.Command{
	Use:   "list [item-id-or-uuid]",
	Short: "List occurrences for a Rollbar item",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireToken(); err != nil {
			return err
		}

		if occurrencesListJSON {
			occurrencesListOutput = "json"
		}

		switch occurrencesListOutput {
		case "json", "text":
		default:
			return fmt.Errorf("invalid --output %q (expected: text|json)", occurrencesListOutput)
		}

		itemID, itemUUID, err := resolveOccurrenceListItemIdentifier(args, occurrencesListItemID, occurrencesListItemUUID)
		if err != nil {
			return err
		}

		identifier := itemUUID
		if identifier == "" {
			identifier = strconv.FormatInt(itemID, 10)
		}

		client := newRollbarClient()
		resp, err := client.ListItemInstances(cmd.Context(), identifier, occurrencesListPage)
		if err != nil {
			return err
		}

		if occurrencesListOutput == "json" {
			return writeJSON(resp.Raw)
		}
		return ui.RenderOccurrences(resp.Instances)
	},
}

var occurrencesGetCmd = &cobra.Command{
	Use:   "get [id-or-uuid]",
	Short: "Get a single occurrence by ID or UUID",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireToken(); err != nil {
			return err
		}

		if occurrencesGetJSON {
			occurrencesGetOutput = "json"
		}

		switch occurrencesGetOutput {
		case "json", "text":
		default:
			return fmt.Errorf("invalid --output %q (expected: text|json)", occurrencesGetOutput)
		}

		id, uuid, err := resolveOccurrenceIdentifier(args, occurrencesGetID, occurrencesGetUUID)
		if err != nil {
			return err
		}

		client := newRollbarClient()
		var resp *rollbar.GetOccurrenceResponse
		if uuid != "" {
			resp, err = client.GetOccurrenceByUUID(cmd.Context(), uuid)
		} else {
			resp, err = client.GetOccurrenceByID(cmd.Context(), id)
		}
		if err != nil {
			return err
		}

		if occurrencesGetOutput == "json" {
			return writeJSON(resp.Raw)
		}
		return ui.RenderOccurrence(resp.Occurrence)
	},
}

func resolveOccurrenceIdentifier(args []string, id int64, uuid string) (int64, string, error) {
	var arg string
	if len(args) > 0 {
		arg = strings.TrimSpace(args[0])
	}
	uuid = strings.TrimSpace(uuid)

	sources := 0
	if arg != "" {
		sources++
	}
	if id > 0 {
		sources++
	}
	if uuid != "" {
		sources++
	}

	if sources == 0 {
		return 0, "", fmt.Errorf("missing occurrence identifier: pass [id-or-uuid], --id, or --uuid")
	}
	if sources > 1 {
		return 0, "", fmt.Errorf("provide only one occurrence identifier: [id-or-uuid], --id, or --uuid")
	}

	if arg != "" {
		if n, err := strconv.ParseInt(arg, 10, 64); err == nil && n > 0 {
			return n, "", nil
		}
		return 0, arg, nil
	}

	if id > 0 {
		return id, "", nil
	}

	return 0, uuid, nil
}

func resolveOccurrenceListItemIdentifier(args []string, id int64, uuid string) (int64, string, error) {
	var arg string
	if len(args) > 0 {
		arg = strings.TrimSpace(args[0])
	}
	uuid = strings.TrimSpace(uuid)

	sources := 0
	if arg != "" {
		sources++
	}
	if id > 0 {
		sources++
	}
	if uuid != "" {
		sources++
	}

	if sources == 0 {
		return 0, "", fmt.Errorf("missing item identifier: pass [item-id-or-uuid], --item-id, or --item-uuid")
	}
	if sources > 1 {
		return 0, "", fmt.Errorf("provide only one item identifier: [item-id-or-uuid], --item-id, or --item-uuid")
	}

	if arg != "" {
		if n, err := strconv.ParseInt(arg, 10, 64); err == nil && n > 0 {
			return n, "", nil
		}
		return 0, arg, nil
	}

	if id > 0 {
		return id, "", nil
	}

	return 0, uuid, nil
}

func init() {
	rootCmd.AddCommand(occurrencesCmd)
	occurrencesCmd.AddCommand(occurrencesListCmd)
	occurrencesCmd.AddCommand(occurrencesGetCmd)

	occurrencesListCmd.Flags().Int64Var(&occurrencesListItemID, "item-id", 0, "Item ID")
	occurrencesListCmd.Flags().StringVar(&occurrencesListItemUUID, "item-uuid", "", "Item UUID")
	occurrencesListCmd.Flags().IntVar(&occurrencesListPage, "page", 1, "Page number")
	occurrencesListCmd.Flags().StringVarP(&occurrencesListOutput, "output", "o", "text", "Output format: text|json")
	occurrencesListCmd.Flags().BoolVar(&occurrencesListJSON, "json", false, "Shortcut for --output json")

	occurrencesGetCmd.Flags().Int64Var(&occurrencesGetID, "id", 0, "Occurrence ID")
	occurrencesGetCmd.Flags().StringVar(&occurrencesGetUUID, "uuid", "", "Occurrence UUID")
	occurrencesGetCmd.Flags().StringVarP(&occurrencesGetOutput, "output", "o", "text", "Output format: text|json")
	occurrencesGetCmd.Flags().BoolVar(&occurrencesGetJSON, "json", false, "Shortcut for --output json")
}
