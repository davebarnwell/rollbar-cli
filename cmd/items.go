package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/davebarnwell/rollbar-cli/internal/rollbar"
	"github.com/davebarnwell/rollbar-cli/internal/ui"
)

type itemsListOptions struct {
	Page        int
	Pages       int
	Status      string
	Environment string
	Level       []string
	Output      string
	JSON        bool
	RawJSON     bool
	NDJSON      bool
	Fields      []string
	NoHeaders   bool
	Sort        string
	Limit       int
	Since       string
	Until       string
	Last        time.Duration
}

type itemsGetOptions struct {
	ID              int64
	UUID            string
	Output          string
	JSON            bool
	RawJSON         bool
	WithInstances   bool
	InstancesPage   int
	PayloadMode     string
	PayloadSections []string
	MaxPayloadBytes int
}

type itemsUpdateOptions struct {
	ID                      int64
	UUID                    string
	Status                  string
	ResolvedInVersion       string
	Title                   string
	Level                   string
	AssignedUserID          int64
	ClearAssignedUser       bool
	AssignedTeamID          int64
	ClearAssignedTeam       bool
	SnoozeEnabled           bool
	SnoozeExpirationSeconds int
	Output                  string
	JSON                    bool
	RawJSON                 bool
}

type itemsResolveOptions struct {
	ResolvedInVersion string
	Output            string
	JSON              bool
	RawJSON           bool
}

type itemsMuteOptions struct {
	Output  string
	JSON    bool
	RawJSON bool
}

type itemsAssignOptions struct {
	AssignedUserID    int64
	ClearAssignedUser bool
	AssignedTeamID    int64
	ClearAssignedTeam bool
	Output            string
	JSON              bool
	RawJSON           bool
}

type itemsSnoozeOptions struct {
	Duration time.Duration
	Disable  bool
	Output   string
	JSON     bool
	RawJSON  bool
}

type itemsWatchOptions struct {
	Interval time.Duration
	Count    int
}

type itemGetJSONOutput struct {
	Item      rollbar.Item           `json:"item"`
	Instances []rollbar.ItemInstance `json:"instances,omitempty"`
}

type itemListJSONOutput struct {
	Items []rollbar.Item `json:"items"`
}

