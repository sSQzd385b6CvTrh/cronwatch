// Package ratelimit provides rate limiting primitives for cronwatch alerts.
package ratelimit

import (
	"sync"
	"time"
)

// BurstLimiter allows up to Burst alerts in a row before enforcing a cooldown.
// Once the burst is exhausted the caller must wait until CooldownDuration has
// elapsed since the last allowed event before any further events are permitted.
type BurstLimiter struct {
	mu              sync.Mutex
	burst           int
	cooldown        time.Duration
	remaining        int
	lastAllowedAt   time.Time
	clock           func() time.Time
}

// NewBurstLimiter returns a BurstLimiter that allows up to burst events before
// requiring a cooldown pause of the given duration.
func NewBurstLimiter(burst int, cooldown time.Duration) *BurstLimiter {
	if burst < 1 {
		burst = 1
	}
	return &BurstLimiter{
		burst:     burst,
		cooldown:  cooldown,
		remaining: burst,
		clock:     time.Now,
	}
}

func newBurstLimiterWithClock(burst int, cooldown time.Duration, clk func() time.Time) *BurstLimiter {
	bl := NewBurstLimiter(burst, cooldown)
	bl.clock = clk
	return bl
}

// Allow reports whether the current event should be allowed through.
func (bl *BurstLimiter) Allow() bool {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := bl.clock()

	// If we are in cooldown, check whether it has expired.
	if bl.remaining == 0 {
		if now.Sub(bl.lastAllowedAt) < bl.cooldown {
			return false
		}
		// Cooldown expired — reset the burst counter.
		bl.remaining = bl.burst
	}

	bl.remaining--
	bl.lastAllowedAt = now
	return true
}

// Remaining returns the number of burst slots still available without
// triggering a cooldown.
func (bl *BurstLimiter) Remaining() int {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	return bl.remaining
}

// Reset restores the burst counter to its maximum and clears any active
// cooldown, immediately allowing new events.
func (bl *BurstLimiter) Reset() {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.remaining = bl.burst
	bl.lastAllowedAt = time.Time{}
}
