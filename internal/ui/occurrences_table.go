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

func RenderOccurrences(occurrences []rollbar.ItemInstance) error {
	return RenderOccurrencesWithOptions(occurrences, OccurrenceRenderOptions{})
}

type OccurrenceRenderOptions struct {
	Fields    []string
	NoHeaders bool
	Payload   PayloadRenderOptions
}

func RenderOccurrencesWithOptions(occurrences []rollbar.ItemInstance, opts OccurrenceRenderOptions) error {
	if len(occurrences) == 0 {
		_, err := fmt.Fprintln(os.Stdout, "No occurrences found.")
		return err
	}
	return renderOccurrencesPlain(os.Stdout, occurrences, opts)
}

func RenderOccurrence(occurrence rollbar.ItemInstance) error {
	return RenderOccurrenceWithOptions(occurrence, OccurrenceRenderOptions{})
}

func RenderOccurrenceWithOptions(occurrence rollbar.ItemInstance, opts OccurrenceRenderOptions) error {
	return renderItemInstances(os.Stdout, []rollbar.ItemInstance{occurrence}, opts.Payload)
}

func renderOccurrencesPlain(w io.Writer, occurrences []rollbar.ItemInstance, opts OccurrenceRenderOptions) error {
	fields := opts.Fields
	if len(fields) == 0 {
		fields = []string{"id", "uuid", "level", "environment", "timestamp", "stack_frames"}
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if !opts.NoHeaders {
		if _, err := fmt.Fprintln(tw, strings.Join(fieldHeaders(fields), "\t")); err != nil {
			return err
		}
	}
	for _, occurrence := range occurrences {
		if _, err := fmt.Fprintln(tw, strings.Join(occurrenceFieldValues(occurrence, fields), "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func occurrenceFieldValues(occurrence rollbar.ItemInstance, fields []string) []string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		switch field {
		case "id":
			values = append(values, strconv.FormatInt(occurrence.ID, 10))
		case "uuid":
			values = append(values, fallback(occurrence.UUID))
		case "level":
			values = append(values, fallback(occurrence.Level))
		case "environment":
			values = append(values, fallback(occurrence.Environment))
		case "timestamp":
			values = append(values, formatUnix(occurrence.Timestamp))
		case "stack_frames":
			values = append(values, strconv.Itoa(len(occurrence.StackFrames)))
		default:
			values = append(values, "-")
		}
	}
	return values
}
