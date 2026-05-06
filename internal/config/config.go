// Package config handles loading and validating cronwatch configuration.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job describes a single monitored cron job.
type Job struct {
	Name        string        `yaml:"name"`
	Schedule    string        `yaml:"schedule"`
	DriftLimit  time.Duration `yaml:"drift_limit"`
	MissedAfter time.Duration `yaml:"missed_after"`
}

// Notifier holds sink-specific configuration.
type Notifier struct {
	LogFile string `yaml:"log_file"`
}

// Config is the top-level configuration structure.
type Config struct {
	Jobs      []Job     `yaml:"jobs"`
	Notifier  Notifier  `yaml:"notifier"`
	PollEvery time.Duration `yaml:"poll_every"`
}

// Load reads a YAML config file from path and returns a validated Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Apply sensible defaults.
	if cfg.PollEvery == 0 {
		cfg.PollEvery = time.Minute
	}
	for i := range cfg.Jobs {
		if cfg.Jobs[i].DriftLimit == 0 {
			cfg.Jobs[i].DriftLimit = 5 * time.Minute
		}
		if cfg.Jobs[i].MissedAfter == 0 {
			cfg.Jobs[i].MissedAfter = 2 * cfg.PollEvery
		}
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("config: at least one job must be defined")
	}
	seen := make(map[string]bool, len(c.Jobs))
	for _, j := range c.Jobs {
		if j.Name == "" {
			return fmt.Errorf("config: job missing name")
		}
		if j.Schedule == "" {
			return fmt.Errorf("config: job %q missing schedule", j.Name)
		}
		if seen[j.Name] {
			return fmt.Errorf("config: duplicate job name %q", j.Name)
		}
		seen[j.Name] = true
	}
	return nil
}
