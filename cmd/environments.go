package cmd

import (
	"github.com/spf13/cobra"

	"rollbar-cli/internal/rollbar"
	"rollbar-cli/internal/ui"
)

type environmentsListOptions struct {
	Output    string
	JSON      bool
	RawJSON   bool
	NDJSON    bool
	Fields    []string
	NoHeaders bool
}

type environmentListJSONOutput struct {
	Environments []rollbar.Environment `json:"environments"`
}

type environmentListRawOutput struct {
	Pages []map[string]any `json:"pages"`
}

func newEnvironmentsCmd(cfg *cliConfig) *cobra.Command {
	var listOpts environmentsListOptions

	environmentsCmd := &cobra.Command{
		Use:   "environments",
		Short: "Query Rollbar environments",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all environments in the Rollbar account",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			output, err := resolveOutputModeWithAliases(listOpts.Output, listOpts.JSON, listOpts.RawJSON, listOpts.NDJSON, outputText, outputJSON, outputRawJSON, outputNDJSON)
			if err != nil {
				return err
			}

			client := newRollbarClient(cfg)
			resp, err := client.ListEnvironments(cmd.Context())
			if err != nil {
				return err
			}

			switch output {
			case outputRawJSON:
				return writeJSON(environmentListRawOutput{Pages: resp.RawPages})
			case outputJSON:
				return writeJSON(environmentListJSONOutput{Environments: resp.Environments})
			case outputNDJSON:
				records := make([]any, 0, len(resp.Environments))
				for _, environment := range resp.Environments {
					records = append(records, environment)
				}
				return writeNDJSON(records)
			default:
				return ui.RenderEnvironmentsWithOptions(resp.Environments, ui.EnvironmentRenderOptions{
					Fields:    normalizeFields(listOpts.Fields),
					NoHeaders: listOpts.NoHeaders,
				})
			}
		},
	}

	listCmd.Flags().StringVarP(&listOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json|ndjson")
	listCmd.Flags().BoolVar(&listOpts.JSON, "json", false, "Shortcut for --output json")
	listCmd.Flags().BoolVar(&listOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	listCmd.Flags().BoolVar(&listOpts.NDJSON, "ndjson", false, "Shortcut for --output ndjson")
	listCmd.Flags().StringSliceVar(&listOpts.Fields, "fields", nil, "Fields to render in text output")
	listCmd.Flags().BoolVar(&listOpts.NoHeaders, "no-headers", false, "Hide table headers in text output")

	environmentsCmd.AddCommand(listCmd)
	return environmentsCmd
}
