// Package audit implements a lightweight, append-only audit log for cronwatch.
//
// Every significant event in the daemon lifecycle — job registration, run
// recorded, drift detected, missed run — is written as a newline-delimited
// JSON record to a configurable file path and simultaneously stored in an
// in-memory ring buffer for fast, lock-safe retrieval.
//
// # Usage
//
//	l, err := audit.New("/var/log/cronwatch/audit.jsonl", 0)
//	if err != nil { /* handle */ }
//	defer l.Close()
//
//	l.Log(audit.KindRun, "daily-backup", "")
//	l.Log(audit.KindDrift, "daily-backup", "exceeded threshold by 4.2s")
//
//	// Expose over HTTP:
//	http.Handle("/audit", audit.Handler(l))
//
// # Ring buffer
//
// The in-memory ring buffer retains at most maxEntries records (default 1000).
// When the buffer is full, the oldest entry is silently overwritten. Retrieve
// recent entries with Logger.Recent(n).
//
// # HTTP endpoint
//
// Handler returns an http.Handler that serves entries as JSON. Callers may
// filter by kind (?kind=drift) and limit results (?n=50).
package audit
