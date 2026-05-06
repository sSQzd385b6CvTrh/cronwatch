package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultGrafanaTimeout = 10 * time.Second

// GrafanaSink sends alerts to Grafana OnCall via its webhook integration.
type GrafanaSink struct {
	url    string
	client *http.Client
}

type grafanaPayload struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	State   string `json:"state"`
}

// NewGrafanaSink constructs a GrafanaSink targeting the given OnCall webhook URL.
func NewGrafanaSink(webhookURL string, timeout time.Duration) (*GrafanaSink, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("grafana: webhook URL must not be empty")
	}
	if timeout <= 0 {
		timeout = defaultGrafanaTimeout
	}
	return &GrafanaSink{
		url: webhookURL,
		client: &http.Client{Timeout: timeout},
	}, nil
}

// Send delivers the alert to Grafana OnCall.
func (g *GrafanaSink) Send(a Alert) error {
	payload := grafanaPayload{
		Title:   fmt.Sprintf("[cronwatch] %s", a.JobName),
		Message: a.Message,
		State:   "alerting",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("grafana: marshal payload: %w", err)
	}
	resp, err := g.client.Post(g.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("grafana: http post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("grafana: unexpected status %d", resp.StatusCode)
	}
	return nil
}
