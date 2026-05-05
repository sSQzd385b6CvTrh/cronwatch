// Package tracker records cron job execution history and detects
// missed runs or timing drift relative to a parsed schedule.
package tracker

import (
	"fmt"
	"sync"
	"time"

	"github.com/example/cronwatch/internal/schedule"
)

// Entry holds the last-seen execution time and schedule for a named job.
type Entry struct {
	Name     string
	Schedule *schedule.Schedule
	LastRun  time.Time
	Missed   int
}

// Tracker maintains a registry of cron jobs and checks them for drift
// or missed executions.
type Tracker struct {
	mu      sync.Mutex
	entries map[string]*Entry
	// DriftThreshold is the maximum allowed deviation from the expected
	// fire time before an alert is raised.
	DriftThreshold time.Duration
}

// New returns a Tracker with the given drift threshold.
func New(driftThreshold time.Duration) *Tracker {
	return &Tracker{
		entries:        make(map[string]*Entry),
		DriftThreshold: driftThreshold,
	}
}

// Register adds a named job with the given cron expression.
func (t *Tracker) Register(name, expr string) error {
	sched, err := schedule.Parse(expr)
	if err != nil {
		return fmt.Errorf("tracker: register %q: %w", name, err)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries[name] = &Entry{Name: name, Schedule: sched}
	return nil
}

// RecordRun notes that the named job ran at the given time.
// It returns a DriftAlert if the run deviated beyond DriftThreshold.
func (t *Tracker) RecordRun(name string, at time.Time) (*DriftAlert, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[name]
	if !ok {
		return nil, fmt.Errorf("tracker: unknown job %q", name)
	}
	var alert *DriftAlert
	if !e.LastRun.IsZero() {
		expected := e.Schedule.NextAfter(e.LastRun)
		drift := at.Sub(expected)
		if drift < 0 {
			drift = -drift
		}
		if drift > t.DriftThreshold {
			alert = &DriftAlert{
				Job:      name,
				Expected: expected,
				Actual:   at,
				Drift:    at.Sub(expected),
			}
		}
	}
	e.LastRun = at
	e.Missed = 0
	return alert, nil
}

// CheckMissed inspects all registered jobs as of now and returns
// MissedAlert values for any job whose expected fire time has passed
// without a recorded run.
func (t *Tracker) CheckMissed(now time.Time) []MissedAlert {
	t.mu.Lock()
	defer t.mu.Unlock()
	var alerts []MissedAlert
	for _, e := range t.entries {
		if e.LastRun.IsZero() {
			continue
		}
		expected := e.Schedule.NextAfter(e.LastRun)
		if now.After(expected.Add(t.DriftThreshold)) {
			e.Missed++
			alerts = append(alerts, MissedAlert{
				Job:      e.Name,
				Expected: expected,
				MissedAt: now,
			})
		}
	}
	return alerts
}
