package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// JobConfig holds the configuration for a single monitored cron job.
type JobConfig struct {
	Name     string        `yaml:"name"`
	Schedule string        `yaml:"schedule"`
	Drift    time.Duration `yaml:"drift"`
}

// NotifierConfig holds sink-specific configuration.
type NotifierConfig struct {
	Log        bool              `yaml:"log"`
	WebhookURL string            `yaml:"webhook_url"`
	SlackURL   string            `yaml:"slack_url"`
	PagerDuty  PagerDutyConfig   `yaml:"pagerduty"`
	Email      EmailConfig       `yaml:"email"`
}

// PagerDutyConfig holds PagerDuty-specific settings.
type PagerDutyConfig struct {
	RoutingKey string        `yaml:"routing_key"`
	Timeout    time.Duration `yaml:"timeout"`
}

// EmailConfig holds SMTP settings.
type EmailConfig struct {
	Host       string   `yaml:"host"`
	Port       int      `yaml:"port"`
	From       string   `yaml:"from"`
	Recipients []string `yaml:"recipients"`
}

// Config is the top-level cronwatch configuration.
type Config struct {
	PollEvery time.Duration  `yaml:"poll_every"`
	Jobs      []JobConfig    `yaml:"jobs"`
	Notifier  NotifierConfig `yaml:"notifier"`
}

const defaultPollEvery = 60 * time.Second

// Load reads and validates a YAML config file from path.
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
