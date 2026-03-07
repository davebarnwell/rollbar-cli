package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/davebarnwell/rollbar-cli/internal/rollbar"
	"github.com/davebarnwell/rollbar-cli/internal/ui"
)

type usersListOptions struct {
	Output    string
	JSON      bool
	RawJSON   bool
	NDJSON    bool
	Fields    []string
	NoHeaders bool
}

type usersGetOptions struct {
	ID      int64
	Output  string
	JSON    bool
	RawJSON bool
	NDJSON  bool
}

type userListJSONOutput struct {
	Users []rollbar.User `json:"users"`
}

type userGetJSONOutput struct {
	User rollbar.User `json:"user"`
}

func newUsersCmd(cfg *cliConfig) *cobra.Command {
	var listOpts usersListOptions
	var getOpts usersGetOptions

	usersCmd := &cobra.Command{
		Use:   "users",
		Short: "Query Rollbar users",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List users in the Rollbar account",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			output, err := resolveOutputModeWithAliases(listOpts.Output, listOpts.JSON, listOpts.RawJSON, listOpts.NDJSON, outputText, outputJSON, outputRawJSON, outputNDJSON)
			if err != nil {
				return err
			}

			client := newRollbarClient(cfg)
			resp, err := client.ListUsers(cmd.Context())
			if err != nil {
				return err
			}

			switch output {
			case outputRawJSON:
				return writeJSON(resp.Raw)
			case outputJSON:
				return writeJSON(userListJSONOutput{Users: resp.Users})
			case outputNDJSON:
				records := make([]any, 0, len(resp.Users))
				for _, user := range resp.Users {
					records = append(records, user)
				}
				return writeNDJSON(records)
			default:
				return ui.RenderUsersWithOptions(resp.Users, ui.UserRenderOptions{
					Fields:    normalizeFields(listOpts.Fields),
					NoHeaders: listOpts.NoHeaders,
				})
			}
		},
	}

	getCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a user by ID",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			output, err := resolveOutputModeWithAliases(getOpts.Output, getOpts.JSON, getOpts.RawJSON, getOpts.NDJSON, outputText, outputJSON, outputRawJSON, outputNDJSON)
			if err != nil {
				return err
			}

			id, err := resolveUserID(cmd, args, getOpts)
			if err != nil {
				return err
			}

			client := newRollbarClient(cfg)
			resp, err := client.GetUserByID(cmd.Context(), id)
			if err != nil {
				return err
			}

			switch output {
			case outputRawJSON:
				return writeJSON(resp.Raw)
			case outputJSON:
				return writeJSON(userGetJSONOutput{User: resp.User})
			case outputNDJSON:
				return writeNDJSON([]any{resp.User})
			default:
				return ui.RenderUser(resp.User)
			}
		},
	}

	listCmd.Flags().StringVarP(&listOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json|ndjson")
	listCmd.Flags().BoolVar(&listOpts.JSON, "json", false, "Shortcut for --output json")
	listCmd.Flags().BoolVar(&listOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	listCmd.Flags().BoolVar(&listOpts.NDJSON, "ndjson", false, "Shortcut for --output ndjson")
	listCmd.Flags().StringSliceVar(&listOpts.Fields, "fields", nil, "Fields to render in text output")
	listCmd.Flags().BoolVar(&listOpts.NoHeaders, "no-headers", false, "Hide table headers in text output")

	getCmd.Flags().Int64Var(&getOpts.ID, "id", 0, "User ID")
	getCmd.Flags().StringVarP(&getOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json|ndjson")
	getCmd.Flags().BoolVar(&getOpts.JSON, "json", false, "Shortcut for --output json")
	getCmd.Flags().BoolVar(&getOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	getCmd.Flags().BoolVar(&getOpts.NDJSON, "ndjson", false, "Shortcut for --output ndjson")

	usersCmd.AddCommand(listCmd, getCmd)
	return usersCmd
}

func resolveUserID(cmd *cobra.Command, args []string, opts usersGetOptions) (int64, error) {
	idSet := cmd.Flags().Changed("id")
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	sources := 0
	if arg != "" {
		sources++
	}
	if idSet {
		sources++
	}
	if sources == 0 {
		return 0, fmt.Errorf("missing user identifier: pass [id] or --id")
	}
	if sources > 1 {
		return 0, fmt.Errorf("provide only one user identifier: [id] or --id")
	}

	if arg != "" {
		if !isIntegerToken(arg) {
			return 0, fmt.Errorf("invalid user id %q: must be > 0", arg)
		}
		return parsePositiveUserID(arg)
	}

	if opts.ID <= 0 {
		return 0, fmt.Errorf("invalid user id: must be > 0")
	}
	return opts.ID, nil
}

func parsePositiveUserID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid user id %q: must be > 0", raw)
	}
	return id, nil
}
