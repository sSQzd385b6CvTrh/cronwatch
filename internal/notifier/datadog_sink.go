package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultDatadogTimeout  = 10 * time.Second
	datadogEventsEndpoint  = "https://api.datadoghq.com/api/v1/events"
)

// DatadogSink sends alert events to the Datadog Events API.
type DatadogSink struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

// NewDatadogSink constructs a DatadogSink. apiKey must be non-empty.
func NewDatadogSink(apiKey string, timeout time.Duration) (*DatadogSink, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("datadog: api key must not be empty")
	}
	if timeout <= 0 {
		timeout = defaultDatadogTimeout
	}
	return &DatadogSink{
		apiKey:   apiKey,
		endpoint: datadogEventsEndpoint,
		client:   &http.Client{Timeout: timeout},
	}, nil
}

type datadogEvent struct {
	Title string   `json:"title"`
	Text  string   `json:"text"`
	Tags  []string `json:"tags"`
	AlertType string `json:"alert_type"`
}

// Send delivers the alert as a Datadog event.
func (d *DatadogSink) Send(a Alert) error {
	payload := datadogEvent{
		Title:     fmt.Sprintf("cronwatch: %s", a.Job),
		Text:      a.Message,
		Tags:      []string{"source:cronwatch", fmt.Sprintf("job:%s", a.Job)},
		AlertType: "error",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("datadog: marshal: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, d.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("datadog: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("datadog: send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("datadog: unexpected status %d", resp.StatusCode)
	}
	return nil
}
