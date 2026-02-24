package cmd

import "testing"

func TestRequireTokenFromEnv(t *testing.T) {
	old := cfg
	defer func() { cfg = old }()

	cfg.Token = ""
	t.Setenv("ROLLBAR_ACCESS_TOKEN", "env-token")

	if err := requireToken(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "env-token" {
		t.Fatalf("expected token from env, got %q", cfg.Token)
	}
}

func TestRequireTokenMissing(t *testing.T) {
	old := cfg
	defer func() { cfg = old }()

	cfg.Token = ""
	t.Setenv("ROLLBAR_ACCESS_TOKEN", "")

	if err := requireToken(); err == nil {
		t.Fatalf("expected missing-token error")
	}
}

func TestRequireTokenKeepsExplicitValue(t *testing.T) {
	old := cfg
	defer func() { cfg = old }()

	cfg.Token = "flag-token"
	t.Setenv("ROLLBAR_ACCESS_TOKEN", "env-token")

	if err := requireToken(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "flag-token" {
		t.Fatalf("expected explicit token to win, got %q", cfg.Token)
	}
}
