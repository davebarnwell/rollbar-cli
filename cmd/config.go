package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type fileConfig struct {
	DefaultProfile string                 `json:"default_profile"`
	Profiles       map[string]fileProfile `json:"profiles"`
}

type fileProfile struct {
	Token   string `json:"token"`
	BaseURL string `json:"base_url"`
	Timeout string `json:"timeout"`
}

func applyConfigDefaults(cmd *cobra.Command, cfg *cliConfig) error {
	profile, err := loadSelectedProfile(cfg)
	if err != nil {
		return err
	}

	if profile != nil {
		if !cmd.Flags().Changed("token") && strings.TrimSpace(cfg.Token) == "" && strings.TrimSpace(profile.Token) != "" {
			cfg.Token = strings.TrimSpace(profile.Token)
		}
		if !cmd.Flags().Changed("base-url") && cfg.BaseURL == defaultBaseURL && strings.TrimSpace(profile.BaseURL) != "" {
			cfg.BaseURL = strings.TrimSpace(profile.BaseURL)
		}
		if !cmd.Flags().Changed("timeout") && cfg.Timeout == defaultTimeout && strings.TrimSpace(profile.Timeout) != "" {
			parsed, err := time.ParseDuration(strings.TrimSpace(profile.Timeout))
			if err != nil {
				return fmt.Errorf("parse timeout for profile %q: %w", cfg.Profile, err)
			}
			cfg.Timeout = parsed
		}
	}

	if !cmd.Flags().Changed("token") && strings.TrimSpace(cfg.Token) == "" {
		cfg.Token = strings.TrimSpace(os.Getenv("ROLLBAR_ACCESS_TOKEN"))
	}
	if !cmd.Flags().Changed("base-url") && cfg.BaseURL == defaultBaseURL {
		if envBaseURL := strings.TrimSpace(os.Getenv("ROLLBAR_BASE_URL")); envBaseURL != "" {
			cfg.BaseURL = envBaseURL
		}
	}
	if !cmd.Flags().Changed("timeout") && cfg.Timeout == defaultTimeout {
		if envTimeout := strings.TrimSpace(os.Getenv("ROLLBAR_TIMEOUT")); envTimeout != "" {
			parsed, err := time.ParseDuration(envTimeout)
			if err != nil {
				return fmt.Errorf("parse ROLLBAR_TIMEOUT: %w", err)
			}
			cfg.Timeout = parsed
		}
	}

	return nil
}

func loadSelectedProfile(cfg *cliConfig) (*fileProfile, error) {
	path, err := resolveConfigPath(cfg)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}

	var fc fileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", path, err)
	}

	profileName := strings.TrimSpace(cfg.Profile)
	if profileName == "" {
		profileName = strings.TrimSpace(fc.DefaultProfile)
	}
	if profileName == "" {
		return nil, nil
	}

	profile, ok := fc.Profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile %q not found in %s", profileName, path)
	}
	cfg.Profile = profileName
	return &profile, nil
}

func resolveConfigPath(cfg *cliConfig) (string, error) {
	if strings.TrimSpace(cfg.ConfigPath) != "" {
		return filepath.Clean(cfg.ConfigPath), nil
	}

	if envPath := strings.TrimSpace(os.Getenv("ROLLBAR_CLI_CONFIG")); envPath != "" {
		return filepath.Clean(envPath), nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	defaultPath := filepath.Join(configDir, "rollbar-cli", "config.json")
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, nil
	}
	if os.IsNotExist(err) {
		return "", nil
	}
	return "", err
}
