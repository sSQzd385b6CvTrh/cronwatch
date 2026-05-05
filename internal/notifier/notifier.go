// Package notifier provides alert delivery mechanisms for cronwatch.
package notifier

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Alert carries information about a cron job anomaly.
type Alert struct {
	JobName   string
	Level     Level
	Message   string
	OccuredAt time.Time
}

// Notifier delivers alerts to one or more sinks.
type Notifier struct {
	sinks []Sink
}

// Sink is the interface implemented by alert destinations.
type Sink interface {
	Send(Alert) error
}

// New creates a Notifier with the provided sinks.
// If no sinks are provided, a LogSink writing to stderr is used.
func New(sinks ...Sink) *Notifier {
	if len(sinks) == 0 {
		sinks = []Sink{NewLogSink(os.Stderr)}
	}
	return &Notifier{sinks: sinks}
}

// Notify sends an alert to all registered sinks.
// Errors from individual sinks are collected and returned as a combined error.
func (n *Notifier) Notify(a Alert) error {
	if a.OccuredAt.IsZero() {
		a.OccuredAt = time.Now()
	}
	var errs []error
	for _, s := range n.sinks {
		if err := s.Send(a); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notifier: %d sink(s) failed: %v", len(errs), errs)
	}
	return nil
}
