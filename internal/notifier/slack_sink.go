package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cronwatch/cronwatch/internal/tracker"
)

// SlackSink delivers alert notifications to a Slack incoming webhook URL.
type SlackSink struct {
	webhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// NewSlackSink creates a SlackSink that posts to the given Slack webhook URL.
// An optional timeout may be provided; if zero, a 10-second default is used.
func NewSlackSink(webhookURL string, timeout time.Duration) (*SlackSink, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack sink: webhook URL must not be empty")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &SlackSink{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: timeout},
	}, nil
}

// Send formats the alert as a Slack message and posts it to the webhook.
func (s *SlackSink) Send(a tracker.Alert) error {
	msg := fmt.Sprintf("*[cronwatch]* job=%s kind=%s\n%s", a.JobName, a.Kind, a.Message)
	if !a.At.IsZero() {
		msg += fmt.Sprintf(" (at %s)", a.At.Format(time.RFC3339))
	}

	body, err := json.Marshal(slackPayload{Text: msg})
	if err != nil {
		return fmt.Errorf("slack sink: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack sink: post to webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack sink: unexpected status %d", resp.StatusCode)
	}
	return nil
}
