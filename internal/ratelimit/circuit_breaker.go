package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // blocking requests
	StateHalfOpen              // probing for recovery
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker trips open after consecutive failures and resets after a
// cooldown period, allowing a single probe request in half-open state.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failures      int
	threshold    int
	cooldown     time.Duration
	openedAt     time.Time
	now          func() time.Time
}

// NewCircuitBreaker returns a CircuitBreaker that opens after threshold
// consecutive failures and attempts recovery after cooldown.
func NewCircuitBreaker(threshold int, cooldown time.Duration) (*CircuitBreaker, error) {
	if threshold <= 0 {
		return nil, fmt.Errorf("ratelimit: circuit breaker threshold must be > 0")
	}
	if cooldown <= 0 {
		return nil, fmt.Errorf("ratelimit: circuit breaker cooldown must be > 0")
	}
	return &CircuitBreaker{
		threshold: threshold,
		cooldown:  cooldown,
		now:       time.Now,
	}, nil
}

func newCircuitBreakerWithClock(threshold int, cooldown time.Duration, now func() time.Time) *CircuitBreaker {
	cb, _ := NewCircuitBreaker(threshold, cooldown)
	cb.now = now
	return cb
}

// Allow reports whether the call should proceed. It returns false when the
// circuit is open and the cooldown has not yet elapsed.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if cb.now().Sub(cb.openedAt) >= cb.cooldown {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		// Only one probe at a time; block further calls until outcome recorded.
		return false
	}
	return false
}

// RecordSuccess resets the breaker to closed state.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = StateClosed
}

// RecordFailure increments the failure counter and opens the circuit when the
// threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.threshold {
		cb.state = StateOpen
		cb.openedAt = cb.now()
	}
}

// CurrentState returns the current circuit state.
func (cb *CircuitBreaker) CurrentState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
