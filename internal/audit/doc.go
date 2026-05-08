// Package audit provides a structured, append-only audit log for cronwatch
// daemon events such as job registrations, missed runs, drift detections, and
// configuration reloads.
//
// # Overview
//
// An [Logger] writes [Entry] values to an underlying file and simultaneously
// retains the most recent entries in an in-memory ring buffer for fast
// retrieval via the HTTP handler.
//
// # Entry kinds
//
// Each [Entry] carries a Kind field that identifies the type of event:
//
//   - KindJobRegistered  – a new cron job was registered with the daemon
//   - KindMissedRun      – a scheduled run was not observed within its window
//   - KindDriftDetected  – execution timing drifted beyond the configured threshold
//   - KindConfigReloaded – the daemon reloaded its configuration from disk
//
// # Ring buffer
//
// The ring buffer (see ring_buffer.go) is a fixed-capacity circular structure.
// When the buffer is full, the oldest entry is silently discarded to make room
// for the new one. The default capacity is 500 entries, matching the drift
// sample cap used by the metrics subsystem.
//
// # HTTP handler
//
// [Handler] exposes a GET endpoint that returns recent audit entries as JSON.
// Callers may filter by kind (?kind=missed_run) and limit the result count
// (?n=50). Only GET requests are accepted; all other methods receive 405.
//
// # Thread safety
//
// All exported methods on [Logger] are safe for concurrent use.
package audit
