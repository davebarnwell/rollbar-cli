package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"rollbar-cli/internal/rollbar"
	"rollbar-cli/internal/ui"
)

var (
	itemsPage        int
	itemsStatus      string
	itemsEnvironment string
	itemsLevel       []string
	itemsOutput      string
	itemsJSON        bool

	itemsGetID     int64
	itemsGetUUID   string
	itemsGetOutput string
	itemsGetJSON   bool

	itemsUpdateID                     int64
	itemsUpdateUUID                   string
	itemsUpdateStatus                 string
	itemsUpdateResolvedInVersion      string
	itemsUpdateTitle                  string
	itemsUpdateLevel                  string
	itemsUpdateAssignedUserID         int64
	itemsUpdateClearAssignedUser      bool
	itemsUpdateAssignedTeamID         int64
	itemsUpdateClearAssignedTeam      bool
	itemsUpdateSnoozeEnabled          bool
	itemsUpdateSnoozeExpirationSecond int
	itemsUpdateOutput                 string
	itemsUpdateJSON                   bool
)

var itemsCmd = &cobra.Command{
	Use:   "items",
	Short: "Query Rollbar items",
}

var itemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List items in a Rollbar project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireToken(); err != nil {
			return err
		}

		if itemsJSON {
			itemsOutput = "json"
		}

		switch itemsOutput {
		case "json", "text":
		default:
			return fmt.Errorf("invalid --output %q (expected: text|json)", itemsOutput)
		}

		client := newRollbarClient()

		resp, err := client.ListItems(cmd.Context(), rollbar.ListItemsOptions{
			Page:        itemsPage,
			Status:      itemsStatus,
			Environment: itemsEnvironment,
			Level:       itemsLevel,
		})
		if err != nil {
			return err
		}

		if itemsOutput == "json" {
			return writeJSON(resp.Raw)
		}

		return ui.RenderItems(resp.Items)
	},
}

var itemsGetCmd = &cobra.Command{
	Use:   "get [id-or-uuid]",
	Short: "Get a single Rollbar item by ID or UUID",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireToken(); err != nil {
			return err
		}

		if itemsGetJSON {
			itemsGetOutput = "json"
		}

		switch itemsGetOutput {
		case "json", "text":
		default:
			return fmt.Errorf("invalid --output %q (expected: text|json)", itemsGetOutput)
		}

		id, uuid, err := resolveItemIdentifier(args, itemsGetID, itemsGetUUID)
		if err != nil {
			return err
		}

		client := newRollbarClient()

		var resp *rollbar.GetItemResponse
		if uuid != "" {
			resp, err = client.GetItemByUUID(cmd.Context(), uuid)
		} else {
			resp, err = client.GetItemByID(cmd.Context(), id)
		}
		if err != nil {
			return err
		}

		if itemsGetOutput == "json" {
			return writeJSON(resp.Raw)
		}

		return ui.RenderItem(resp.Item)
	},
}

var itemsUpdateCmd = &cobra.Command{
	Use:   "update [id-or-uuid]",
	Short: "Update a Rollbar item",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireToken(); err != nil {
			return err
		}

		if itemsUpdateJSON {
			itemsUpdateOutput = "json"
		}
		switch itemsUpdateOutput {
		case "json", "text":
		default:
			return fmt.Errorf("invalid --output %q (expected: text|json)", itemsUpdateOutput)
		}

		updateBody, err := buildUpdateBody(cmd)
		if err != nil {
			return err
		}

		id, uuid, err := resolveItemIdentifier(args, itemsUpdateID, itemsUpdateUUID)
		if err != nil {
			return err
		}

		client := newRollbarClient()

		if uuid != "" {
			getResp, err := client.GetItemByUUID(cmd.Context(), uuid)
			if err != nil {
				return err
			}
			if getResp.Item.ID <= 0 {
				return fmt.Errorf("could not resolve UUID %q to a valid item id", uuid)
			}
			id = getResp.Item.ID
		}

		updateResp, err := client.UpdateItemByID(cmd.Context(), id, updateBody)
		if err != nil {
			return err
		}

		if itemsUpdateOutput == "json" {
			return writeJSON(updateResp.Raw)
		}

		if updateResp.Item.ID > 0 {
			return ui.RenderItem(updateResp.Item)
		}

		getResp, err := client.GetItemByID(cmd.Context(), id)
		if err == nil && getResp.Item.ID > 0 {
			return ui.RenderItem(getResp.Item)
		}

		return writeStdoutf("Item %d updated.\n", id)
	},
}

