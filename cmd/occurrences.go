package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"rollbar-cli/internal/rollbar"
	"rollbar-cli/internal/ui"
)

type occurrencesListOptions struct {
	ItemID          int64
	ItemUUID        string
	Page            int
	Output          string
	JSON            bool
	RawJSON         bool
	NDJSON          bool
	Fields          []string
	NoHeaders       bool
	PayloadMode     string
	PayloadSections []string
	MaxPayloadBytes int
}

type occurrencesGetOptions struct {
	ID              int64
	UUID            string
	Output          string
	JSON            bool
	RawJSON         bool
	PayloadMode     string
	PayloadSections []string
	MaxPayloadBytes int
}

type occurrenceListJSONOutput struct {
	Occurrences []rollbar.ItemInstance `json:"occurrences"`
}

func newOccurrencesCmd(cfg *cliConfig) *cobra.Command {
	var listOpts occurrencesListOptions
	var getOpts occurrencesGetOptions

	occurrencesCmd := &cobra.Command{
		Use:     "occurrences",
		Aliases: []string{"occurences"},
		Short:   "Query Rollbar occurrences",
	}

	listCmd := &cobra.Command{
		Use:   "list [item-id-or-uuid]",
		Short: "List occurrences for a Rollbar item",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			output, err := resolveOutputModeWithAliases(listOpts.Output, listOpts.JSON, listOpts.RawJSON, listOpts.NDJSON, outputText, outputJSON, outputRawJSON, outputNDJSON)
			if err != nil {
				return err
			}

			itemID, itemUUID, err := resolveOccurrenceListItemIdentifier(cmd, args, listOpts)
			if err != nil {
				return err
			}

			client := newRollbarClient(cfg)
			identifier := itemUUID
			if identifier == "" {
				identifier = fmt.Sprintf("%d", itemID)
			}

			resp, err := client.ListItemInstances(cmd.Context(), identifier, listOpts.Page)
			if err != nil {
				return err
			}

			switch output {
			case outputRawJSON:
				return writeJSON(resp.Raw)
			case outputJSON:
				return writeJSON(occurrenceListJSONOutput{Occurrences: resp.Instances})
			case outputNDJSON:
				records := make([]any, 0, len(resp.Instances))
				for _, instance := range resp.Instances {
					records = append(records, instance)
				}
				return writeNDJSON(records)
			default:
				return ui.RenderOccurrencesWithOptions(resp.Instances, ui.OccurrenceRenderOptions{
					Fields:    normalizeFields(listOpts.Fields),
					NoHeaders: listOpts.NoHeaders,
					Payload: ui.PayloadRenderOptions{
						Mode:            listOpts.PayloadMode,
						Sections:        normalizeFields(listOpts.PayloadSections),
						MaxPayloadBytes: listOpts.MaxPayloadBytes,
					},
				})
			}
		},
	}

	getCmd := &cobra.Command{
		Use:   "get [id-or-uuid]",
		Short: "Get a single occurrence by ID or UUID",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			output, err := resolveOutputModeWithAliases(getOpts.Output, getOpts.JSON, getOpts.RawJSON, false, outputText, outputJSON, outputRawJSON)
			if err != nil {
				return err
			}

			id, uuid, err := resolveOccurrenceIdentifier(cmd, args, getOpts)
			if err != nil {
				return err
			}

			client := newRollbarClient(cfg)
			var resp *rollbar.GetOccurrenceResponse
			if uuid != "" {
				resp, err = client.GetOccurrenceByUUID(cmd.Context(), uuid)
			} else {
				resp, err = client.GetOccurrenceByID(cmd.Context(), id)
			}
			if err != nil {
				return err
			}

			switch output {
			case outputRawJSON:
				return writeJSON(resp.Raw)
			case outputJSON:
				return writeJSON(map[string]rollbar.ItemInstance{"occurrence": resp.Occurrence})
			default:
				return ui.RenderOccurrenceWithOptions(resp.Occurrence, ui.OccurrenceRenderOptions{
					Payload: ui.PayloadRenderOptions{
						Mode:            getOpts.PayloadMode,
						Sections:        normalizeFields(getOpts.PayloadSections),
						MaxPayloadBytes: getOpts.MaxPayloadBytes,
					},
				})
			}
		},
	}

	listCmd.Flags().Int64Var(&listOpts.ItemID, "item-id", 0, "Item ID")
	listCmd.Flags().StringVar(&listOpts.ItemUUID, "item-uuid", "", "Item UUID")
	listCmd.Flags().IntVar(&listOpts.Page, "page", 1, "Page number")
	listCmd.Flags().StringVarP(&listOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json|ndjson")
	listCmd.Flags().BoolVar(&listOpts.JSON, "json", false, "Shortcut for --output json")
	listCmd.Flags().BoolVar(&listOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	listCmd.Flags().BoolVar(&listOpts.NDJSON, "ndjson", false, "Shortcut for --output ndjson")
	listCmd.Flags().StringSliceVar(&listOpts.Fields, "fields", nil, "Fields to render in text output")
	listCmd.Flags().BoolVar(&listOpts.NoHeaders, "no-headers", false, "Hide table headers in text output")
	listCmd.Flags().StringVar(&listOpts.PayloadMode, "payload", "summary", "Payload mode for text output: none|summary|full")
	listCmd.Flags().StringSliceVar(&listOpts.PayloadSections, "payload-section", nil, "Payload sections to include")
	listCmd.Flags().IntVar(&listOpts.MaxPayloadBytes, "max-payload-bytes", 4096, "Maximum payload size to render in text output")

	getCmd.Flags().Int64Var(&getOpts.ID, "id", 0, "Occurrence ID")
	getCmd.Flags().StringVar(&getOpts.UUID, "uuid", "", "Occurrence UUID")
	getCmd.Flags().StringVarP(&getOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json")
	getCmd.Flags().BoolVar(&getOpts.JSON, "json", false, "Shortcut for --output json")
	getCmd.Flags().BoolVar(&getOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	getCmd.Flags().StringVar(&getOpts.PayloadMode, "payload", "summary", "Payload mode for text output: none|summary|full")
	getCmd.Flags().StringSliceVar(&getOpts.PayloadSections, "payload-section", nil, "Payload sections to include")
	getCmd.Flags().IntVar(&getOpts.MaxPayloadBytes, "max-payload-bytes", 4096, "Maximum payload size to render in text output")

	occurrencesCmd.AddCommand(listCmd, getCmd)
	return occurrencesCmd
}

func resolveOccurrenceIdentifier(cmd *cobra.Command, args []string, opts occurrencesGetOptions) (int64, string, error) {
	idSet := cmd.Flags().Changed("id")
	uuidSet := cmd.Flags().Changed("uuid")
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	return resolveIdentifierValue(arg, opts.ID, opts.UUID, idSet, uuidSet, "occurrence", "[id-or-uuid]", "--id", "--uuid")
}

func resolveOccurrenceListItemIdentifier(cmd *cobra.Command, args []string, opts occurrencesListOptions) (int64, string, error) {
	idSet := cmd.Flags().Changed("item-id")
	uuidSet := cmd.Flags().Changed("item-uuid")
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	return resolveIdentifierValue(arg, opts.ItemID, opts.ItemUUID, idSet, uuidSet, "item", "[item-id-or-uuid]", "--item-id", "--item-uuid")
}
