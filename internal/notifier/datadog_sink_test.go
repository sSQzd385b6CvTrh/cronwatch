package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDatadogSink_EmptyAPIKey(t *testing.T) {
	_, err := NewDatadogSink("", 0)
	if err == nil {
		t.Fatal("expected error for empty api key")
	}
}

func TestNewDatadogSink_DefaultTimeout(t *testing.T) {
	s, err := NewDatadogSink("key123", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.client.Timeout != defaultDatadogTimeout {
		t.Errorf("expected default timeout %v, got %v", defaultDatadogTimeout, s.client.Timeout)
	}
}

func TestDatadogSink_Send_Success(t *testing.T) {
	var received datadogEvent
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("DD-API-KEY") == "" {
			t.Error("missing DD-API-KEY header")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	s, _ := NewDatadogSink("testkey", 5*time.Second)
	s.endpoint = ts.URL

	a := Alert{Job: "backup", Message: "drift detected"}
	if err := s.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if received.Title == "" {
		t.Error("expected non-empty title in payload")
	}
	if received.AlertType != "error" {
		t.Errorf("expected alert_type=error, got %q", received.AlertType)
	}
}

func TestDatadogSink_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	s, _ := NewDatadogSink("testkey", 5*time.Second)
	s.endpoint = ts.URL

	err := s.Send(Alert{Job: "job", Message: "msg"})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestDatadogSink_Send_Unreachable(t *testing.T) {
	s, _ := NewDatadogSink("testkey", 1*time.Second)
	s.endpoint = "http://127.0.0.1:19999"

	err := s.Send(Alert{Job: "job", Message: "msg"})
	if err == nil {
		t.Fatal("expected error for unreachable endpoint")
	}
}
