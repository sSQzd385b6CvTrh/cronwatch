package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_GET_ReturnsJSON(t *testing.T) {
	c := New()
	c.SetJobsTracked(4)
	c.IncRuns()
	c.IncAlerts()
	c.IncMissed()
	c.RecordDrift("nightly", 3*time.Second)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	Handler(c).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type: want application/json, got %s", ct)
	}

	var out jsonSnapshot
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if out.JobsTracked != 4 {
		t.Errorf("jobs_tracked: want 4, got %d", out.JobsTracked)
	}
	if out.RunsRecorded != 1 {
		t.Errorf("runs_recorded: want 1, got %d", out.RunsRecorded)
	}
	if out.AlertsTotal != 1 {
		t.Errorf("alerts_total: want 1, got %d", out.AlertsTotal)
	}
	if out.MissedTotal != 1 {
		t.Errorf("missed_total: want 1, got %d", out.MissedTotal)
	}
	if len(out.DriftSamples) != 1 {
		t.Fatalf("want 1 drift sample, got %d", len(out.DriftSamples))
	}
	if out.DriftSamples[0].Job != "nightly" {
		t.Errorf("drift job: want nightly, got %s", out.DriftSamples[0].Job)
	}
	if out.DriftSamples[0].DriftMs != 3000 {
		t.Errorf("drift_ms: want 3000, got %d", out.DriftSamples[0].DriftMs)
	}
}

func TestHandler_NonGET_Returns405(t *testing.T) {
	c := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	Handler(c).ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("want 405, got %d", rec.Code)
	}
}

func TestHandler_EmptyDriftSamples(t *testing.T) {
	c := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	Handler(c).ServeHTTP(rec, req)

	var out jsonSnapshot
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.DriftSamples != nil {
		t.Errorf("expected nil drift_samples for empty collector")
	}
}
