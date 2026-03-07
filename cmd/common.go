package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/davebarnwell/rollbar-cli/internal/rollbar"
)

type payloadOptions struct {
	Mode       string
	Sections   []string
	MaxBytes   int
	RedactKeys bool
}

func newRollbarClient(cfg *cliConfig) *rollbar.Client {
	return rollbar.NewClient(rollbar.Config{
		AccessToken: cfg.Token,
		BaseURL:     cfg.BaseURL,
		Timeout:     cfg.Timeout,
	})
}

func resolveIdentifierValue(arg string, id int64, uuid string, idSet bool, uuidSet bool, kind string, positionalLabel string, idLabel string, uuidLabel string) (int64, string, error) {
	arg = strings.TrimSpace(arg)
	uuid = strings.TrimSpace(uuid)

	sources := 0
	if arg != "" {
		sources++
	}
	if idSet {
		sources++
	}
	if uuidSet {
		sources++
	}

	if sources == 0 {
		return 0, "", fmt.Errorf("missing %s identifier: pass %s, %s, or %s", kind, positionalLabel, idLabel, uuidLabel)
	}
	if sources > 1 {
		return 0, "", fmt.Errorf("provide only one %s identifier: %s, %s, or %s", kind, positionalLabel, idLabel, uuidLabel)
	}

	if arg != "" {
		if isIntegerToken(arg) {
			n, err := strconv.ParseInt(arg, 10, 64)
			if err != nil || n <= 0 {
				return 0, "", fmt.Errorf("invalid %s id %q: must be > 0", kind, arg)
			}
			return n, "", nil
		}
		return 0, arg, nil
	}

	if idSet {
		if id <= 0 {
			return 0, "", fmt.Errorf("invalid %s id: must be > 0", kind)
		}
		return id, "", nil
	}

	if uuid == "" {
		return 0, "", fmt.Errorf("invalid %s UUID: must not be empty", kind)
	}
	return 0, uuid, nil
}

func isIntegerToken(v string) bool {
	if v == "" {
		return false
	}
	start := 0
	if v[0] == '+' || v[0] == '-' {
		if len(v) == 1 {
			return false
		}
		start = 1
	}
	for i := start; i < len(v); i++ {
		if v[i] < '0' || v[i] > '9' {
			return false
		}
	}
	return true
}

func parseTimeFilter(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t.UTC(), nil
		}
	}
	if isIntegerToken(value) {
		sec, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("parse timestamp %q: %w", value, err)
		}
		return time.Unix(sec, 0).UTC(), nil
	}

	return time.Time{}, fmt.Errorf("unsupported time value %q: use RFC3339, YYYY-MM-DD, or unix seconds", value)
}

func filterItems(items []rollbar.Item, since time.Time, until time.Time, limit int) []rollbar.Item {
	filtered := make([]rollbar.Item, 0, len(items))
	for _, item := range items {
		ts := time.Unix(item.LastOccurrenceTimestamp, 0).UTC()
		if !since.IsZero() && (item.LastOccurrenceTimestamp == 0 || ts.Before(since)) {
			continue
		}
		if !until.IsZero() && item.LastOccurrenceTimestamp > 0 && ts.After(until) {
			continue
		}
		filtered = append(filtered, item)
	}

	if limit > 0 && len(filtered) > limit {
		return filtered[:limit]
	}
	return filtered
}

func sortItems(items []rollbar.Item, sortBy string) {
	switch sortBy {
	case "", "last_seen":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].LastOccurrenceTimestamp > items[j].LastOccurrenceTimestamp
		})
	case "last_seen_asc":
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].LastOccurrenceTimestamp < items[j].LastOccurrenceTimestamp
		})
	case "counter":
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
	}
}

func sortOccurrences(instances []rollbar.ItemInstance) {
	sort.SliceStable(instances, func(i, j int) bool {
		return instances[i].Timestamp > instances[j].Timestamp
	})
}

func limitOccurrences(instances []rollbar.ItemInstance, limit int) []rollbar.ItemInstance {
	if limit > 0 && len(instances) > limit {
		return instances[:limit]
	}
	return instances
}

func applyPayloadOptions(instances []rollbar.ItemInstance, opts payloadOptions) []rollbar.ItemInstance {
	if len(instances) == 0 {
		return nil
	}

	out := make([]rollbar.ItemInstance, 0, len(instances))
	for _, instance := range instances {
		instance.Payload = shapePayload(instance.Payload, opts)
		out = append(out, instance)
	}
	return out
}

func shapePayload(payload map[string]any, opts payloadOptions) map[string]any {
	if len(payload) == 0 {
		return nil
	}
	mode := strings.ToLower(strings.TrimSpace(opts.Mode))
	if mode == "" {
		mode = "summary"
	}
	if mode == "none" {
		return nil
	}

	selected := payload
	if len(opts.Sections) > 0 {
		selected = make(map[string]any)
		for _, section := range opts.Sections {
			if value, ok := payload[section]; ok {
				selected[section] = value
			}
		}
		if len(selected) == 0 {
			return nil
		}
	}

	if opts.RedactKeys {
		selected = redactPayload(selected)
	}
	if mode == "summary" {
		selected = summarizePayload(selected)
	}
	if opts.MaxBytes > 0 {
		return truncatePayload(selected, opts.MaxBytes)
	}
	return selected
}

func redactPayload(payload map[string]any) map[string]any {
	out := make(map[string]any, len(payload))
	for key, value := range payload {
		if shouldRedactKey(key) {
			out[key] = "[REDACTED]"
			continue
		}
		out[key] = redactValue(value)
	}
	return out
}

func redactValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return redactPayload(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, entry := range typed {
			out = append(out, redactValue(entry))
		}
		return out
	default:
		return value
	}
}

func shouldRedactKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, fragment := range []string{"token", "password", "secret", "authorization", "cookie", "api_key", "apikey"} {
		if strings.Contains(key, fragment) {
			return true
		}
	}
	return false
}

func summarizePayload(payload map[string]any) map[string]any {
	out := make(map[string]any, len(payload))
	for key, value := range payload {
		out[key] = summarizeValue(value)
	}
	return out
}

func summarizeValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		if len(keys) > 8 {
			keys = append(keys[:8], "...")
		}
		return map[string]any{
			"kind": "object",
			"keys": keys,
		}
	case []any:
		return map[string]any{
			"kind":   "array",
			"length": len(typed),
		}
	case string:
		return map[string]any{
			"kind":   "string",
			"length": len(typed),
		}
	default:
		return value
	}
}

func truncatePayload(payload map[string]any, maxBytes int) map[string]any {
	raw, err := json.Marshal(payload)
	if err != nil || len(raw) <= maxBytes {
		return payload
	}

	preview := string(raw[:maxBytes])
	return map[string]any{
		"truncated":   true,
		"size_bytes":  len(raw),
		"preview":     preview,
		"preview_len": maxBytes,
	}
}

func renderRows(headers []string, rows [][]string, includeHeaders bool) error {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if includeHeaders {
		if _, err := fmt.Fprintln(tw, strings.Join(headers, "\t")); err != nil {
			return err
		}
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func formatUnix(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}
