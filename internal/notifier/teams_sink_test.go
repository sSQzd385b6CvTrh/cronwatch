package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewTeamsSink_EmptyURL(t *testing.T) {
	_, err := NewTeamsSink("", 0)
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewTeamsSink_DefaultTimeout(t *testing.T) {
	s, err := NewTeamsSink("https://example.com/webhook", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.client.Timeout != defaultTeamsTimeout {
		t.Errorf("want %v, got %v", defaultTeamsTimeout, s.client.Timeout)
	}
}

func TestTeamsSink_Send_Success(t *testing.T) {
	var received teamsPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s, _ := NewTeamsSink(srv.URL, 5*time.Second)
	a := Alert{
		JobName:    "nightly-backup",
		Kind:       KindMissed,
		Scheduled:  time.Now().Add(-time.Hour),
		DetectedAt: time.Now(),
	}
	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Text == "" {
		t.Error("expected non-empty text in payload")
	}
}

func TestTeamsSink_Send_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s, _ := NewTeamsSink(srv.URL, 5*time.Second)
	err := s.Send(Alert{JobName: "job", Kind: KindMissed, Scheduled: time.Now(), DetectedAt: time.Now()})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestTeamsSink_Send_Unreachable(t *testing.T) {
	s, _ := NewTeamsSink("http://127.0.0.1:19999", 500*time.Millisecond)
	err := s.Send(Alert{JobName: "job", Kind: KindDrift, Scheduled: time.Now(), DetectedAt: time.Now()})
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