func newRollbarClient() *rollbar.Client {
	return rollbar.NewClient(rollbar.Config{
		AccessToken: cfg.Token,
		BaseURL:     cfg.BaseURL,
		Timeout:     cfg.Timeout,
	})
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeStdoutf(format string, args ...any) error {
	_, err := fmt.Fprintf(os.Stdout, format, args...)
	return err
}

func buildUpdateBody(cmd *cobra.Command) (map[string]any, error) {
	body := make(map[string]any)

	if cmd.Flags().Changed("status") {
		status := strings.ToLower(strings.TrimSpace(itemsUpdateStatus))
		switch status {
		case "active", "resolved", "muted":
			body["status"] = status
		default:
			return nil, fmt.Errorf("invalid --status %q (expected: active|resolved|muted)", itemsUpdateStatus)
		}
	}

	if cmd.Flags().Changed("resolved-in-version") {
		if len(itemsUpdateResolvedInVersion) > 40 {
			return nil, fmt.Errorf("--resolved-in-version cannot exceed 40 characters")
		}
		body["resolved_in_version"] = itemsUpdateResolvedInVersion
	}

	if cmd.Flags().Changed("title") {
		titleLen := len(itemsUpdateTitle)
		if titleLen < 1 || titleLen > 255 {
			return nil, fmt.Errorf("--title must be between 1 and 255 characters")
		}
		body["title"] = itemsUpdateTitle
	}

	if cmd.Flags().Changed("level") {
		level := strings.ToLower(strings.TrimSpace(itemsUpdateLevel))
		switch level {
		case "critical", "error", "warning", "info", "debug":
			body["level"] = level
		default:
			return nil, fmt.Errorf("invalid --level %q (expected: critical|error|warning|info|debug)", itemsUpdateLevel)
		}
	}

	if itemsUpdateClearAssignedUser && cmd.Flags().Changed("assigned-user-id") {
		return nil, fmt.Errorf("use either --assigned-user-id or --clear-assigned-user, not both")
	}
	if itemsUpdateClearAssignedUser {
		body["assigned_user_id"] = nil
	}
	if cmd.Flags().Changed("assigned-user-id") {
		if itemsUpdateAssignedUserID <= 0 {
			return nil, fmt.Errorf("--assigned-user-id must be > 0")
		}
		body["assigned_user_id"] = itemsUpdateAssignedUserID
	}

	if itemsUpdateClearAssignedTeam && cmd.Flags().Changed("assigned-team-id") {
		return nil, fmt.Errorf("use either --assigned-team-id or --clear-assigned-team, not both")
	}
	if itemsUpdateClearAssignedTeam {
		body["assigned_team_id"] = nil
	}
	if cmd.Flags().Changed("assigned-team-id") {
		if itemsUpdateAssignedTeamID <= 0 {
			return nil, fmt.Errorf("--assigned-team-id must be > 0")
		}
		body["assigned_team_id"] = itemsUpdateAssignedTeamID
	}

	if cmd.Flags().Changed("snooze-enabled") {
		body["snooze_enabled"] = itemsUpdateSnoozeEnabled
	}
	if cmd.Flags().Changed("snooze-expiration-seconds") {
		if itemsUpdateSnoozeExpirationSecond <= 0 {
			return nil, fmt.Errorf("--snooze-expiration-seconds must be > 0")
		}
		body["snooze_expiration_in_seconds"] = itemsUpdateSnoozeExpirationSecond
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("no updates provided: set at least one update flag")
	}

	return body, nil
}

func resolveItemIdentifier(args []string, id int64, uuid string) (int64, string, error) {
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
		return 0, "", fmt.Errorf("missing item identifier: pass [id-or-uuid], --id, or --uuid")
	}
	if sources > 1 {
		return 0, "", fmt.Errorf("provide only one item identifier: [id-or-uuid], --id, or --uuid")
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
	rootCmd.AddCommand(itemsCmd)
	itemsCmd.AddCommand(itemsListCmd)
	itemsCmd.AddCommand(itemsGetCmd)
	itemsCmd.AddCommand(itemsUpdateCmd)

	itemsListCmd.Flags().IntVar(&itemsPage, "page", 1, "Page number")
	itemsListCmd.Flags().StringVar(&itemsStatus, "status", "", "Filter by item status (e.g. active, resolved)")
	itemsListCmd.Flags().StringVar(&itemsEnvironment, "environment", "", "Filter by environment")
	itemsListCmd.Flags().StringSliceVar(&itemsLevel, "level", nil, "Filter by level; pass multiple times for multiple levels")
	itemsListCmd.Flags().StringVarP(&itemsOutput, "output", "o", "text", "Output format: text|json")
	itemsListCmd.Flags().BoolVar(&itemsJSON, "json", false, "Shortcut for --output json")

	itemsGetCmd.Flags().Int64Var(&itemsGetID, "id", 0, "Item ID")
	itemsGetCmd.Flags().StringVar(&itemsGetUUID, "uuid", "", "Occurrence UUID")
	itemsGetCmd.Flags().StringVarP(&itemsGetOutput, "output", "o", "text", "Output format: text|json")
	itemsGetCmd.Flags().BoolVar(&itemsGetJSON, "json", false, "Shortcut for --output json")

	itemsUpdateCmd.Flags().Int64Var(&itemsUpdateID, "id", 0, "Item ID")
	itemsUpdateCmd.Flags().StringVar(&itemsUpdateUUID, "uuid", "", "Occurrence UUID (resolved to item id before update)")
	itemsUpdateCmd.Flags().StringVar(&itemsUpdateStatus, "status", "", "New status: active|resolved|muted")
	itemsUpdateCmd.Flags().StringVar(&itemsUpdateResolvedInVersion, "resolved-in-version", "", "Resolved version (max 40 chars)")
	itemsUpdateCmd.Flags().StringVar(&itemsUpdateTitle, "title", "", "New title (1-255 chars)")
	itemsUpdateCmd.Flags().StringVar(&itemsUpdateLevel, "level", "", "New level: critical|error|warning|info|debug")
	itemsUpdateCmd.Flags().Int64Var(&itemsUpdateAssignedUserID, "assigned-user-id", 0, "Assign to user ID")
	itemsUpdateCmd.Flags().BoolVar(&itemsUpdateClearAssignedUser, "clear-assigned-user", false, "Clear assigned user")
	itemsUpdateCmd.Flags().Int64Var(&itemsUpdateAssignedTeamID, "assigned-team-id", 0, "Assign to team ID")
	itemsUpdateCmd.Flags().BoolVar(&itemsUpdateClearAssignedTeam, "clear-assigned-team", false, "Clear assigned team")
	itemsUpdateCmd.Flags().BoolVar(&itemsUpdateSnoozeEnabled, "snooze-enabled", false, "Set snooze enabled state")
	itemsUpdateCmd.Flags().IntVar(&itemsUpdateSnoozeExpirationSecond, "snooze-expiration-seconds", 0, "Snooze expiration in seconds")
	itemsUpdateCmd.Flags().StringVarP(&itemsUpdateOutput, "output", "o", "text", "Output format: text|json")
	itemsUpdateCmd.Flags().BoolVar(&itemsUpdateJSON, "json", false, "Shortcut for --output json")
}
