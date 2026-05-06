package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourorg/cronwatch/internal/tracker"
)

const defaultOpsGenieTimeout = 10 * time.Second
const opsGenieAlertsURL = "https://api.opsgenie.com/v2/alerts"

// OpsGenieSink delivers alerts to OpsGenie via its REST API.
type OpsGenieSink struct {
	apiKey  string
	url     string
	client  *http.Client
}

// NewOpsGenieSink constructs an OpsGenieSink. apiKey must be non-empty.
func NewOpsGenieSink(apiKey string, timeout time.Duration) (*OpsGenieSink, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("opsgenie: API key must not be empty")
	}
	if timeout <= 0 {
		timeout = defaultOpsGenieTimeout
	}
	return &OpsGenieSink{
		apiKey: apiKey,
		url:    opsGenieAlertsURL,
		client: &http.Client{Timeout: timeout},
	}, nil
}

type opsGeniePayload struct {
	Message     string            `json:"message"`
	Description string            `json:"description"`
	Alias       string            `json:"alias"`
	Priority    string            `json:"priority"`
	Details     map[string]string `json:"details"`
}

// Send posts an alert to OpsGenie.
func (s *OpsGenieSink) Send(a tracker.Alert) error {
	payload := opsGeniePayload{
		Message:     fmt.Sprintf("cronwatch: %s — %s", a.JobName, a.Kind),
		Description: a.Message,
		Alias:       fmt.Sprintf("cronwatch-%s-%s", a.JobName, a.Kind),
		Priority:    "P2",
		Details: map[string]string{
			"job":       a.JobName,
			"kind":      string(a.Kind),
			"triggered": a.TriggeredAt.UTC().Format(time.RFC3339),
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: marshal payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("opsgenie: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("opsgenie: send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opsgenie: unexpected status %d", resp.StatusCode)
	}
	return nil
}
