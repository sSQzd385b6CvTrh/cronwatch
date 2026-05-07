// Package metrics provides in-memory counters and gauges for cronwatch
// runtime statistics such as jobs tracked, alerts fired, and drift samples.
package metrics

import (
	"sync"
	"time"
)

// Snapshot is a point-in-time copy of all collected metrics.
type Snapshot struct {
	JobsTracked   int
	RunsRecorded  int64
	AlertsTotal   int64
	MissedTotal   int64
	DriftSamples  []DriftSample
	CollectedAt   time.Time
}

// DriftSample records a single drift observation.
type DriftSample struct {
	Job       string
	Drift     time.Duration
	RecordedAt time.Time
}

// Collector accumulates runtime metrics.
type Collector struct {
	mu           sync.Mutex
	jobsTracked  int
	runsRecorded int64
	alertsTotal  int64
	missedTotal  int64
	driftSamples []DriftSample
}

// New returns an initialised Collector.
func New() *Collector {
	return &Collector{}
}

// SetJobsTracked replaces the current job count.
func (c *Collector) SetJobsTracked(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.jobsTracked = n
}

// IncRuns increments the total runs counter.
func (c *Collector) IncRuns() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.runsRecorded++
}

// IncAlerts increments the total alerts counter.
func (c *Collector) IncAlerts() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.alertsTotal++
}

// IncMissed increments the missed-jobs counter.
func (c *Collector) IncMissed() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.missedTotal++
}

// RecordDrift appends a drift sample, keeping at most 500 entries.
func (c *Collector) RecordDrift(job string, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.driftSamples) >= 500 {
		c.driftSamples = c.driftSamples[1:]
	}
	c.driftSamples = append(c.driftSamples, DriftSample{
		Job:        job,
		Drift:      d,
		RecordedAt: time.Now().UTC(),
	})
}

// Snapshot returns an immutable copy of current metrics.
func (c *Collector) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	samples := make([]DriftSample, len(c.driftSamples))
	copy(samples, c.driftSamples)
	return Snapshot{
		JobsTracked:  c.jobsTracked,
		RunsRecorded: c.runsRecorded,
		AlertsTotal:  c.alertsTotal,
		MissedTotal:  c.missedTotal,
		DriftSamples: samples,
		CollectedAt:  time.Now().UTC(),
	}
}