func newItemsCmd(cfg *cliConfig) *cobra.Command {
	var (
		listOpts    itemsListOptions
		getOpts     itemsGetOptions
		updateOpts  itemsUpdateOptions
		resolveOpts itemsResolveOptions
		muteOpts    itemsMuteOptions
		assignOpts  itemsAssignOptions
		snoozeOpts  itemsSnoozeOptions
		watchOpts   itemsWatchOptions
	)

	itemsCmd := &cobra.Command{
		Use:   "items",
		Short: "Query Rollbar items",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List items in a Rollbar project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}
			return runItemsList(cmd, cfg, listOpts)
		},
	}

	getCmd := &cobra.Command{
		Use:   "get [id-or-uuid]",
		Short: "Get a single Rollbar item by ID or UUID",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			output, err := resolveOutputModeWithAliases(getOpts.Output, getOpts.JSON, getOpts.RawJSON, false, outputText, outputJSON, outputRawJSON)
			if err != nil {
				return err
			}

			resp, instancesResp, err := getItemAndInstances(cmd, cfg, args, getOpts)
			if err != nil {
				return err
			}

			switch output {
			case outputRawJSON:
				if instancesResp == nil {
					return writeJSON(resp.Raw)
				}
				return writeJSON(map[string]any{
					"item":      resp.Raw,
					"instances": instancesResp.Raw,
				})
			case outputJSON:
				out := itemGetJSONOutput{Item: resp.Item}
				if instancesResp != nil {
					out.Instances = instancesResp.Instances
				}
				return writeJSON(out)
			default:
				payloadOpts := ui.PayloadRenderOptions{
					Mode:            getOpts.PayloadMode,
					Sections:        normalizeFields(getOpts.PayloadSections),
					MaxPayloadBytes: getOpts.MaxPayloadBytes,
				}
				if instancesResp != nil {
					return ui.RenderItemWithInstancesOptions(resp.Item, instancesResp.Instances, ui.ItemDetailsRenderOptions{
						Payload: payloadOpts,
					})
				}
				return ui.RenderItemWithOptions(resp.Item, ui.ItemDetailsRenderOptions{Payload: payloadOpts})
			}
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update [id-or-uuid]",
		Short: "Update a Rollbar item",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			body, err := buildItemUpdateBody(cmd, updateOpts)
			if err != nil {
				return err
			}

			output, err := resolveOutputModeWithAliases(updateOpts.Output, updateOpts.JSON, updateOpts.RawJSON, false, outputText, outputJSON, outputRawJSON)
			if err != nil {
				return err
			}

			return executeItemUpdate(cmd, cfg, args, body, output, updateOpts.ID, updateOpts.UUID)
		},
	}

	resolveCmd := &cobra.Command{
		Use:   "resolve [id-or-uuid]",
		Short: "Resolve a Rollbar item",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}
			body := map[string]any{"status": "resolved"}
			if version := strings.TrimSpace(resolveOpts.ResolvedInVersion); version != "" {
				body["resolved_in_version"] = version
			}
			output, err := resolveOutputModeWithAliases(resolveOpts.Output, resolveOpts.JSON, resolveOpts.RawJSON, false, outputText, outputJSON, outputRawJSON)
			if err != nil {
				return err
			}
			return executeItemUpdate(cmd, cfg, args, body, output, 0, "")
		},
	}

	muteCmd := &cobra.Command{
		Use:   "mute [id-or-uuid]",
		Short: "Mute a Rollbar item",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}
			output, err := resolveOutputModeWithAliases(muteOpts.Output, muteOpts.JSON, muteOpts.RawJSON, false, outputText, outputJSON, outputRawJSON)
			if err != nil {
				return err
			}
			return executeItemUpdate(cmd, cfg, args, map[string]any{"status": "muted"}, output, 0, "")
		},
	}

	assignCmd := &cobra.Command{
		Use:   "assign [id-or-uuid]",
		Short: "Assign a Rollbar item to a user and/or team",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			body := make(map[string]any)
			if assignOpts.ClearAssignedUser && assignOpts.AssignedUserID > 0 {
				return fmt.Errorf("use either --assigned-user-id or --clear-assigned-user, not both")
			}
			if assignOpts.ClearAssignedTeam && assignOpts.AssignedTeamID > 0 {
				return fmt.Errorf("use either --assigned-team-id or --clear-assigned-team, not both")
			}
			if assignOpts.ClearAssignedUser {
				body["assigned_user_id"] = nil
			}
			if assignOpts.AssignedUserID > 0 {
				body["assigned_user_id"] = assignOpts.AssignedUserID
			}
			if assignOpts.ClearAssignedTeam {
				body["assigned_team_id"] = nil
			}
			if assignOpts.AssignedTeamID > 0 {
				body["assigned_team_id"] = assignOpts.AssignedTeamID
			}
			if len(body) == 0 {
				return fmt.Errorf("no assignment changes provided")
			}

			output, err := resolveOutputModeWithAliases(assignOpts.Output, assignOpts.JSON, assignOpts.RawJSON, false, outputText, outputJSON, outputRawJSON)
			if err != nil {
				return err
			}
			return executeItemUpdate(cmd, cfg, args, body, output, 0, "")
		},
	}

	snoozeCmd := &cobra.Command{
		Use:   "snooze [id-or-uuid]",
		Short: "Snooze or unsnooze a Rollbar item",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			body := make(map[string]any)
			switch {
			case snoozeOpts.Disable:
				body["snooze_enabled"] = false
			case snoozeOpts.Duration > 0:
				body["snooze_enabled"] = true
				body["snooze_expiration_in_seconds"] = int(snoozeOpts.Duration.Seconds())
			default:
				return fmt.Errorf("set either --duration or --disable")
			}

			output, err := resolveOutputModeWithAliases(snoozeOpts.Output, snoozeOpts.JSON, snoozeOpts.RawJSON, false, outputText, outputJSON, outputRawJSON)
			if err != nil {
				return err
			}
			return executeItemUpdate(cmd, cfg, args, body, output, 0, "")
		},
	}

	watchCmd := &cobra.Command{
		Use:   "watch",
		Short: "Poll the item list on an interval",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}
			if watchOpts.Interval <= 0 {
				return fmt.Errorf("--interval must be > 0")
			}
			if watchOpts.Count <= 0 {
				return fmt.Errorf("--count must be > 0")
			}

			for i := 0; i < watchOpts.Count; i++ {
				if i > 0 {
					if err := writeStdoutf("\n[%s]\n", time.Now().UTC().Format(time.RFC3339)); err != nil {
						return err
					}
				}
				if err := runItemsList(cmd, cfg, prepareWatchListOptions(listOpts)); err != nil {
					return err
				}
				if i+1 < watchOpts.Count {
					select {
					case <-cmd.Context().Done():
						return cmd.Context().Err()
					case <-time.After(watchOpts.Interval):
					}
				}
			}
			return nil
		},
	}

	listCmd.Flags().IntVar(&listOpts.Page, "page", 1, "Starting page number")
	listCmd.Flags().IntVar(&listOpts.Pages, "pages", 1, "Number of pages to fetch")
	listCmd.Flags().StringVar(&listOpts.Status, "status", "", "Filter by item status")
	listCmd.Flags().StringVar(&listOpts.Environment, "environment", "", "Filter by environment")
	listCmd.Flags().StringSliceVar(&listOpts.Level, "level", nil, "Filter by level; pass multiple times for multiple levels")
	listCmd.Flags().StringVarP(&listOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json|ndjson")
	listCmd.Flags().BoolVar(&listOpts.JSON, "json", false, "Shortcut for --output json")
	listCmd.Flags().BoolVar(&listOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	listCmd.Flags().BoolVar(&listOpts.NDJSON, "ndjson", false, "Shortcut for --output ndjson")
	listCmd.Flags().StringSliceVar(&listOpts.Fields, "fields", nil, "Fields to render in text output")
	listCmd.Flags().BoolVar(&listOpts.NoHeaders, "no-headers", false, "Hide table headers in text output")
	listCmd.Flags().StringVar(&listOpts.Sort, "sort", "last_seen_desc", "Sort order: last_seen_desc|last_seen_asc|counter_desc|counter_asc|title|level")
	listCmd.Flags().IntVar(&listOpts.Limit, "limit", 0, "Maximum number of items to return after filtering")
	listCmd.Flags().StringVar(&listOpts.Since, "since", "", "Only include items seen at or after this time")
	listCmd.Flags().StringVar(&listOpts.Until, "until", "", "Only include items seen at or before this time")
	listCmd.Flags().DurationVar(&listOpts.Last, "last", 0, "Only include items seen within this duration")

	getCmd.Flags().Int64Var(&getOpts.ID, "id", 0, "Item ID")
	getCmd.Flags().StringVar(&getOpts.UUID, "uuid", "", "Item UUID")
	getCmd.Flags().StringVarP(&getOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json")
	getCmd.Flags().BoolVar(&getOpts.JSON, "json", false, "Shortcut for --output json")
	getCmd.Flags().BoolVar(&getOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	getCmd.Flags().BoolVar(&getOpts.WithInstances, "instances", false, "Include item instances")
	getCmd.Flags().IntVar(&getOpts.InstancesPage, "instances-page", 1, "Instances page to fetch when --instances is set")
	getCmd.Flags().StringVar(&getOpts.PayloadMode, "payload", "summary", "Payload mode for text output: none|summary|full")
	getCmd.Flags().StringSliceVar(&getOpts.PayloadSections, "payload-section", nil, "Payload sections to include")
	getCmd.Flags().IntVar(&getOpts.MaxPayloadBytes, "max-payload-bytes", 4096, "Maximum payload size to render in text output")

	updateCmd.Flags().Int64Var(&updateOpts.ID, "id", 0, "Item ID")
	updateCmd.Flags().StringVar(&updateOpts.UUID, "uuid", "", "Item UUID")
	updateCmd.Flags().StringVar(&updateOpts.Status, "status", "", "New status: active|resolved|muted")
	updateCmd.Flags().StringVar(&updateOpts.ResolvedInVersion, "resolved-in-version", "", "Resolved version (max 40 chars)")
	updateCmd.Flags().StringVar(&updateOpts.Title, "title", "", "New title (1-255 chars)")
	updateCmd.Flags().StringVar(&updateOpts.Level, "level", "", "New level: critical|error|warning|info|debug")
	updateCmd.Flags().Int64Var(&updateOpts.AssignedUserID, "assigned-user-id", 0, "Assign to user ID")
	updateCmd.Flags().BoolVar(&updateOpts.ClearAssignedUser, "clear-assigned-user", false, "Clear assigned user")
	updateCmd.Flags().Int64Var(&updateOpts.AssignedTeamID, "assigned-team-id", 0, "Assign to team ID")
	updateCmd.Flags().BoolVar(&updateOpts.ClearAssignedTeam, "clear-assigned-team", false, "Clear assigned team")
	updateCmd.Flags().BoolVar(&updateOpts.SnoozeEnabled, "snooze-enabled", false, "Set snooze enabled state")
	updateCmd.Flags().IntVar(&updateOpts.SnoozeExpirationSeconds, "snooze-expiration-seconds", 0, "Snooze expiration in seconds")
	updateCmd.Flags().StringVarP(&updateOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json")
	updateCmd.Flags().BoolVar(&updateOpts.JSON, "json", false, "Shortcut for --output json")
	updateCmd.Flags().BoolVar(&updateOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")

	resolveCmd.Flags().StringVar(&resolveOpts.ResolvedInVersion, "resolved-in-version", "", "Resolved version")
	resolveCmd.Flags().StringVarP(&resolveOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json")
	resolveCmd.Flags().BoolVar(&resolveOpts.JSON, "json", false, "Shortcut for --output json")
	resolveCmd.Flags().BoolVar(&resolveOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")

	muteCmd.Flags().StringVarP(&muteOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json")
	muteCmd.Flags().BoolVar(&muteOpts.JSON, "json", false, "Shortcut for --output json")
	muteCmd.Flags().BoolVar(&muteOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")

	assignCmd.Flags().Int64Var(&assignOpts.AssignedUserID, "assigned-user-id", 0, "Assign to user ID")
	assignCmd.Flags().BoolVar(&assignOpts.ClearAssignedUser, "clear-assigned-user", false, "Clear assigned user")
	assignCmd.Flags().Int64Var(&assignOpts.AssignedTeamID, "assigned-team-id", 0, "Assign to team ID")
	assignCmd.Flags().BoolVar(&assignOpts.ClearAssignedTeam, "clear-assigned-team", false, "Clear assigned team")
	assignCmd.Flags().StringVarP(&assignOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json")
	assignCmd.Flags().BoolVar(&assignOpts.JSON, "json", false, "Shortcut for --output json")
	assignCmd.Flags().BoolVar(&assignOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")

	snoozeCmd.Flags().DurationVar(&snoozeOpts.Duration, "duration", 0, "How long to snooze the item")
	snoozeCmd.Flags().BoolVar(&snoozeOpts.Disable, "disable", false, "Disable snooze")
	snoozeCmd.Flags().StringVarP(&snoozeOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json")
	snoozeCmd.Flags().BoolVar(&snoozeOpts.JSON, "json", false, "Shortcut for --output json")
	snoozeCmd.Flags().BoolVar(&snoozeOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")

	watchCmd.Flags().AddFlagSet(listCmd.Flags())
	watchCmd.Flags().DurationVar(&watchOpts.Interval, "interval", 30*time.Second, "Polling interval")
	watchCmd.Flags().IntVar(&watchOpts.Count, "count", 1, "Number of polls to run")

	itemsCmd.AddCommand(listCmd, getCmd, updateCmd, resolveCmd, muteCmd, assignCmd, snoozeCmd, watchCmd)
	return itemsCmd
}

func runItemsList(cmd *cobra.Command, cfg *cliConfig, opts itemsListOptions) error {
	output, err := resolveOutputModeWithAliases(opts.Output, opts.JSON, opts.RawJSON, opts.NDJSON, outputText, outputJSON, outputRawJSON, outputNDJSON)
	if err != nil {
		return err
	}

	items, raw, err := collectAndShapeItems(cmd, cfg, opts)
	if err != nil {
		return err
	}

	switch output {
	case outputRawJSON:
		return writeJSON(raw)
	case outputJSON:
		return writeJSON(itemListJSONOutput{Items: items})
	case outputNDJSON:
		records := make([]any, 0, len(items))
		for _, item := range items {
			records = append(records, item)
		}
		return writeNDJSON(records)
	default:
		client := newRollbarClient(cfg)
		return ui.RenderItemsWithOptions(items, ui.ItemListRenderOptions{
			Fields:    normalizeFields(opts.Fields),
			NoHeaders: opts.NoHeaders,
			Interactions: &ui.ItemListInteractions{
				FetchOccurrences: func(item rollbar.Item) ([]rollbar.ItemInstance, error) {
					resp, err := client.ListItemInstances(cmd.Context(), strconv.FormatInt(item.ID, 10), 1)
					if err != nil {
						return nil, err
					}
					sortOccurrences(resp.Instances)
					return resp.Instances, nil
				},
				ResolveItem: func(item rollbar.Item) (rollbar.Item, error) {
					return updateItemForTUI(cmd, client, item.ID, map[string]any{"status": "resolved"})
				},
				MuteItem: func(item rollbar.Item) (rollbar.Item, error) {
					return updateItemForTUI(cmd, client, item.ID, map[string]any{"status": "muted"})
				},
			},
		})
	}
}

func prepareWatchListOptions(opts itemsListOptions) itemsListOptions {
	output, err := resolveOutputModeWithAliases(opts.Output, opts.JSON, opts.RawJSON, opts.NDJSON, outputText, outputJSON, outputRawJSON, outputNDJSON)
	if err != nil {
		return opts
	}
	normalizedFields := normalizeFields(opts.Fields)
	if output == outputText && len(normalizedFields) == 0 {
		opts.Fields = ui.DefaultItemListFields()
		return opts
	}
	opts.Fields = normalizedFields
	return opts
}

func updateItemForTUI(cmd *cobra.Command, client *rollbar.Client, id int64, body map[string]any) (rollbar.Item, error) {
	resp, err := client.UpdateItemByID(cmd.Context(), id, body)
	if err != nil {
		return rollbar.Item{}, err
	}
	if resp.Item.ID > 0 {
		return resp.Item, nil
	}
	getResp, err := client.GetItemByID(cmd.Context(), id)
	if err != nil {
		return rollbar.Item{}, err
	}
	return getResp.Item, nil
}

func collectAndShapeItems(cmd *cobra.Command, cfg *cliConfig, opts itemsListOptions) ([]rollbar.Item, map[string]any, error) {
	if opts.Pages <= 0 {
		opts.Pages = 1
	}
	if opts.Limit < 0 {
		return nil, nil, fmt.Errorf("--limit must be >= 0")
	}

	client := newRollbarClient(cfg)
	startPage := opts.Page
	if startPage <= 0 {
		startPage = 1
	}

	items := make([]rollbar.Item, 0)
	rawPages := make([]map[string]any, 0, opts.Pages)
	for pageOffset := 0; pageOffset < opts.Pages; pageOffset++ {
		resp, err := client.ListItems(cmd.Context(), rollbar.ListItemsOptions{
			Page:        startPage + pageOffset,
			Status:      opts.Status,
			Environment: opts.Environment,
			Level:       opts.Level,
		})
		if err != nil {
			return nil, nil, err
		}
		items = append(items, resp.Items...)
		rawPages = append(rawPages, resp.Raw)
		if len(resp.Items) == 0 {
			break
		}
	}

	since, until, err := parseItemTimeRange(opts)
	if err != nil {
		return nil, nil, err
	}
	items = filterItemsByTime(items, since, until)
	sortItemsWithDirection(items, opts.Sort)
	if opts.Limit > 0 && len(items) > opts.Limit {
		items = items[:opts.Limit]
	}

	return items, map[string]any{"pages": rawPages}, nil
}

func getItemAndInstances(cmd *cobra.Command, cfg *cliConfig, args []string, opts itemsGetOptions) (*rollbar.GetItemResponse, *rollbar.ListItemInstancesResponse, error) {
	idSet := cmd.Flags().Changed("id")
	uuidSet := cmd.Flags().Changed("uuid")
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	id, uuid, err := resolveIdentifierValue(arg, opts.ID, opts.UUID, idSet, uuidSet, "item", "[id-or-uuid]", "--id", "--uuid")
	if err != nil {
		return nil, nil, err
	}

	client := newRollbarClient(cfg)
	var resp *rollbar.GetItemResponse
	if uuid != "" {
		resp, err = client.GetItemByUUID(cmd.Context(), uuid)
	} else {
		resp, err = client.GetItemByID(cmd.Context(), id)
	}
	if err != nil {
		return nil, nil, err
	}

	if !opts.WithInstances {
		return resp, nil, nil
	}

	instanceIdentifier := uuid
	if resp.Item.ID > 0 {
		instanceIdentifier = strconv.FormatInt(resp.Item.ID, 10)
	}
	if instanceIdentifier == "" {
		instanceIdentifier = strconv.FormatInt(id, 10)
	}
	instancesResp, err := client.ListItemInstances(cmd.Context(), instanceIdentifier, opts.InstancesPage)
	if err != nil {
		return nil, nil, err
	}
	return resp, instancesResp, nil
}

func executeItemUpdate(cmd *cobra.Command, cfg *cliConfig, args []string, body map[string]any, output string, explicitID int64, explicitUUID string) error {
	idSet := explicitID > 0 || cmd.Flags().Changed("id")
	uuidSet := explicitUUID != "" || cmd.Flags().Changed("uuid")
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	id, uuid, err := resolveIdentifierValue(arg, explicitID, explicitUUID, idSet, uuidSet, "item", "[id-or-uuid]", "--id", "--uuid")
	if err != nil {
		return err
	}

	client := newRollbarClient(cfg)
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

	updateResp, err := client.UpdateItemByID(cmd.Context(), id, body)
	if err != nil {
		return err
	}

	switch output {
	case outputRawJSON:
		return writeJSON(updateResp.Raw)
	case outputJSON:
		return writeJSON(itemGetJSONOutput{Item: updateResp.Item})
	default:
		if updateResp.Item.ID > 0 {
			return ui.RenderItem(updateResp.Item)
		}
		getResp, err := client.GetItemByID(cmd.Context(), id)
		if err == nil && getResp.Item.ID > 0 {
			return ui.RenderItem(getResp.Item)
		}
		return writeStdoutf("Item %d updated.\n", id)
	}
}

func buildItemUpdateBody(cmd *cobra.Command, opts itemsUpdateOptions) (map[string]any, error) {
	body := make(map[string]any)

	if cmd.Flags().Changed("status") {
		status := strings.ToLower(strings.TrimSpace(opts.Status))
		switch status {
		case "active", "resolved", "muted":
			body["status"] = status
		default:
			return nil, fmt.Errorf("invalid --status %q (expected: active|resolved|muted)", opts.Status)
		}
	}

	if cmd.Flags().Changed("resolved-in-version") {
		if len(opts.ResolvedInVersion) > 40 {
			return nil, fmt.Errorf("--resolved-in-version cannot exceed 40 characters")
		}
		body["resolved_in_version"] = opts.ResolvedInVersion
	}

	if cmd.Flags().Changed("title") {
		titleLen := len(opts.Title)
		if titleLen < 1 || titleLen > 255 {
			return nil, fmt.Errorf("--title must be between 1 and 255 characters")
		}
		body["title"] = opts.Title
	}

	if cmd.Flags().Changed("level") {
		level := strings.ToLower(strings.TrimSpace(opts.Level))
		switch level {
		case "critical", "error", "warning", "info", "debug":
			body["level"] = level
		default:
			return nil, fmt.Errorf("invalid --level %q (expected: critical|error|warning|info|debug)", opts.Level)
		}
	}

	if opts.ClearAssignedUser && cmd.Flags().Changed("assigned-user-id") {
		return nil, fmt.Errorf("use either --assigned-user-id or --clear-assigned-user, not both")
	}
	if opts.ClearAssignedUser {
		body["assigned_user_id"] = nil
	}
	if cmd.Flags().Changed("assigned-user-id") {
		if opts.AssignedUserID <= 0 {
			return nil, fmt.Errorf("--assigned-user-id must be > 0")
		}
		body["assigned_user_id"] = opts.AssignedUserID
	}

	if opts.ClearAssignedTeam && cmd.Flags().Changed("assigned-team-id") {
		return nil, fmt.Errorf("use either --assigned-team-id or --clear-assigned-team, not both")
	}
	if opts.ClearAssignedTeam {
		body["assigned_team_id"] = nil
	}
	if cmd.Flags().Changed("assigned-team-id") {
		if opts.AssignedTeamID <= 0 {
			return nil, fmt.Errorf("--assigned-team-id must be > 0")
		}
		body["assigned_team_id"] = opts.AssignedTeamID
	}

	if cmd.Flags().Changed("snooze-enabled") {
		body["snooze_enabled"] = opts.SnoozeEnabled
	}
	if cmd.Flags().Changed("snooze-expiration-seconds") {
		if opts.SnoozeExpirationSeconds <= 0 {
			return nil, fmt.Errorf("--snooze-expiration-seconds must be > 0")
		}
		body["snooze_expiration_in_seconds"] = opts.SnoozeExpirationSeconds
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("no updates provided: set at least one update flag")
	}
	return body, nil
}

func parseItemTimeRange(opts itemsListOptions) (time.Time, time.Time, error) {
	if opts.Last > 0 && strings.TrimSpace(opts.Since) != "" {
		return time.Time{}, time.Time{}, fmt.Errorf("use either --last or --since, not both")
	}

	var since time.Time
	var until time.Time
	var err error

	if strings.TrimSpace(opts.Since) != "" {
		since, err = parseTimeFilter(opts.Since)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parse --since: %w", err)
		}
	}
	if strings.TrimSpace(opts.Until) != "" {
		until, err = parseTimeFilter(opts.Until)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parse --until: %w", err)
		}
	}
	if opts.Last > 0 {
		since = time.Now().UTC().Add(-opts.Last)
	}
	return since, until, nil
}

func filterItemsByTime(items []rollbar.Item, since time.Time, until time.Time) []rollbar.Item {
	return filterItems(items, since, until, 0)
}

func sortItemsWithDirection(items []rollbar.Item, sortBy string) {
	switch strings.TrimSpace(strings.ToLower(sortBy)) {
	case "", "last_seen_desc":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].LastOccurrenceTimestamp > items[j].LastOccurrenceTimestamp
		})
	case "last_seen_asc":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].LastOccurrenceTimestamp < items[j].LastOccurrenceTimestamp
		})
	case "counter_desc":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].Counter > items[j].Counter
		})
	case "counter_asc":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].Counter < items[j].Counter
		})
	case "title":
		sort.SliceStable(items, func(i, j int) bool {
			return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title)
		})
	case "level":
		sort.SliceStable(items, func(i, j int) bool {
			return strings.ToLower(items[i].Level) < strings.ToLower(items[j].Level)
		})
	default:
		sortItems(items, "")
	}
}

func normalizeFields(fields []string) []string {
	normalized := make([]string, 0, len(fields))
	seen := make(map[string]struct{})
	for _, field := range fields {
		value := strings.TrimSpace(strings.ToLower(field))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}
