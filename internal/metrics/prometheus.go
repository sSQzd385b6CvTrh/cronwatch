// Package metrics provides runtime counters and Prometheus exposition.
package metrics

import (
	"fmt"
	"io"
	"strings"
)

// WritePrometheus writes a Prometheus text-format exposition of s to w.
// It emits the standard cronwatch metric family:
//
//	cronwatch_jobs_checked_total
//	cronwatch_alerts_sent_total
//	cronwatch_missed_runs_total
//	cronwatch_drift_seconds (summary-style, last 500 samples)
func WritePrometheus(w io.Writer, s Snapshot) error {
	lines := []string{
		"# HELP cronwatch_jobs_checked_total Total number of job check evaluations performed.",
		"# TYPE cronwatch_jobs_checked_total counter",
		fmt.Sprintf("cronwatch_jobs_checked_total %d", s.JobsChecked),

		"# HELP cronwatch_alerts_sent_total Total number of alerts dispatched to sinks.",
		"# TYPE cronwatch_alerts_sent_total counter",
		fmt.Sprintf("cronwatch_alerts_sent_total %d", s.AlertsSent),

		"# HELP cronwatch_missed_runs_total Total number of missed cron run detections.",
		"# TYPE cronwatch_missed_runs_total counter",
		fmt.Sprintf("cronwatch_missed_runs_total %d", s.MissedRuns),
	}

	// Emit drift samples as an untyped gauge series with an index label so
	// consumers can feed them into recording rules or histograms downstream.
	lines = append(lines,
		"# HELP cronwatch_drift_seconds Observed drift durations in seconds (ring buffer, newest last).",
		"# TYPE cronwatch_drift_seconds gauge",
	)
	for i, d := range s.DriftSamples {
		lines = append(lines,
			fmt.Sprintf(`cronwatch_drift_seconds{sample="%d"} %.6f`, i, d.Seconds()),
		)
	}

	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}
