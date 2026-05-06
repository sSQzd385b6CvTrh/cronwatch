package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/example/cronwatch/internal/tracker"
)

const defaultPagerDutyURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutySink sends alerts to PagerDuty via the Events API v2.
type PagerDutySink struct {
	routingKey string
	eventsURL  string
	client     *http.Client
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary  string `json:"summary"`
	Source   string `json:"source"`
	Severity string `json:"severity"`
}

// NewPagerDutySink creates a PagerDutySink. routingKey must be non-empty.
func NewPagerDutySink(routingKey string, timeout time.Duration) (*PagerDutySink, error) {
	if routingKey == "" {
		return nil, fmt.Errorf("pagerduty: routing key must not be empty")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &PagerDutySink{
		routingKey: routingKey,
		eventsURL:  defaultPagerDutyURL,
		client:     &http.Client{Timeout: timeout},
	}, nil
}

// Send delivers the alert to PagerDuty.
func (s *PagerDutySink) Send(a tracker.Alert) error {
	body := pdPayload{
		RoutingKey:  s.routingKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:  fmt.Sprintf("[cronwatch] %s: %s", a.JobName, a.Message),
			Source:   "cronwatch",
			Severity: "error",
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal: %w", err)
	}
	resp, err := s.client.Post(s.eventsURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("pagerduty: request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}
