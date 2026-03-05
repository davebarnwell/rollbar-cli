package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
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

type ItemListRenderOptions struct {
	Fields       []string
	NoHeaders    bool
	Interactions *ItemListInteractions
}

type PayloadRenderOptions struct {
	Mode            string
	Sections        []string
	MaxPayloadBytes int
}

type ItemDetailsRenderOptions struct {
	Payload PayloadRenderOptions
}

type ItemListInteractions struct {
	FetchOccurrences func(item rollbar.Item) ([]rollbar.ItemInstance, error)
	ResolveItem      func(item rollbar.Item) (rollbar.Item, error)
	MuteItem         func(item rollbar.Item) (rollbar.Item, error)
	CopyItemID       func(item rollbar.Item) error
	Payload          PayloadRenderOptions
}

type model struct {
	table           table.Model
	items           []rollbar.Item
	showDetails     bool
	showOccurrences bool
	occurrences     []rollbar.ItemInstance
	statusMessage   string
	interactions    *ItemListInteractions
}

func RenderItems(items []rollbar.Item) error {
	return RenderItemsWithOptions(items, ItemListRenderOptions{})
}

func RenderItemsWithOptions(items []rollbar.Item, opts ItemListRenderOptions) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(os.Stdout, "No items found.")
		return err
	}

	if shouldUseItemTUI(opts) && term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd())) {
		return renderItemsTUI(items, opts)
	}

	return renderItemsPlain(os.Stdout, items, opts)
}

func RenderItem(item rollbar.Item) error {
	return RenderItemWithOptions(item, ItemDetailsRenderOptions{})
}

func RenderItemWithInstances(item rollbar.Item, instances []rollbar.ItemInstance) error {
	return RenderItemWithInstancesOptions(item, instances, ItemDetailsRenderOptions{})
}

func RenderItemWithOptions(item rollbar.Item, opts ItemDetailsRenderOptions) error {
	return renderItemWithOptions(os.Stdout, item, opts)
}

func RenderItemWithInstancesOptions(item rollbar.Item, instances []rollbar.ItemInstance, opts ItemDetailsRenderOptions) error {
	if err := renderItemWithOptions(os.Stdout, item, opts); err != nil {
		return err
	}
	return renderItemInstances(os.Stdout, instances, opts.Payload)
}

func renderItem(w io.Writer, item rollbar.Item) error {
	return renderItemWithOptions(w, item, ItemDetailsRenderOptions{})
}

func renderItemWithOptions(w io.Writer, item rollbar.Item, _ ItemDetailsRenderOptions) error {
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

func renderItemInstances(w io.Writer, instances []rollbar.ItemInstance, payloadOpts PayloadRenderOptions) error {
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
		payload, truncated := formatPayload(instance.Payload, payloadOpts)
		if payload == nil {
			if _, err := fmt.Fprintln(w, "    -"); err != nil {
				return err
			}
			continue
		}

		payloadJSON, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		if truncated {
			payloadJSON = append(payloadJSON, []byte("\n[truncated]")...)
		}
		payloadText := indentLines(string(payloadJSON), "    ")
		if _, err := fmt.Fprintln(w, payloadText); err != nil {
			return err
		}
	}

	return nil
}

