package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

// jsonSnapshot is the JSON-serialisable form of a Snapshot.
type jsonSnapshot struct {
	JobsTracked  int           `json:"jobs_tracked"`
	RunsRecorded int64         `json:"runs_recorded"`
	AlertsTotal  int64         `json:"alerts_total"`
	MissedTotal  int64         `json:"missed_total"`
	DriftSamples []jsonDrift   `json:"drift_samples"`
	CollectedAt  time.Time     `json:"collected_at"`
}

type jsonDrift struct {
	Job        string        `json:"job"`
	DriftMs    int64         `json:"drift_ms"`
	RecordedAt time.Time     `json:"recorded_at"`
}

// Handler returns an http.Handler that serves the current metrics snapshot
// as JSON on GET requests.
func Handler(c *Collector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s := c.Snapshot()
		out := jsonSnapshot{
			JobsTracked:  s.JobsTracked,
			RunsRecorded: s.RunsRecorded,
			AlertsTotal:  s.AlertsTotal,
			MissedTotal:  s.MissedTotal,
			CollectedAt:  s.CollectedAt,
		}
		for _, d := range s.DriftSamples {
			out.DriftSamples = append(out.DriftSamples, jsonDrift{
				Job:        d.Job,
				DriftMs:    d.Drift.Milliseconds(),
				RecordedAt: d.RecordedAt,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})
}
