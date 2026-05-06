package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job describes a single monitored cron job entry.
type Job struct {
	Name      string        `yaml:"name"`
	Schedule  string        `yaml:"schedule"`
	DriftMax  time.Duration `yaml:"drift_max"`
}

// EmailConfig mirrors notifier.EmailConfig for YAML unmarshalling.
type EmailConfig struct {
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to"`
}

// NotifierConfig groups all supported notification sink configurations.
type NotifierConfig struct {
	WebhookURL string       `yaml:"webhook_url"`
	Email      *EmailConfig `yaml:"email,omitempty"`
}

// Config is the top-level cronwatch configuration.
type Config struct {
	PollEvery time.Duration  `yaml:"poll_every"`
	Jobs      []Job          `yaml:"jobs"`
	Notifier  NotifierConfig `yaml:"notifier"`
}

const defaultPollEvery = 60 * time.Second

// Load reads and validates a YAML configuration file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %q: %w", path, err)
	}

	if len(cfg.Jobs) == 0 {
		return nil, fmt.Errorf("config: no jobs defined")
	}
	for i, j := range cfg.Jobs {
		if j.Name == "" {
			return nil, fmt.Errorf("config: job[%d] missing name", i)
		}
		if j.Schedule == "" {
			return nil, fmt.Errorf("config: job %q missing schedule", j.Name)
		}
	}
	if cfg.PollEvery <= 0 {
		cfg.PollEvery = defaultPollEvery
	}
	return &cfg, nil
}
