// Package config loads and validates cronwatch configuration from a YAML file.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// JobConfig describes a single monitored cron job.
type JobConfig struct {
	Name        string        `yaml:"name"`
	Schedule    string        `yaml:"schedule"`
	DriftWindow time.Duration `yaml:"drift_window"`
}

// NotifierConfig holds sink-specific configuration.
type NotifierConfig struct {
	LogEnabled bool   `yaml:"log"`

	WebhookURL string `yaml:"webhook_url"`

	SlackURL string `yaml:"slack_url"`

	EmailHost       string   `yaml:"email_host"`
	EmailPort       int      `yaml:"email_port"`
	EmailFrom       string   `yaml:"email_from"`
	EmailRecipients []string `yaml:"email_recipients"`

	PagerDutyRoutingKey string `yaml:"pagerduty_routing_key"`

	OpsGenieAPIKey string `yaml:"opsgenie_api_key"`
}

// Config is the top-level cronwatch configuration.
type Config struct {
	PollEvery time.Duration  `yaml:"poll_every"`
	Jobs      []JobConfig    `yaml:"jobs"`
	Notifier  NotifierConfig `yaml:"notifier"`
}

const defaultPollEvery = 60 * time.Second

// Load reads and validates a Config from the YAML file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if cfg.PollEvery <= 0 {
		cfg.PollEvery = defaultPollEvery
	}

	if len(cfg.Jobs) == 0 {
		return nil, fmt.Errorf("config: at least one job must be defined")
	}

	for i, j := range cfg.Jobs {
		if j.Name == "" {
			return nil, fmt.Errorf("config: job[%d]: name is required", i)
		}
		if j.Schedule == "" {
			return nil, fmt.Errorf("config: job[%d] %q: schedule is required", i, j.Name)
		}
	}

	return &cfg, nil
}
