package ui

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"text/tabwriter"

	"rollbar-cli/internal/rollbar"
)

func RenderOccurrences(occurrences []rollbar.ItemInstance) error {
	if len(occurrences) == 0 {
		_, err := fmt.Fprintln(os.Stdout, "No occurrences found.")
		return err
	}
	return renderOccurrencesPlain(os.Stdout, occurrences)
}

func RenderOccurrence(occurrence rollbar.ItemInstance) error {
	return renderItemInstances(os.Stdout, []rollbar.ItemInstance{occurrence})
}

func renderOccurrencesPlain(w io.Writer, occurrences []rollbar.ItemInstance) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tUUID\tLEVEL\tENVIRONMENT\tTIMESTAMP\tSTACK_FRAMES"); err != nil {
		return err
	}
	for _, occurrence := range occurrences {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\t%s\n",
			occurrence.ID,
			fallback(occurrence.UUID),
			fallback(occurrence.Level),
			fallback(occurrence.Environment),
			formatUnix(occurrence.Timestamp),
			strconv.Itoa(len(occurrence.StackFrames)),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
