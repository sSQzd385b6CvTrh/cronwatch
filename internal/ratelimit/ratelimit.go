// Package ratelimit provides per-job and global alert rate-limiting so that
// a flapping cron job does not flood notification sinks.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a simple cooldown-based rate limiter. After an event is allowed
// through, subsequent calls to Allow return false until the cooldown elapses.
type Limiter struct {
	mu          sync.Mutex
	cooldown    time.Duration
	lastAllowed time.Time
	clock       func() time.Time
}

// New creates a Limiter with the given cooldown. A zero cooldown means every
// call is allowed.
func New(cooldown time.Duration) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		clock:    time.Now,
	}
}

// Allow returns true if the cooldown has elapsed since the last allowed event
// (or if no event has been allowed yet). When true is returned the internal
// timestamp is updated.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock()
	if l.cooldown == 0 {
		l.lastAllowed = now
		return true
	}
	if l.lastAllowed.IsZero() || now.Sub(l.lastAllowed) >= l.cooldown {
		l.lastAllowed = now
		return true
	}
	return false
}

// NextAllowedAt returns the earliest time at which Allow will return true.
// If Allow would return true right now, the zero value is returned.
func (l *Limiter) NextAllowedAt() time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lastAllowed.IsZero() || l.cooldown == 0 {
		return time.Time{}
	}
	next := l.lastAllowed.Add(l.cooldown)
	if l.clock().Before(next) {
		return next
	}
	return time.Time{}
}

// Remaining returns how long until the limiter resets. Returns 0 if already
// allowed.
func (l *Limiter) Remaining() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lastAllowed.IsZero() || l.cooldown == 0 {
		return 0
	}
	d := l.cooldown - l.clock().Sub(l.lastAllowed)
	if d < 0 {
		return 0
	}
	return d
}
