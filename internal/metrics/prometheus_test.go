package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWritePrometheus_ContainsExpectedMetrics(t *testing.T) {
	s := New()
	s.IncJobsChecked()
	s.IncJobsChecked()
	s.IncAlertsSent()
	s.IncMissedRuns()
	s.RecordDrift(2 * time.Second)

	var sb strings.Builder
	if err := WritePrometheus(&sb, s.Snapshot()); err != nil {
		t.Fatalf("WritePrometheus error: %v", err)
	}
	out := sb.String()

	cases := []string{
		"cronwatch_jobs_checked_total 2",
		"cronwatch_alerts_sent_total 1",
		"cronwatch_missed_runs_total 1",
		`cronwatch_drift_seconds{sample="0"} 2.000000`,
		"# TYPE cronwatch_drift_seconds gauge",
	}
	for _, want := range cases {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestPrometheusHandler_GET_ReturnsTextPlain(t *testing.T) {
	s := New()
	s.IncJobsChecked()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/prometheus", nil)
	PrometheusHandler(s).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("unexpected Content-Type: %s", ct)
	}
	if !strings.Contains(rec.Body.String(), "cronwatch_jobs_checked_total 1") {
		t.Errorf("body missing counter line:\n%s", rec.Body.String())
	}
}

func TestPrometheusHandler_NonGET_Returns405(t *testing.T) {
	s := New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/metrics/prometheus", nil)
	PrometheusHandler(s).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestWritePrometheus_NoDriftSamples(t *testing.T) {
	s := New()
	var sb strings.Builder
	if err := WritePrometheus(&sb, s.Snapshot()); err != nil {
		t.Fatalf("WritePrometheus error: %v", err)
	}
	out := sb.String()
	if strings.Contains(out, `cronwatch_drift_seconds{sample=`) {
		t.Errorf("expected no drift sample lines, got:\n%s", out)
	}
}
