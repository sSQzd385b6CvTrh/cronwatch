package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// splunkEvent is the payload sent to Splunk HTTP Event Collector.
type splunkEvent struct {
	Time   float64        `json:"time"`
	Source string         `json:"source"`
	Event  map[string]any `json:"event"`
}

// SplunkSink delivers alerts to a Splunk HTTP Event Collector endpoint.
type SplunkSink struct {
	url    string
	token  string
	client *http.Client
}

// NewSplunkSink constructs a SplunkSink.
// url must be the full HEC endpoint, e.g. https://splunk:8088/services/collector.
// token is the HEC token used for authentication.
func NewSplunkSink(url, token string, timeout time.Duration) (*SplunkSink, error) {
	if url == "" {
		return nil, fmt.Errorf("splunk: url must not be empty")
	}
	if token == "" {
		return nil, fmt.Errorf("splunk: token must not be empty")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &SplunkSink{
		url:   url,
		token: token,
		client: &http.Client{Timeout: timeout},
	}, nil
}

// Send encodes the Alert as a Splunk HEC event and posts it.
func (s *SplunkSink) Send(a Alert) error {
	payload := splunkEvent{
		Time:   float64(a.At.Unix()),
		Source: "cronwatch",
		Event: map[string]any{
			"job":     a.JobName,
			"kind":    a.Kind,
			"message": a.Message,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("splunk: marshal: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("splunk: build request: %w", err)
	}
	req.Header.Set("Authorization", "Splunk "+s.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("splunk: send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("splunk: unexpected status %d", resp.StatusCode)
	}
	return nil
}
