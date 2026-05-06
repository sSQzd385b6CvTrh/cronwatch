package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewVictorOpsSink_EmptyURL(t *testing.T) {
	_, err := NewVictorOpsSink("", 0)
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewVictorOpsSink_DefaultTimeout(t *testing.T) {
	s, err := NewVictorOpsSink("https://example.com/vo", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.client.Timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", s.client.Timeout)
	}
}

func TestVictorOpsSink_Send_Success(t *testing.T) {
	var received victorOpsPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s, _ := NewVictorOpsSink(ts.URL, 5*time.Second)
	a := Alert{
		JobName: "backup",
		Message: "missed run",
		At:      time.Now(),
	}
	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.EntityID != "backup" {
		t.Errorf("expected entity_id 'backup', got %q", received.EntityID)
	}
	if received.MessageType != "CRITICAL" {
		t.Errorf("expected message_type CRITICAL, got %q", received.MessageType)
	}
}

func TestVictorOpsSink_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s, _ := NewVictorOpsSink(ts.URL, 5*time.Second)
	err := s.Send(Alert{JobName: "job", Message: "msg", At: time.Now()})
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestVictorOpsSink_Send_Unreachable(t *testing.T) {
	s, _ := NewVictorOpsSink("http://127.0.0.1:19999", 500*time.Millisecond)
	err := s.Send(Alert{JobName: "job", Message: "msg", At: time.Now()})
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
