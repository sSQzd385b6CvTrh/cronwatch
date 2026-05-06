package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookSink_Send_Success(t *testing.T) {
	var received webhookPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type: %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sink := NewWebhookSink(ts.URL, 0)
	a := Alert{
		Job:       "backup",
		Kind:      KindMissed,
		Message:   "job missed",
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	if err := sink.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Job != "backup" {
		t.Errorf("job = %q, want %q", received.Job, "backup")
	}
	if received.Kind != string(KindMissed) {
		t.Errorf("kind = %q, want %q", received.Kind, KindMissed)
	}
	if received.Message != "job missed" {
		t.Errorf("message = %q, want %q", received.Message, "job missed")
	}
}

func TestWebhookSink_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	sink := NewWebhookSink(ts.URL, 0)
	err := sink.Send(Alert{Job: "test", Kind: KindDrift, Message: "drift"})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestWebhookSink_Send_Unreachable(t *testing.T) {
	sink := NewWebhookSink("http://127.0.0.1:0/hook", 500*time.Millisecond)
	err := sink.Send(Alert{Job: "test", Kind: KindMissed, Message: "missed"})
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestNewWebhookSink_DefaultTimeout(t *testing.T) {
	sink := NewWebhookSink("http://example.com", 0)
	if sink.client.Timeout != 10*time.Second {
		t.Errorf("timeout = %v, want 10s", sink.client.Timeout)
	}
}
