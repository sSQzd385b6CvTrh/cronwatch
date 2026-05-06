package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwatch-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTemp(t, `
jobs:
  - name: backup
    schedule: "0 2 * * *"
    drift_limit: 10m
  - name: cleanup
    schedule: "*/15 * * * *"
poll_every: 30s
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Jobs) != 2 {
		t.Fatalf("want 2 jobs, got %d", len(cfg.Jobs))
	}
	if cfg.PollEvery != 30*time.Second {
		t.Errorf("want poll_every=30s, got %v", cfg.PollEvery)
	}
	if cfg.Jobs[0].DriftLimit != 10*time.Minute {
		t.Errorf("want drift_limit=10m, got %v", cfg.Jobs[0].DriftLimit)
	}
	// Default drift_limit applied to second job.
	if cfg.Jobs[1].DriftLimit != 5*time.Minute {
		t.Errorf("want default drift_limit=5m, got %v", cfg.Jobs[1].DriftLimit)
	}
}

func TestLoad_DefaultPollEvery(t *testing.T) {
	path := writeTemp(t, `
jobs:
  - name: sync
    schedule: "* * * * *"
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollEvery != time.Minute {
		t.Errorf("want default poll_every=1m, got %v", cfg.PollEvery)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	path := writeTemp(t, `jobs: []\n`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty jobs")
	}
}

func TestLoad_DuplicateJobName(t *testing.T) {
	path := writeTemp(t, `
jobs:
  - name: dup
    schedule: "* * * * *"
  - name: dup
    schedule: "0 * * * *"
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate job name")
	}
}

func TestLoad_MissingSchedule(t *testing.T) {
	path := writeTemp(t, `
jobs:
  - name: broken
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing schedule")
	}
}
