// Package metrics provides lightweight, concurrency-safe counters and gauges
// for the cronwatch daemon.
//
// A Collector accumulates four runtime statistics:
//
//   - JobsTracked   – current number of registered cron jobs (gauge).
//   - RunsRecorded  – total number of job-run events received (counter).
//   - AlertsTotal   – total number of alerts dispatched (counter).
//   - MissedTotal   – total number of missed-job detections (counter).
//
// Additionally, up to 500 DriftSample entries are kept in a ring-like buffer
// so operators can inspect recent drift history without an external time-series
// store.
//
// The Handler function exposes a Snapshot as JSON over HTTP and is intended to
// be mounted at a diagnostics endpoint (e.g. /metrics or /debug/metrics) by
// the main daemon.
//
// Example:
//
//	col := metrics.New()
//	http.Handle("/metrics", metrics.Handler(col))
package metrics
