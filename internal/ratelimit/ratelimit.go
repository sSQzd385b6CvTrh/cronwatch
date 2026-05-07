// Package ratelimit provides per-job alert rate limiting to prevent
// notification storms when a cron job repeatedly misses or drifts.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter suppresses duplicate alerts for the same job within a
// configurable cooldown window.
type Limiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
	now      func() time.Time // injectable for testing
}

// New returns a Limiter that enforces the given cooldown between alerts
// for the same job name. A zero or negative cooldown means every alert
// is allowed through.
func New(cooldown time.Duration) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow returns true if an alert for job should be sent, and records the
// current time as the last alert time for that job. If the cooldown has
// not elapsed since the previous alert, Allow returns false.
func (l *Limiter) Allow(job string) bool {
	if l.cooldown <= 0 {
		return true
	}

	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if t, ok := l.last[job]; ok && now.Sub(t) < l.cooldown {
		return false
	}

	l.last[job] = now
	return true
}

// Reset clears the rate-limit state for a specific job, allowing the
// next alert to be sent immediately regardless of cooldown.
func (l *Limiter) Reset(job string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.last, job)
}

// ResetAll clears rate-limit state for every tracked job.
func (l *Limiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.last = make(map[string]time.Time)
}
