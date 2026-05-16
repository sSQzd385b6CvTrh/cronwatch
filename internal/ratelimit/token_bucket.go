package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// TokenBucket implements a classic token bucket rate limiter.
// Tokens accumulate at a fixed rate up to a maximum capacity.
// Each Allow call consumes one token; if no tokens are available the
// call is denied without blocking.
type TokenBucket struct {
	mu       sync.Mutex
	clock    func() time.Time
	tokens   float64
	cap      float64
	rate     float64 // tokens per second
	lastTick time.Time
}

// NewTokenBucket creates a TokenBucket with the given capacity and
// refill rate (tokens per second). Both must be positive.
func NewTokenBucket(capacity float64, ratePerSec float64) (*TokenBucket, error) {
	return newTokenBucketWithClock(capacity, ratePerSec, time.Now)
}

func newTokenBucketWithClock(capacity, ratePerSec float64, clock func() time.Time) (*TokenBucket, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("ratelimit: token bucket capacity must be positive, got %v", capacity)
	}
	if ratePerSec <= 0 {
		return nil, fmt.Errorf("ratelimit: token bucket rate must be positive, got %v", ratePerSec)
	}
	return &TokenBucket{
		clock:    clock,
		tokens:   capacity,
		cap:      capacity,
		rate:     ratePerSec,
		lastTick: clock(),
	}, nil
}

// Allow returns true and consumes one token if a token is available.
// It refills tokens proportional to the elapsed time since the last call.
func (tb *TokenBucket) Allow(job string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := tb.clock()
	elapsed := now.Sub(tb.lastTick).Seconds()
	tb.lastTick = now

	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.cap {
		tb.tokens = tb.cap
	}

	if tb.tokens < 1 {
		return false
	}
	tb.tokens--
	return true
}

// Available returns the current token count (floor).
func (tb *TokenBucket) Available() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return int(tb.tokens)
}
