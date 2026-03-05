package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
)

const (
	outputText    = "text"
	outputJSON    = "json"
	outputRawJSON = "raw-json"
	outputNDJSON  = "ndjson"
)

func resolveOutputMode(output string, jsonShortcut bool, allowed ...string) (string, error) {
	if jsonShortcut {
		output = outputJSON
	}
	output = strings.TrimSpace(strings.ToLower(output))
	if output == "" {
		output = outputText
	}
	if !slices.Contains(allowed, output) {
		return "", fmt.Errorf("invalid --output %q (expected: %s)", output, strings.Join(allowed, "|"))
	}
	return output, nil
}

func resolveOutputModeWithAliases(output string, jsonShortcut bool, rawJSON bool, ndjson bool, allowed ...string) (string, error) {
	aliasCount := 0
	if jsonShortcut {
		aliasCount++
	}
	if rawJSON {
		aliasCount++
	}
	if ndjson {
		aliasCount++
	}
	if aliasCount > 1 {
		return "", fmt.Errorf("use only one of --json, --raw-json, or --ndjson")
	}
	switch {
	case rawJSON:
		output = outputRawJSON
	case ndjson:
		output = outputNDJSON
	case jsonShortcut:
		output = outputJSON
	}
	return resolveOutputMode(output, false, allowed...)
}

func resolveOutput(output string, jsonShortcut bool, rawJSON bool, ndjson bool) (string, error) {
	switch {
	case rawJSON:
		output = outputRawJSON
	case ndjson:
		output = outputNDJSON
	}
	return resolveOutputMode(output, jsonShortcut, outputText, outputJSON, outputRawJSON, outputNDJSON)
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeNDJSON(items []any) error {
	enc := json.NewEncoder(os.Stdout)
	for _, item := range items {
		if err := enc.Encode(item); err != nil {
			return err
		}
	}
	return nil
}

func writeStdoutf(format string, args ...any) error {
	_, err := fmt.Fprintf(os.Stdout, format, args...)
	return err
}
