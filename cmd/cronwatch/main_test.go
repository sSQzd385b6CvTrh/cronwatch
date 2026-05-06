package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/cronwatch/internal/config"
)

// writeTemp writes content to a temp file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "cronwatch.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
	return p
}

func TestLoadConfig_IntegrationSmoke(t *testing.T) {
	const yaml = `
poll_every: 30s
jobs:
  - name: backup
    schedule: "0 2 * * *"
    drift_threshold: 5m
  - name: cleanup
    schedule: "*/15 * * * *"
    drift_threshold: 2m
`
	p := writeTemp(t, yaml)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected job[0] name 'backup', got %q", cfg.Jobs[0].Name)
	}
	if cfg.Jobs[1].Name != "cleanup" {
		t.Errorf("expected job[1] name 'cleanup', got %q", cfg.Jobs[1].Name)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/cronwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing config file, got nil")
	}
}

func TestLoadConfig_EmptyJobs(t *testing.T) {
	const yaml = `
poll_every: 1m
jobs: []
`
	p := writeTemp(t, yaml)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected error when jobs list is empty, got nil")
	}
}
