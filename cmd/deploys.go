package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/davebarnwell/rollbar-cli/internal/rollbar"
	"github.com/davebarnwell/rollbar-cli/internal/ui"
)

var validDeployStatuses = map[string]struct{}{
	"started":   {},
	"succeeded": {},
	"failed":    {},
	"timed_out": {},
}

type deploysListOptions struct {
	Page      int
	Limit     int
	Output    string
	JSON      bool
	RawJSON   bool
	NDJSON    bool
	Fields    []string
	NoHeaders bool
}

type deploysGetOptions struct {
	ID      int64
	Output  string
	JSON    bool
	RawJSON bool
	NDJSON  bool
}

type deploysCreateOptions struct {
	Environment     string
	Revision        string
	Status          string
	Comment         string
	LocalUsername   string
	RollbarUsername string
	Output          string
	JSON            bool
	RawJSON         bool
	NDJSON          bool
}

type deploysUpdateOptions struct {
	ID      int64
	Status  string
	Output  string
	JSON    bool
	RawJSON bool
	NDJSON  bool
}

type deployListJSONOutput struct {
	Deploys []rollbar.Deploy `json:"deploys"`
}

type deployGetJSONOutput struct {
	Deploy rollbar.Deploy `json:"deploy"`
}

func newDeploysCmd(cfg *cliConfig) *cobra.Command {
	var (
		listOpts   deploysListOptions
		getOpts    deploysGetOptions
		createOpts deploysCreateOptions
		updateOpts deploysUpdateOptions
	)

	deploysCmd := &cobra.Command{
		Use:     "deploys",
		Aliases: []string{"deploy", "deployment", "deployments"},
		Short:   "Query and manage Rollbar deploys",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List deploys in a Rollbar project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}
			return runDeploysList(cmd, cfg, listOpts)
		},
	}

	getCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a deploy by ID",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			output, err := resolveOutputModeWithAliases(getOpts.Output, getOpts.JSON, getOpts.RawJSON, getOpts.NDJSON, outputText, outputJSON, outputRawJSON, outputNDJSON)
			if err != nil {
				return err
			}

			id, err := resolveDeployID(cmd, args, getOpts.ID)
			if err != nil {
				return err
			}

			client := newRollbarClient(cfg)
			resp, err := client.GetDeployByID(cmd.Context(), id)
			if err != nil {
				return err
			}
			return writeSingleDeployOutput(resp.Deploy, resp.Raw, output)
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a deploy record",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			body, err := buildDeployCreateBody(createOpts)
			if err != nil {
				return err
			}

			output, err := resolveOutputModeWithAliases(createOpts.Output, createOpts.JSON, createOpts.RawJSON, createOpts.NDJSON, outputText, outputJSON, outputRawJSON, outputNDJSON)
			if err != nil {
				return err
			}

			client := newRollbarClient(cfg)
			resp, err := client.CreateDeploy(cmd.Context(), body)
			if err != nil {
				return err
			}
			return writeSingleDeployOutput(resp.Deploy, resp.Raw, output)
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update a deploy record",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireToken(cfg); err != nil {
				return err
			}

			id, err := resolveDeployID(cmd, args, updateOpts.ID)
			if err != nil {
				return err
			}

			body, err := buildDeployUpdateBody(updateOpts)
			if err != nil {
				return err
			}

			output, err := resolveOutputModeWithAliases(updateOpts.Output, updateOpts.JSON, updateOpts.RawJSON, updateOpts.NDJSON, outputText, outputJSON, outputRawJSON, outputNDJSON)
			if err != nil {
				return err
			}

			client := newRollbarClient(cfg)
			resp, err := client.UpdateDeployByID(cmd.Context(), id, body)
			if err != nil {
				return err
			}
			return writeSingleDeployOutput(resp.Deploy, resp.Raw, output)
		},
	}

	listCmd.Flags().IntVar(&listOpts.Page, "page", 1, "Page number")
	listCmd.Flags().IntVar(&listOpts.Limit, "limit", 0, "Maximum number of deploys to return")
	listCmd.Flags().StringVarP(&listOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json|ndjson")
	listCmd.Flags().BoolVar(&listOpts.JSON, "json", false, "Shortcut for --output json")
	listCmd.Flags().BoolVar(&listOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	listCmd.Flags().BoolVar(&listOpts.NDJSON, "ndjson", false, "Shortcut for --output ndjson")
	listCmd.Flags().StringSliceVar(&listOpts.Fields, "fields", nil, "Fields to render in text output")
	listCmd.Flags().BoolVar(&listOpts.NoHeaders, "no-headers", false, "Hide table headers in text output")

	getCmd.Flags().Int64Var(&getOpts.ID, "id", 0, "Deploy ID")
	getCmd.Flags().StringVarP(&getOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json|ndjson")
	getCmd.Flags().BoolVar(&getOpts.JSON, "json", false, "Shortcut for --output json")
	getCmd.Flags().BoolVar(&getOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	getCmd.Flags().BoolVar(&getOpts.NDJSON, "ndjson", false, "Shortcut for --output ndjson")

	createCmd.Flags().StringVar(&createOpts.Environment, "environment", "", "Deploy environment")
	createCmd.Flags().StringVar(&createOpts.Revision, "revision", "", "Deploy revision")
	createCmd.Flags().StringVar(&createOpts.Status, "status", "", "Deploy status: started|succeeded|failed|timed_out")
	createCmd.Flags().StringVar(&createOpts.Comment, "comment", "", "Deploy comment")
	createCmd.Flags().StringVar(&createOpts.LocalUsername, "local-username", "", "Local deploy username")
	createCmd.Flags().StringVar(&createOpts.RollbarUsername, "rollbar-username", "", "Rollbar username")
	createCmd.Flags().StringVarP(&createOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json|ndjson")
	createCmd.Flags().BoolVar(&createOpts.JSON, "json", false, "Shortcut for --output json")
	createCmd.Flags().BoolVar(&createOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	createCmd.Flags().BoolVar(&createOpts.NDJSON, "ndjson", false, "Shortcut for --output ndjson")

	updateCmd.Flags().Int64Var(&updateOpts.ID, "id", 0, "Deploy ID")
	updateCmd.Flags().StringVar(&updateOpts.Status, "status", "", "Deploy status: started|succeeded|failed|timed_out")
	updateCmd.Flags().StringVarP(&updateOpts.Output, "output", "o", outputText, "Output format: text|json|raw-json|ndjson")
	updateCmd.Flags().BoolVar(&updateOpts.JSON, "json", false, "Shortcut for --output json")
	updateCmd.Flags().BoolVar(&updateOpts.RawJSON, "raw-json", false, "Shortcut for --output raw-json")
	updateCmd.Flags().BoolVar(&updateOpts.NDJSON, "ndjson", false, "Shortcut for --output ndjson")

	deploysCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd)
	return deploysCmd
}

func runDeploysList(cmd *cobra.Command, cfg *cliConfig, opts deploysListOptions) error {
	output, err := resolveOutputModeWithAliases(opts.Output, opts.JSON, opts.RawJSON, opts.NDJSON, outputText, outputJSON, outputRawJSON, outputNDJSON)
	if err != nil {
		return err
	}

	deploys, raw, err := collectDeploys(cmd, cfg, opts)
	if err != nil {
		return err
	}

	switch output {
	case outputRawJSON:
		return writeJSON(raw)
	case outputJSON:
		return writeJSON(deployListJSONOutput{Deploys: deploys})
	case outputNDJSON:
		records := make([]any, 0, len(deploys))
		for _, deploy := range deploys {
			records = append(records, deploy)
		}
		return writeNDJSON(records)
	default:
		return ui.RenderDeploysWithOptions(deploys, ui.DeployRenderOptions{
			Fields:    normalizeFields(opts.Fields),
			NoHeaders: opts.NoHeaders,
		})
	}
}

func collectDeploys(cmd *cobra.Command, cfg *cliConfig, opts deploysListOptions) ([]rollbar.Deploy, map[string]any, error) {
	if opts.Limit < 0 {
		return nil, nil, fmt.Errorf("--limit must be >= 0")
	}

	client := newRollbarClient(cfg)
	page := opts.Page
	if page <= 0 {
		page = 1
	}

	resp, err := client.ListDeploys(cmd.Context(), rollbar.ListDeploysOptions{
		Page:  page,
		Limit: opts.Limit,
	})
	if err != nil {
		return nil, nil, err
	}

	deploys := resp.Deploys
	if opts.Limit > 0 && len(deploys) > opts.Limit {
		deploys = deploys[:opts.Limit]
	}
	return deploys, resp.Raw, nil
}

func writeSingleDeployOutput(deploy rollbar.Deploy, raw map[string]any, output string) error {
	switch output {
	case outputRawJSON:
		return writeJSON(raw)
	case outputJSON:
		return writeJSON(deployGetJSONOutput{Deploy: deploy})
	case outputNDJSON:
		return writeNDJSON([]any{deploy})
	default:
		return ui.RenderDeploy(deploy)
	}
}

func resolveDeployID(cmd *cobra.Command, args []string, id int64) (int64, error) {
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
		return 0, fmt.Errorf("missing deploy identifier: pass [id] or --id")
	}
	if sources > 1 {
		return 0, fmt.Errorf("provide only one deploy identifier: [id] or --id")
	}

	if arg != "" {
		if !isIntegerToken(arg) {
			return 0, fmt.Errorf("invalid deploy id %q: must be > 0", arg)
		}
		return parsePositiveDeployID(arg)
	}

	if id <= 0 {
		return 0, fmt.Errorf("invalid deploy id: must be > 0")
	}
	return id, nil
}

func parsePositiveDeployID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid deploy id %q: must be > 0", raw)
	}
	return id, nil
}

func buildDeployCreateBody(opts deploysCreateOptions) (map[string]any, error) {
	environment := strings.TrimSpace(opts.Environment)
	if environment == "" {
		return nil, fmt.Errorf("missing required flag: --environment")
	}
	revision := strings.TrimSpace(opts.Revision)
	if revision == "" {
		return nil, fmt.Errorf("missing required flag: --revision")
	}

	body := map[string]any{
		"environment": environment,
		"revision":    revision,
	}

	if status, err := normalizeDeployStatus(opts.Status); err != nil {
		return nil, err
	} else if status != "" {
		body["status"] = status
	}
	if comment := strings.TrimSpace(opts.Comment); comment != "" {
		body["comment"] = comment
	}
	if localUsername := strings.TrimSpace(opts.LocalUsername); localUsername != "" {
		body["local_username"] = localUsername
	}
	if rollbarUsername := strings.TrimSpace(opts.RollbarUsername); rollbarUsername != "" {
		body["rollbar_username"] = rollbarUsername
	}
	return body, nil
}

func buildDeployUpdateBody(opts deploysUpdateOptions) (map[string]any, error) {
	status, err := normalizeDeployStatus(opts.Status)
	if err != nil {
		return nil, err
	}
	if status == "" {
		return nil, fmt.Errorf("missing required flag: --status")
	}

	return map[string]any{"status": status}, nil
}

func normalizeDeployStatus(raw string) (string, error) {
	status := strings.TrimSpace(strings.ToLower(raw))
	if status == "" {
		return "", nil
	}
	if _, ok := validDeployStatuses[status]; !ok {
		return "", fmt.Errorf("invalid deploy status %q: expected started|succeeded|failed|timed_out", raw)
	}
	return status, nil
}
