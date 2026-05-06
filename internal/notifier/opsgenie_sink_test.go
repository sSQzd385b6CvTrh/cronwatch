package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/tracker"
)

func TestNewOpsGenieSink_EmptyAPIKey(t *testing.T) {
	_, err := NewOpsGenieSink("", 0)
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestNewOpsGenieSink_DefaultTimeout(t *testing.T) {
	s, err := NewOpsGenieSink("test-key", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.client.Timeout != defaultOpsGenieTimeout {
		t.Errorf("expected default timeout %v, got %v", defaultOpsGenieTimeout, s.client.Timeout)
	}
}

func TestOpsGenieSink_Send_Success(t *testing.T) {
	var received opsGeniePayload
	var authHeader string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	s, _ := NewOpsGenieSink("secret-key", time.Second)
	s.url = srv.URL

	a := tracker.Alert{
		JobName:     "backup",
		Kind:        tracker.AlertDrift,
		Message:     "job ran 5m late",
		TriggeredAt: time.Now(),
	}
	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if authHeader != "GenieKey secret-key" {
		t.Errorf("unexpected auth header: %q", authHeader)
	}
	if received.Message == "" {
		t.Error("expected non-empty message in payload")
	}
	if received.Details["job"] != "backup" {
		t.Errorf("expected job detail 'backup', got %q", received.Details["job"])
	}
}

func TestOpsGenieSink_Send_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	s, _ := NewOpsGenieSink("bad-key", time.Second)
	s.url = srv.URL

	err := s.Send(tracker.Alert{JobName: "job", Kind: tracker.AlertMissed, TriggeredAt: time.Now()})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestOpsGenieSink_Send_Unreachable(t *testing.T) {
	s, _ := NewOpsGenieSink("key", 50*time.Millisecond)
	s.url = "http://127.0.0.1:19999"

	err := s.Send(tracker.Alert{JobName: "job", Kind: tracker.AlertMissed, TriggeredAt: time.Now()})
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
