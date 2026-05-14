// Package ratelimit provides various rate-limiting strategies for alert delivery.
package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// QuotaLimiter enforces a maximum number of events per fixed time window (e.g.
// 10 alerts per hour). Unlike SlidingWindow it resets on a hard calendar
// boundary rather than rolling from the first event.
type QuotaLimiter struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	count    int
	windowAt time.Time
	now      func() time.Time
}

// NewQuotaLimiter returns a QuotaLimiter that allows at most max events per
// window duration. Returns an error if max < 1 or window <= 0.
func NewQuotaLimiter(max int, window time.Duration) (*QuotaLimiter, error) {
	return newQuotaLimiterWithClock(max, window, time.Now)
}

func newQuotaLimiterWithClock(max int, window time.Duration, now func() time.Time) (*QuotaLimiter, error) {
	if max < 1 {
		return nil, fmt.Errorf("ratelimit: quota max must be >= 1, got %d", max)
	}
	if window <= 0 {
		return nil, fmt.Errorf("ratelimit: quota window must be positive, got %s", window)
	}
	return &QuotaLimiter{
		max:      max,
		window:   window,
		windowAt: now().Truncate(window),
		now:      now,
	}, nil
}

// Allow returns true if the event is within quota for the current window.
func (q *QuotaLimiter) Allow(key string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.now()
	current := now.Truncate(q.window)
	if current.After(q.windowAt) {
		q.windowAt = current
		q.count = 0
	}

	if q.count >= q.max {
		return false
	}
	q.count++
	return true
}

// Remaining returns the number of events still allowed in the current window.
func (q *QuotaLimiter) Remaining() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.now()
	if now.Truncate(q.window).After(q.windowAt) {
		return q.max
	}
	r := q.max - q.count
	if r < 0 {
		return 0
	}
	return r
}

// Reset forces the quota counter back to zero for the current window.
func (q *QuotaLimiter) Reset() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.count = 0
	q.windowAt = q.now().Truncate(q.window)
}
