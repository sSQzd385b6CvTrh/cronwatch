package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/tracker"
)

func TestNewPagerDutySink_EmptyRoutingKey(t *testing.T) {
	_, err := NewPagerDutySink("", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for empty routing key")
	}
}

func TestNewPagerDutySink_DefaultTimeout(t *testing.T) {
	s, err := NewPagerDutySink("test-key", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.client.Timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", s.client.Timeout)
	}
}

func TestPagerDutySink_Send_Success(t *testing.T) {
	var received pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	s, _ := NewPagerDutySink("rk-abc", 5*time.Second)
	s.eventsURL = ts.URL

	alert := tracker.Alert{JobName: "backup", Message: "missed run"}
	if err := s.Send(alert); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.RoutingKey != "rk-abc" {
		t.Errorf("routing key: got %q, want %q", received.RoutingKey, "rk-abc")
	}
	if received.EventAction != "trigger" {
		t.Errorf("event_action: got %q, want trigger", received.EventAction)
	}
	if received.Payload.Source != "cronwatch" {
		t.Errorf("source: got %q, want cronwatch", received.Payload.Source)
	}
}

func TestPagerDutySink_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	s, _ := NewPagerDutySink("rk-abc", 5*time.Second)
	s.eventsURL = ts.URL

	err := s.Send(tracker.Alert{JobName: "job", Message: "drift"})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestPagerDutySink_Send_Unreachable(t *testing.T) {
	s, _ := NewPagerDutySink("rk-abc", 100*time.Millisecond)
	s.eventsURL = "http://127.0.0.1:1" // nothing listening

	err := s.Send(tracker.Alert{JobName: "job", Message: "missed"})
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
