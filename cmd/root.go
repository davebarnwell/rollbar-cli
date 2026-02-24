package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	cfg cliConfig
)

type cliConfig struct {
	Token   string
	BaseURL string
	Timeout time.Duration
}

var rootCmd = &cobra.Command{
	Use:   "rollbar-cli",
	Short: "A CLI for querying Rollbar data",
	Long:  "rollbar-cli is a command-line utility for querying Rollbar and rendering results as JSON or a TUI table.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfg.Token, "token", "", "Rollbar access token (or set ROLLBAR_ACCESS_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&cfg.BaseURL, "base-url", "https://api.rollbar.com", "Rollbar API base URL")
	rootCmd.PersistentFlags().DurationVar(&cfg.Timeout, "timeout", 15*time.Second, "HTTP timeout")
}

func requireToken() error {
	if cfg.Token == "" {
		cfg.Token = os.Getenv("ROLLBAR_ACCESS_TOKEN")
	}
	if cfg.Token == "" {
		return fmt.Errorf("missing Rollbar token: pass --token or set ROLLBAR_ACCESS_TOKEN")
	}
	return nil
}
