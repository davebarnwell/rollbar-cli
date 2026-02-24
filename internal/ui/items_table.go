package ui

import (
	"fmt"
	"io"
	"os"
	"strconv"
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
		fmt.Println("No items found.")
		return nil
	}

	if term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd())) {
		return renderItemsTUI(items)
	}

	return renderItemsPlain(os.Stdout, items)
}

func RenderItem(item rollbar.Item) error {
	fmt.Printf("ID: %d\n", item.ID)
	fmt.Printf("Counter: %d\n", item.Counter)
	fmt.Printf("Title: %s\n", fallback(item.Title))
	fmt.Printf("Level: %s\n", fallback(item.Level))
	fmt.Printf("Status: %s\n", fallback(item.Status))
	fmt.Printf("Environment: %s\n", fallback(item.Environment))
	fmt.Printf("Total Occurrences: %d\n", item.TotalOccurrences)
	fmt.Printf("Last Seen: %s\n", formatUnix(item.LastOccurrenceTimestamp))
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
