package ui

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"rollbar-cli/internal/rollbar"
)

type EnvironmentRenderOptions struct {
	Fields    []string
	NoHeaders bool
}

func RenderEnvironments(environments []rollbar.Environment) error {
	return RenderEnvironmentsWithOptions(environments, EnvironmentRenderOptions{})
}

func RenderEnvironmentsWithOptions(environments []rollbar.Environment, opts EnvironmentRenderOptions) error {
	if len(environments) == 0 {
		_, err := fmt.Fprintln(os.Stdout, "No environments found.")
		return err
	}
	return renderEnvironmentsPlain(os.Stdout, environments, opts)
}

func renderEnvironmentsPlain(w io.Writer, environments []rollbar.Environment, opts EnvironmentRenderOptions) error {
	fields := opts.Fields
	if len(fields) == 0 {
		fields = []string{"id", "project_id", "environment"}
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if !opts.NoHeaders {
		if _, err := fmt.Fprintln(tw, strings.Join(fieldHeaders(fields), "\t")); err != nil {
			return err
		}
	}
	for _, environment := range environments {
		if _, err := fmt.Fprintln(tw, strings.Join(environmentFieldValues(environment, fields), "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func environmentFieldValues(environment rollbar.Environment, fields []string) []string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		switch field {
		case "id":
			values = append(values, strconv.FormatInt(environment.ID, 10))
		case "project_id":
			values = append(values, strconv.FormatInt(environment.ProjectID, 10))
		case "environment", "name":
			values = append(values, fallback(environment.Name))
		default:
			values = append(values, "-")
		}
	}
	return values
}
