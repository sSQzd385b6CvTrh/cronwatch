package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookSink delivers alert notifications to an HTTP endpoint via POST.
type WebhookSink struct {
	url    string
	client *http.Client
}

// webhookPayload is the JSON body sent to the webhook endpoint.
type webhookPayload struct {
	Job       string    `json:"job"`
	Kind      string    `json:"kind"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// NewWebhookSink creates a WebhookSink that posts to the given URL.
// A zero timeout defaults to 10 seconds.
func NewWebhookSink(url string, timeout time.Duration) *WebhookSink {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookSink{
		url: url,
		client: &http.Client{Timeout: timeout},
	}
}

// Send marshals the alert as JSON and POSTs it to the configured URL.
func (w *WebhookSink) Send(a Alert) error {
	payload := webhookPayload{
		Job:       a.Job,
		Kind:      string(a.Kind),
		Message:   a.Message,
		Timestamp: a.Timestamp,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal: %w", err)
	}

	resp, err := w.client.Post(w.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
