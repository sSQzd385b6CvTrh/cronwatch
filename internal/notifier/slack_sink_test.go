package notifier

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/tracker"
)

func TestNewSlackSink_EmptyURL(t *testing.T) {
	_, err := NewSlackSink("", 0)
	if err == nil {
		t.Fatal("expected error for empty webhook URL")
	}
}

func TestNewSlackSink_DefaultTimeout(t *testing.T) {
	s, err := NewSlackSink("https://hooks.slack.com/test", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.client.Timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", s.client.Timeout)
	}
}

func TestSlackSink_Send_Success(t *testing.T) {
	var gotContentType string
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s, err := NewSlackSink(srv.URL, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := tracker.Alert{
		JobName: "backup",
		Kind:    tracker.AlertDrift,
		Message: "ran 5m late",
		At:      time.Now(),
	}
	if err := s.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected application/json, got %q", gotContentType)
	}
	if len(gotBody) == 0 {
		t.Error("expected non-empty body")
	}
}

func TestSlackSink_Send_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s, err := NewSlackSink(srv.URL, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := tracker.Alert{JobName: "cleanup", Kind: tracker.AlertMissed, Message: "missed run"}
	if err := s.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestSlackSink_Send_Unreachable(t *testing.T) {
	s, err := NewSlackSink("http://127.0.0.1:1", time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := tracker.Alert{JobName: "job", Kind: tracker.AlertMissed, Message: "missed"}
	if err := s.Send(a); err == nil {
		t.Fatal("expected error for unreachable host")
	}
}
