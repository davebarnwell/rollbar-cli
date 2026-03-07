package ui

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/davebarnwell/rollbar-cli/internal/rollbar"
)

type DeployRenderOptions struct {
	Fields    []string
	NoHeaders bool
}

var defaultDeployListFields = []string{"id", "status", "environment", "revision", "start_time", "finish_time", "comment"}

func RenderDeploy(deploy rollbar.Deploy) error {
	return renderDeploy(os.Stdout, deploy)
}

func RenderDeploys(deploys []rollbar.Deploy) error {
	return RenderDeploysWithOptions(deploys, DeployRenderOptions{})
}

func RenderDeploysWithOptions(deploys []rollbar.Deploy, opts DeployRenderOptions) error {
	if len(deploys) == 0 {
		_, err := fmt.Fprintln(os.Stdout, "No deploys found.")
		return err
	}
	return renderDeploysPlain(os.Stdout, deploys, opts)
}

func DefaultDeployListFields() []string {
	return append([]string(nil), defaultDeployListFields...)
}

func renderDeploy(w io.Writer, deploy rollbar.Deploy) error {
	if _, err := fmt.Fprintf(w, "ID: %d\n", deploy.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Project ID: %d\n", deploy.ProjectID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Environment: %s\n", fallback(deploy.Environment)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Revision: %s\n", fallback(deploy.Revision)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Status: %s\n", fallback(deploy.Status)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Comment: %s\n", fallback(deploy.Comment)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Local Username: %s\n", fallback(deploy.LocalUsername)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Rollbar Username: %s\n", fallback(deploy.RollbarUsername)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Start Time: %s\n", formatUnix(deploy.StartTime)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Finish Time: %s\n", formatUnix(deploy.FinishTime)); err != nil {
		return err
	}
	return nil
}

func renderDeploysPlain(w io.Writer, deploys []rollbar.Deploy, opts DeployRenderOptions) error {
	fields := opts.Fields
	if len(fields) == 0 {
		fields = defaultDeployListFields
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if !opts.NoHeaders {
		if _, err := fmt.Fprintln(tw, strings.Join(fieldHeaders(fields), "\t")); err != nil {
			return err
		}
	}
	for _, deploy := range deploys {
		if _, err := fmt.Fprintln(tw, strings.Join(deployFieldValues(deploy, fields), "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func deployFieldValues(deploy rollbar.Deploy, fields []string) []string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		switch field {
		case "id", "deploy_id":
			values = append(values, strconv.FormatInt(deploy.ID, 10))
		case "project_id":
			values = append(values, strconv.FormatInt(deploy.ProjectID, 10))
		case "environment":
			values = append(values, fallback(deploy.Environment))
		case "revision":
			values = append(values, fallback(deploy.Revision))
		case "status":
			values = append(values, fallback(deploy.Status))
		case "comment":
			values = append(values, fallback(deploy.Comment))
		case "local_username":
			values = append(values, fallback(deploy.LocalUsername))
		case "rollbar_username", "rollbar_name":
			values = append(values, fallback(deploy.RollbarUsername))
		case "start_time":
			values = append(values, formatUnix(deploy.StartTime))
		case "finish_time":
			values = append(values, formatUnix(deploy.FinishTime))
		default:
			values = append(values, "-")
		}
	}
	return values
}
