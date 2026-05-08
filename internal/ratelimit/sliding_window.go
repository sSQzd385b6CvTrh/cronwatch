package ratelimit

import (
	"sync"
	"time"
)

// SlidingWindow is a rate limiter that tracks event timestamps within a rolling
// time window and rejects calls that exceed a maximum count within that window.
type SlidingWindow struct {
	mu       sync.Mutex
	window   time.Duration
	max      int
	timestamps []time.Time
	now      func() time.Time
}

// NewSlidingWindow creates a SlidingWindow that allows at most max events
// within the given rolling window duration.
func NewSlidingWindow(max int, window time.Duration) *SlidingWindow {
	return newSlidingWindowWithClock(max, window, time.Now)
}

func newSlidingWindowWithClock(max int, window time.Duration, now func() time.Time) *SlidingWindow {
	if max <= 0 {
		max = 1
	}
	return &SlidingWindow{
		window: window,
		max:    max,
		now:    now,
	}
}

// Allow returns true and records the event if the caller is within the rate
// limit. Returns false without recording if the limit would be exceeded.
func (s *SlidingWindow) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	cutoff := now.Add(-s.window)

	// Evict timestamps outside the window.
	valid := s.timestamps[:0]
	for _, t := range s.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	s.timestamps = valid

	if len(s.timestamps) >= s.max {
		return false
	}
	s.timestamps = append(s.timestamps, now)
	return true
}

// Count returns the number of events recorded within the current window.
func (s *SlidingWindow) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	cutoff := now.Add(-s.window)
	count := 0
	for _, t := range s.timestamps {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// Reset clears all recorded events.
func (s *SlidingWindow) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timestamps = s.timestamps[:0]
}
