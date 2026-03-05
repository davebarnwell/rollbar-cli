package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRequireTokenFromEnv(t *testing.T) {
	cfg := &cliConfig{}
	t.Setenv("ROLLBAR_ACCESS_TOKEN", "env-token")

	if err := requireToken(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "env-token" {
		t.Fatalf("expected token from env, got %q", cfg.Token)
	}
}

func TestRequireTokenMissing(t *testing.T) {
	cfg := &cliConfig{}
	t.Setenv("ROLLBAR_ACCESS_TOKEN", "")

	if err := requireToken(cfg); err == nil {
		t.Fatalf("expected missing-token error")
	}
}

func TestRequireTokenKeepsExplicitValue(t *testing.T) {
	cfg := &cliConfig{Token: "flag-token"}
	t.Setenv("ROLLBAR_ACCESS_TOKEN", "env-token")

	if err := requireToken(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "flag-token" {
		t.Fatalf("expected explicit token to win, got %q", cfg.Token)
	}
}

func TestApplyConfigDefaultsProfile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"default_profile":"prod","profiles":{"prod":{"token":"profile-token","base_url":"https://example.com","timeout":"7s"}}}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg := &cliConfig{
		BaseURL:    defaultBaseURL,
		Timeout:    defaultTimeout,
		ConfigPath: configPath,
	}
	root := newRootCmd()

	if err := applyConfigDefaults(root, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "profile-token" {
		t.Fatalf("expected token from profile, got %q", cfg.Token)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Fatalf("unexpected base url: %q", cfg.BaseURL)
	}
	if cfg.Timeout.String() != "7s" {
		t.Fatalf("unexpected timeout: %s", cfg.Timeout)
	}
}
