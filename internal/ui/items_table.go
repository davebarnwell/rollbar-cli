package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"rollbar-cli/internal/rollbar"
)

type model struct {
	table table.Model
}

func RenderItems(items []rollbar.Item) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(os.Stdout, "No items found.")
		return err
	}

	if term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd())) {
		return renderItemsTUI(items)
	}

	return renderItemsPlain(os.Stdout, items)
}

func RenderItem(item rollbar.Item) error {
	if err := renderItem(os.Stdout, item); err != nil {
		return err
	}
	return nil
}

func RenderItemWithInstances(item rollbar.Item, instances []rollbar.ItemInstance) error {
	if err := renderItem(os.Stdout, item); err != nil {
		return err
	}
	return renderItemInstances(os.Stdout, instances)
}

func renderItem(w io.Writer, item rollbar.Item) error {
	if _, err := fmt.Fprintf(w, "ID: %d\n", item.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Counter: %d\n", item.Counter); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Title: %s\n", fallback(item.Title)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Level: %s\n", fallback(item.Level)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Status: %s\n", fallback(item.Status)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Environment: %s\n", fallback(item.Environment)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Total Occurrences: %d\n", item.TotalOccurrences); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Last Seen: %s\n", formatUnix(item.LastOccurrenceTimestamp)); err != nil {
		return err
	}
	return nil
}

func renderItemInstances(w io.Writer, instances []rollbar.ItemInstance) error {
	if _, err := fmt.Fprintf(w, "Instances: %d\n", len(instances)); err != nil {
		return err
	}
	for i, instance := range instances {
		if _, err := fmt.Fprintf(w, "\nInstance #%d\n", i+1); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  ID: %d\n", instance.ID); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  UUID: %s\n", fallback(instance.UUID)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  Level: %s\n", fallback(instance.Level)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  Environment: %s\n", fallback(instance.Environment)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  Timestamp: %s\n", formatUnix(instance.Timestamp)); err != nil {
			return err
		}

		if _, err := fmt.Fprintln(w, "  Stack Frames:"); err != nil {
			return err
		}
		if len(instance.StackFrames) == 0 {
			if _, err := fmt.Fprintln(w, "    -"); err != nil {
				return err
			}
		}
		for _, frame := range instance.StackFrames {
			location := fallback(frame.Filename)
			if frame.Line > 0 {
				location = fmt.Sprintf("%s:%d", location, frame.Line)
			}
			method := fallback(frame.Method)
			if _, err := fmt.Fprintf(w, "    %s (%s)\n", location, method); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintln(w, "  Payload:"); err != nil {
			return err
		}
		if len(instance.Payload) == 0 {
			if _, err := fmt.Fprintln(w, "    -"); err != nil {
				return err
			}
			continue
		}

		payloadJSON, err := json.MarshalIndent(instance.Payload, "", "  ")
		if err != nil {
			return err
		}
		payloadText := indentLines(string(payloadJSON), "    ")
		if _, err := fmt.Fprintln(w, payloadText); err != nil {
			return err
		}
	}

	return nil
}

func renderItemsTUI(items []rollbar.Item) error {
	columns := []table.Column{
		{Title: "Counter", Width: 9},
		{Title: "Level", Width: 8},
		{Title: "Status", Width: 10},
		{Title: "Environment", Width: 14},
		{Title: "Last Seen", Width: 19},
		{Title: "Title", Width: 56},
	}
	rows := make([]table.Row, 0, len(items))
	for _, item := range items {
		rows = append(rows, table.Row{
			strconv.FormatInt(item.Counter, 10),
			item.Level,
			item.Status,
			item.Environment,
			formatUnix(item.LastOccurrenceTimestamp),
			item.Title,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(12, len(rows)+1)),
	)
	styles := table.DefaultStyles()
	styles.Header = styles.Header.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Bold(true)
	styles.Selected = styles.Selected.Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Bold(true)
	t.SetStyles(styles)

	p := tea.NewProgram(model{table: t}, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func renderItemsPlain(w io.Writer, items []rollbar.Item) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "COUNTER\tLEVEL\tSTATUS\tENVIRONMENT\tLAST_SEEN\tTITLE"); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\t%s\n",
			item.Counter,
			item.Level,
			item.Status,
			item.Environment,
			formatUnix(item.LastOccurrenceTimestamp),
			item.Title,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(min(15, msg.Height-4))
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("↑/↓ navigate • q to quit")
	return "\n" + m.table.View() + "\n" + help + "\n"
}

func formatUnix(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func fallback(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

func indentLines(v string, prefix string) string {
	if v == "" {
		return prefix
	}
	return prefix + strings.ReplaceAll(v, "\n", "\n"+prefix)
}
