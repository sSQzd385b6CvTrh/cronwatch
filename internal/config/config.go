package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job describes a single monitored cron job.
type Job struct {
	Name     string        `yaml:"name"`
	Schedule string        `yaml:"schedule"`
	DriftMax time.Duration `yaml:"drift_max"`
}

// Webhook holds optional webhook notification settings.
type Webhook struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

// Config is the top-level configuration structure.
type Config struct {
	PollEvery time.Duration `yaml:"poll_every"`
	Webhook   *Webhook      `yaml:"webhook,omitempty"`
	Jobs      []Job         `yaml:"jobs"`
}

const defaultPollEvery = 30 * time.Second

// Load reads and validates the YAML config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse: %w", err)
	}

	if len(cfg.Jobs) == 0 {
		return nil, errors.New("config: no jobs defined")
	}
	for i, j := range cfg.Jobs {
		if j.Name == "" {
			return nil, fmt.Errorf("config: job[%d]: missing name", i)
		}
		if j.Schedule == "" {
			return nil, fmt.Errorf("config: job %q: missing schedule", j.Name)
		}
	}

	if cfg.PollEvery == 0 {
		cfg.PollEvery = defaultPollEvery
	}

	if cfg.Webhook != nil {
		if cfg.Webhook.URL == "" {
			return nil, errors.New("config: webhook: url must not be empty")
		}
		if cfg.Webhook.Timeout == 0 {
			cfg.Webhook.Timeout = 10 * time.Second
		}
	}

	return &cfg, nil
}
