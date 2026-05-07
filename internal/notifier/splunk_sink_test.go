package notifier

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSplunkSink_EmptyURL(t *testing.T) {
	_, err := NewSplunkSink("", "token123", 0)
	if err == nil {
		t.Fatal("expected error for empty url")
	}
}

func TestNewSplunkSink_EmptyToken(t *testing.T) {
	_, err := NewSplunkSink("http://splunk:8088/services/collector", "", 0)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestNewSplunkSink_DefaultTimeout(t *testing.T) {
	s, err := NewSplunkSink("http://splunk:8088/services/collector", "tok", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.client.Timeout != 10*time.Second {
		t.Fatalf("expected default timeout 10s, got %v", s.client.Timeout)
	}
}

func TestSplunkSink_Send_Success(t *testing.T) {
	var gotAuth, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s, err := NewSplunkSink(srv.URL, "mytoken", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := Alert{JobName: "backup", Kind: "missed", Message: "job missed", At: time.Now()}
	if err := s.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotAuth != "Splunk mytoken" {
		t.Errorf("unexpected Authorization header: %q", gotAuth)
	}
	if gotCT != "application/json" {
		t.Errorf("unexpected Content-Type: %q", gotCT)
	}
}

func TestSplunkSink_Send_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	s, _ := NewSplunkSink(srv.URL, "tok", 5*time.Second)
	a := Alert{JobName: "job", Kind: "drift", Message: "late", At: time.Now()}
	if err := s.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestSplunkSink_Send_Unreachable(t *testing.T) {
	s, _ := NewSplunkSink("http://127.0.0.1:19999/collector", "tok", 1*time.Second)
	a := Alert{JobName: "job", Kind: "missed", Message: "gone", At: time.Now()}
	if err := s.Send(a); err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
