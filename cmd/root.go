package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultBaseURL = "https://api.rollbar.com"
	defaultTimeout = 15 * time.Second
)

type cliConfig struct {
	Token      string
	BaseURL    string
	Timeout    time.Duration
	ConfigPath string
	Profile    string
}

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	cfg := &cliConfig{}

	rootCmd := &cobra.Command{
		Use:           "rollbar-cli",
		Short:         "A CLI for querying Rollbar data",
		Long:          "rollbar-cli is a command-line utility for querying Rollbar and rendering results as structured JSON, NDJSON, or terminal views.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return applyConfigDefaults(cmd, cfg)
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfg.Token, "token", "", "Rollbar access token (or set ROLLBAR_ACCESS_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&cfg.BaseURL, "base-url", defaultBaseURL, "Rollbar API base URL")
	rootCmd.PersistentFlags().DurationVar(&cfg.Timeout, "timeout", defaultTimeout, "HTTP timeout")
	rootCmd.PersistentFlags().StringVar(&cfg.ConfigPath, "config", "", "Path to a rollbar-cli JSON config file")
	rootCmd.PersistentFlags().StringVar(&cfg.Profile, "profile", "", "Config profile to use")

	rootCmd.AddCommand(newItemsCmd(cfg))
	rootCmd.AddCommand(newOccurrencesCmd(cfg))
	rootCmd.AddCommand(newCompletionCmd())

	return rootCmd
}

func requireToken(cfg *cliConfig) error {
	if cfg.Token == "" {
		cfg.Token = strings.TrimSpace(os.Getenv("ROLLBAR_ACCESS_TOKEN"))
	}
	if cfg.Token == "" {
		return fmt.Errorf("missing Rollbar token: pass --token, set ROLLBAR_ACCESS_TOKEN, or configure a profile")
	}
	return nil
}
