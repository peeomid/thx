package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Format != defaultFormat {
		t.Fatalf("expected default format %q, got %q", defaultFormat, cfg.Format)
	}
	if cfg.Defaults.When != defaultWhen {
		t.Fatalf("expected default when %q, got %q", defaultWhen, cfg.Defaults.When)
	}
}

func TestLoadConfigEnvOverrides(t *testing.T) {
	t.Setenv("THX_FORMAT", "quiet")
	t.Setenv("THX_DEFAULTS_WHEN", "tomorrow")
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Format != "quiet" {
		t.Fatalf("expected format override, got %q", cfg.Format)
	}
	if cfg.Defaults.When != "tomorrow" {
		t.Fatalf("expected when override, got %q", cfg.Defaults.When)
	}
}

func TestLoadConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(path, []byte("format: json\ndefaults:\n  when: today\n  tags: [work]\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Format != "json" {
		t.Fatalf("expected format json, got %q", cfg.Format)
	}
	if cfg.Defaults.When != "today" {
		t.Fatalf("expected when today, got %q", cfg.Defaults.When)
	}
	if len(cfg.Defaults.Tags) != 1 || cfg.Defaults.Tags[0] != "work" {
		t.Fatalf("unexpected tags: %+v", cfg.Defaults.Tags)
	}
}
