// Package audit provides a rotating audit log of cron job events
// (registrations, runs, drift alerts, missed-run alerts) for post-hoc
// inspection without requiring an external sink.
package audit

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// EventKind classifies an audit entry.
type EventKind string

const (
	KindRegistered EventKind = "registered"
	KindRun        EventKind = "run"
	KindDrift      EventKind = "drift"
	KindMissed     EventKind = "missed"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Kind      EventKind `json:"kind"`
	Job       string    `json:"job"`
	Message   string    `json:"message,omitempty"`
}

// Logger writes audit entries to a newline-delimited JSON file and keeps an
// in-memory ring buffer of the last maxEntries records.
type Logger struct {
	mu         sync.Mutex
	f          *os.File
	ring       []Entry
	maxEntries int
	head       int
	count      int
}

// New opens (or creates) the file at path and returns a Logger that keeps at
// most maxEntries entries in memory. Pass maxEntries <= 0 to use the default
// of 1000.
func New(path string, maxEntries int) (*Logger, error) {
	if maxEntries <= 0 {
		maxEntries = 1000
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Logger{
		f:          f,
		ring:       make([]Entry, maxEntries),
		maxEntries: maxEntries,
	}, nil
}

// Log records an audit entry. Errors writing to disk are silently dropped so
// that a full disk never crashes the daemon.
func (l *Logger) Log(kind EventKind, job, message string) {
	e := Entry{
		Timestamp: time.Now().UTC(),
		Kind:      kind,
		Job:       job,
		Message:   message,
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	l.ring[l.head] = e
	l.head = (l.head + 1) % l.maxEntries
	if l.count < l.maxEntries {
		l.count++
	}

	if l.f != nil {
		b, _ := json.Marshal(e)
		b = append(b, '\n')
		_, _ = l.f.Write(b)
	}
}

// Recent returns up to n most-recent entries in chronological order.
func (l *Logger) Recent(n int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	if n > l.count {
		n = l.count
	}
	out := make([]Entry, n)
	start := (l.head - n + l.maxEntries) % l.maxEntries
	for i := 0; i < n; i++ {
		out[i] = l.ring[(start+i)%l.maxEntries]
	}
	return out
}

// Close flushes and closes the underlying file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f != nil {
		return l.f.Close()
	}
	return nil
}
