package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultTeamsTimeout = 10 * time.Second

// TeamsSink sends alerts to a Microsoft Teams channel via an incoming webhook.
type TeamsSink struct {
	webhookURL string
	client     *http.Client
}

type teamsPayload struct {
	Text string `json:"text"`
}

// NewTeamsSink constructs a TeamsSink. webhookURL must be non-empty.
func NewTeamsSink(webhookURL string, timeout time.Duration) (*TeamsSink, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("teams: webhookURL must not be empty")
	}
	if timeout <= 0 {
		timeout = defaultTeamsTimeout
	}
	return &TeamsSink{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: timeout},
	}, nil
}

// Send posts the alert as a plain-text Teams message card.
func (s *TeamsSink) Send(a Alert) error {
	body, err := json.Marshal(teamsPayload{
		Text: fmt.Sprintf("**[cronwatch] %s** — job: `%s`  scheduled: %s  detected: %s",
			a.Kind, a.JobName, a.Scheduled.Format(time.RFC3339), a.DetectedAt.Format(time.RFC3339)),
	})
	if err != nil {
		return fmt.Errorf("teams: marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("teams: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("teams: unexpected status %d", resp.StatusCode)
	}
	return nil
}
