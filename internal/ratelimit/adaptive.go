package ratelimit

import (
	"sync"
	"time"
)

// AdaptiveLimiter adjusts its cooldown period dynamically based on the
// observed alert volume. When many alerts fire in quick succession the
// cooldown grows (up to maxCooldown); when traffic is quiet it shrinks
// back toward minCooldown.
type AdaptiveLimiter struct {
	mu          sync.Mutex
	clock       func() time.Time
	minCooldown time.Duration
	maxCooldown time.Duration
	current     time.Duration
	lastAllowed time.Time
	consecutive int // consecutive allowed calls in the current window
	threshold   int // consecutive calls before scaling up
}

// NewAdaptiveLimiter returns an AdaptiveLimiter.
// minCooldown is the baseline; maxCooldown is the ceiling.
// threshold is the number of consecutive allowed calls that triggers a
// cooldown increase.
func NewAdaptiveLimiter(minCooldown, maxCooldown time.Duration, threshold int) (*AdaptiveLimiter, error) {
	if minCooldown <= 0 || maxCooldown < minCooldown {
		return nil, ErrInvalidCooldown
	}
	if threshold < 1 {
		threshold = 1
	}
	return &AdaptiveLimiter{
		clock:       time.Now,
		minCooldown: minCooldown,
		maxCooldown: maxCooldown,
		current:     minCooldown,
		threshold:   threshold,
	}, nil
}

func newAdaptiveLimiterWithClock(min, max time.Duration, threshold int, clk func() time.Time) *AdaptiveLimiter {
	al, _ := NewAdaptiveLimiter(min, max, threshold)
	al.clock = clk
	return al
}

// Allow reports whether the call should be permitted and adapts the
// internal cooldown accordingly.
func (a *AdaptiveLimiter) Allow(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.clock()
	if a.lastAllowed.IsZero() || now.Sub(a.lastAllowed) >= a.current {
		a.lastAllowed = now
		a.consecutive++
		if a.consecutive >= a.threshold {
			a.scaleUp()
			a.consecutive = 0
		}
		return true
	}
	// Blocked — gradually relax toward minCooldown.
	a.scaleDown()
	a.consecutive = 0
	return false
}

// CurrentCooldown returns the active cooldown duration (useful for tests).
func (a *AdaptiveLimiter) CurrentCooldown() time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.current
}

func (a *AdaptiveLimiter) scaleUp() {
	next := a.current * 2
	if next > a.maxCooldown {
		next = a.maxCooldown
	}
	a.current = next
}

func (a *AdaptiveLimiter) scaleDown() {
	next := a.current / 2
	if next < a.minCooldown {
		next = a.minCooldown
	}
	a.current = next
}