func renderItemsTUI(items []rollbar.Item, opts ItemListRenderOptions) error {
	columns := []table.Column{
		{Title: "ID", Width: 10},
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
			strconv.FormatInt(item.ID, 10),
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

	p := tea.NewProgram(model{
		table:        t,
		items:        items,
		interactions: opts.Interactions,
	}, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func renderItemsPlain(w io.Writer, items []rollbar.Item, opts ItemListRenderOptions) error {
	fields := opts.Fields
	if len(fields) == 0 {
		fields = []string{"id", "counter", "level", "status", "environment", "last_seen", "title"}
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if !opts.NoHeaders {
		if _, err := fmt.Fprintln(tw, strings.Join(fieldHeaders(fields), "\t")); err != nil {
			return err
		}
	}
	for _, item := range items {
		if _, err := fmt.Fprintln(tw, strings.Join(itemFieldValues(item, fields), "\t")); err != nil {
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
		case "enter":
			item, ok := m.selectedItem()
			if !ok {
				return m, nil
			}
			if m.interactions == nil || m.interactions.FetchOccurrences == nil {
				m.statusMessage = "occurrence drill-down unavailable"
				return m, nil
			}
			occurrences, err := m.interactions.FetchOccurrences(item)
			if err != nil {
				m.statusMessage = fmt.Sprintf("load occurrences failed: %v", err)
				return m, nil
			}
			m.occurrences = occurrences
			m.showOccurrences = true
			m.statusMessage = fmt.Sprintf("loaded %d occurrences for item %d", len(occurrences), item.ID)
			return m, nil
		case "o":
			m.showDetails = !m.showDetails
			if m.showDetails {
				m.statusMessage = "details opened"
			} else {
				m.statusMessage = "details hidden"
			}
			return m, nil
		case "y":
			item, ok := m.selectedItem()
			if !ok {
				return m, nil
			}
			if err := m.copyItemID(item); err != nil {
				m.statusMessage = fmt.Sprintf("copy failed: %v", err)
			} else {
				m.statusMessage = fmt.Sprintf("copied item id %d", item.ID)
			}
			return m, nil
		case "r":
			return m.applyUpdate(func(item rollbar.Item) (rollbar.Item, error) {
				if m.interactions == nil || m.interactions.ResolveItem == nil {
					return item, fmt.Errorf("resolve unavailable")
				}
				return m.interactions.ResolveItem(item)
			}, "resolved")
		case "m":
			return m.applyUpdate(func(item rollbar.Item) (rollbar.Item, error) {
				if m.interactions == nil || m.interactions.MuteItem == nil {
					return item, fmt.Errorf("mute unavailable")
				}
				return m.interactions.MuteItem(item)
			}, "muted")
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
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("↑/↓ navigate • enter occurrences • o details • y copy id • r resolve • m mute • q quit")
	view := "\n" + m.table.View() + "\n"
	if m.showDetails {
		view += m.detailsView() + "\n"
	}
	if m.showOccurrences {
		view += m.occurrencesView() + "\n"
	}
	if m.statusMessage != "" {
		view += lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(m.statusMessage) + "\n"
	}
	return view + help + "\n"
}

func (m model) detailsView() string {
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.items) {
		return ""
	}
	item := m.items[idx]
	details := []string{
		fmt.Sprintf("ID: %d", item.ID),
		fmt.Sprintf("Counter: %d", item.Counter),
		fmt.Sprintf("Status: %s", fallback(item.Status)),
		fmt.Sprintf("Level: %s", fallback(item.Level)),
		fmt.Sprintf("Environment: %s", fallback(item.Environment)),
		fmt.Sprintf("Last Seen: %s", formatUnix(item.LastOccurrenceTimestamp)),
		fmt.Sprintf("Occurrences: %d", item.TotalOccurrences),
		fmt.Sprintf("Title: %s", fallback(item.Title)),
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Render(strings.Join(details, "\n"))
	return box
}

func (m model) occurrencesView() string {
	lines := []string{"Occurrences"}
	if len(m.occurrences) == 0 {
		lines = append(lines, "None")
	} else {
		limit := min(5, len(m.occurrences))
		for i := 0; i < limit; i++ {
			occurrence := m.occurrences[i]
			lines = append(lines, fmt.Sprintf(
				"%d  %s  %s  %s  %s",
				occurrence.ID,
				fallback(occurrence.UUID),
				fallback(occurrence.Level),
				fallback(occurrence.Environment),
				formatUnix(occurrence.Timestamp),
			))
		}
		if len(m.occurrences) > limit {
			lines = append(lines, fmt.Sprintf("... %d more", len(m.occurrences)-limit))
		}
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
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

func shouldUseItemTUI(opts ItemListRenderOptions) bool {
	return len(opts.Fields) == 0 && !opts.NoHeaders
}

func (m model) selectedItem() (rollbar.Item, bool) {
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.items) {
		return rollbar.Item{}, false
	}
	return m.items[idx], true
}

func (m model) applyUpdate(fn func(item rollbar.Item) (rollbar.Item, error), label string) (tea.Model, tea.Cmd) {
	item, ok := m.selectedItem()
	if !ok {
		return m, nil
	}
	updated, err := fn(item)
	if err != nil {
		m.statusMessage = err.Error()
		return m, nil
	}
	idx := m.table.Cursor()
	m.items[idx] = updated
	rows := m.table.Rows()
	if idx >= 0 && idx < len(rows) {
		rows[idx] = table.Row{
			strconv.FormatInt(updated.ID, 10),
			strconv.FormatInt(updated.Counter, 10),
			updated.Level,
			updated.Status,
			updated.Environment,
			formatUnix(updated.LastOccurrenceTimestamp),
			updated.Title,
		}
		m.table.SetRows(rows)
	}
	m.statusMessage = fmt.Sprintf("item %d %s", updated.ID, label)
	return m, nil
}

func (m model) copyItemID(item rollbar.Item) error {
	if m.interactions != nil && m.interactions.CopyItemID != nil {
		return m.interactions.CopyItemID(item)
	}
	if path, err := exec.LookPath("pbcopy"); err == nil {
		cmd := exec.Command(path)
		cmd.Stdin = strings.NewReader(strconv.FormatInt(item.ID, 10))
		return cmd.Run()
	}
	return fmt.Errorf("clipboard support unavailable")
}

func fieldHeaders(fields []string) []string {
	headers := make([]string, 0, len(fields))
	for _, field := range fields {
		headers = append(headers, strings.ToUpper(field))
	}
	return headers
}

func itemFieldValues(item rollbar.Item, fields []string) []string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		switch field {
		case "id":
			values = append(values, strconv.FormatInt(item.ID, 10))
		case "counter":
			values = append(values, strconv.FormatInt(item.Counter, 10))
		case "level":
			values = append(values, fallback(item.Level))
		case "status":
			values = append(values, fallback(item.Status))
		case "environment":
			values = append(values, fallback(item.Environment))
		case "last_seen":
			values = append(values, formatUnix(item.LastOccurrenceTimestamp))
		case "title":
			values = append(values, fallback(item.Title))
		case "total_occurrences":
			values = append(values, strconv.FormatInt(item.TotalOccurrences, 10))
		default:
			values = append(values, "-")
		}
	}
	return values
}

func formatPayload(payload map[string]any, opts PayloadRenderOptions) (map[string]any, bool) {
	mode := strings.TrimSpace(strings.ToLower(opts.Mode))
	switch mode {
	case "", "summary":
		payload = selectPayloadSections(payload, opts.Sections)
	case "none":
		return nil, false
	case "full":
		payload = selectPayloadSections(payload, opts.Sections)
	default:
		payload = selectPayloadSections(payload, opts.Sections)
	}
	if len(payload) == 0 {
		return nil, false
	}

	if mode == "" || mode == "summary" {
		payload = summarizePayload(payload)
	}

	if opts.MaxPayloadBytes <= 0 {
		return payload, false
	}

	raw, err := json.Marshal(payload)
	if err != nil || len(raw) <= opts.MaxPayloadBytes {
		return payload, false
	}

	truncated := make(map[string]any, len(payload))
	for k, v := range payload {
		truncated[k] = v
	}
	truncated["_truncated"] = fmt.Sprintf("payload exceeded %d bytes", opts.MaxPayloadBytes)
	return truncated, true
}

func selectPayloadSections(payload map[string]any, sections []string) map[string]any {
	if len(payload) == 0 {
		return nil
	}
	if len(sections) == 0 {
		return payload
	}
	selected := make(map[string]any)
	for _, section := range sections {
		if value, ok := payload[section]; ok {
			selected[section] = value
		}
	}
	return selected
}

func summarizePayload(payload map[string]any) map[string]any {
	summary := make(map[string]any, len(payload))
	for key, value := range payload {
		switch t := value.(type) {
		case map[string]any:
			summary[key] = summarizeMap(t)
		case []any:
			summary[key] = fmt.Sprintf("%d values", len(t))
		default:
			summary[key] = value
		}
	}
	return summary
}

func summarizeMap(m map[string]any) map[string]any {
	summary := make(map[string]any, len(m))
	for key, value := range m {
		switch t := value.(type) {
		case string:
			summary[key] = truncateString(t, 120)
		case []any:
			summary[key] = fmt.Sprintf("%d values", len(t))
		case map[string]any:
			summary[key] = fmt.Sprintf("%d fields", len(t))
		default:
			summary[key] = value
		}
	}
	return summary
}

func truncateString(v string, max int) string {
	if max <= 0 || len(v) <= max {
		return v
	}
	return v[:max] + "..."
}
