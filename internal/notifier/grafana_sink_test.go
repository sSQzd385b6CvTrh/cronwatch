package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewGrafanaSink_EmptyURL(t *testing.T) {
	_, err := NewGrafanaSink("", 0)
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewGrafanaSink_DefaultTimeout(t *testing.T) {
	s, err := NewGrafanaSink("http://example.com", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.client.Timeout != defaultGrafanaTimeout {
		t.Errorf("expected default timeout %v, got %v", defaultGrafanaTimeout, s.client.Timeout)
	}
}

func TestGrafanaSink_Send_Success(t *testing.T) {
	var received grafanaPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s, _ := NewGrafanaSink(ts.URL, 5*time.Second)
	a := Alert{JobName: "backup", Message: "drift detected"}
	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.State != "alerting" {
		t.Errorf("expected state 'alerting', got %q", received.State)
	}
	if received.Message != "drift detected" {
		t.Errorf("unexpected message: %q", received.Message)
	}
}

func TestGrafanaSink_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s, _ := NewGrafanaSink(ts.URL, 5*time.Second)
	err := s.Send(Alert{JobName: "job", Message: "msg"})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestGrafanaSink_Send_Unreachable(t *testing.T) {
	s, _ := NewGrafanaSink("http://127.0.0.1:19999", 500*time.Millisecond)
	err := s.Send(Alert{JobName: "job", Message: "msg"})
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
