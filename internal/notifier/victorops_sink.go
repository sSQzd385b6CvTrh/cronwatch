package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// VictorOpsSink sends alerts to a VictorOps (Splunk On-Call) REST endpoint.
type VictorOpsSink struct {
	restEndpointURL string
	client          *http.Client
}

type victorOpsPayload struct {
	MessageType       string `json:"message_type"`
	EntityID          string `json:"entity_id"`
	EntityDisplayName string `json:"entity_display_name"`
	StateMessage      string `json:"state_message"`
	Timestamp         int64  `json:"timestamp"`
}

// NewVictorOpsSink constructs a VictorOpsSink.
// restEndpointURL is the full VictorOps REST endpoint URL including the routing key.
func NewVictorOpsSink(restEndpointURL string, timeout time.Duration) (*VictorOpsSink, error) {
	if restEndpointURL == "" {
		return nil, fmt.Errorf("victorops: rest endpoint URL must not be empty")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &VictorOpsSink{
		restEndpointURL: restEndpointURL,
		client:          &http.Client{Timeout: timeout},
	}, nil
}

// Send delivers an Alert to VictorOps.
func (v *VictorOpsSink) Send(a Alert) error {
	payload := victorOpsPayload{
		MessageType:       "CRITICAL",
		EntityID:          a.JobName,
		EntityDisplayName: fmt.Sprintf("cronwatch: %s", a.JobName),
		StateMessage:      a.Message,
		Timestamp:         a.At.Unix(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("victorops: marshal payload: %w", err)
	}
	resp, err := v.client.Post(v.restEndpointURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("victorops: http post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("victorops: unexpected status %d", resp.StatusCode)
	}
	return nil
}
