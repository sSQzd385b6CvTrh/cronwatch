package ratelimit

import (
	"testing"
	"time"
)

func TestNewCircuitBreaker_InvalidThreshold(t *testing.T) {
	_, err := NewCircuitBreaker(0, time.Second)
	if err == nil {
		t.Fatal("expected error for zero threshold")
	}
}

func TestNewCircuitBreaker_InvalidCooldown(t *testing.T) {
	_, err := NewCircuitBreaker(3, 0)
	if err == nil {
		t.Fatal("expected error for zero cooldown")
	}
}

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb, _ := NewCircuitBreaker(3, time.Second)
	if cb.CurrentState() != StateClosed {
		t.Fatalf("expected closed, got %s", cb.CurrentState())
	}
	if !cb.Allow() {
		t.Fatal("expected Allow() == true when closed")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	now := time.Now()
	cb := newCircuitBreakerWithClock(3, 10*time.Second, func() time.Time { return now })

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.CurrentState() != StateOpen {
		t.Fatalf("expected open after threshold failures, got %s", cb.CurrentState())
	}
	if cb.Allow() {
		t.Fatal("expected Allow() == false when open")
	}
}

func TestCircuitBreaker_HalfOpenAfterCooldown(t *testing.T) {
	now := time.Now()
	cb := newCircuitBreakerWithClock(2, 5*time.Second, func() time.Time { return now })

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.CurrentState() != StateOpen {
		t.Fatal("expected open")
	}

	// Advance past cooldown.
	now = now.Add(6 * time.Second)
	if !cb.Allow() {
		t.Fatal("expected Allow() == true in half-open probe")
	}
	if cb.CurrentState() != StateHalfOpen {
		t.Fatalf("expected half-open, got %s", cb.CurrentState())
	}
	// Second call in half-open must be blocked.
	if cb.Allow() {
		t.Fatal("expected second half-open call to be blocked")
	}
}

func TestCircuitBreaker_RecordSuccess_ResetsToClosed(t *testing.T) {
	now := time.Now()
	cb := newCircuitBreakerWithClock(2, 5*time.Second, func() time.Time { return now })

	cb.RecordFailure()
	cb.RecordFailure()
	now = now.Add(6 * time.Second)
	cb.Allow() // transition to half-open

	cb.RecordSuccess()
	if cb.CurrentState() != StateClosed {
		t.Fatalf("expected closed after success, got %s", cb.CurrentState())
	}
	if !cb.Allow() {
		t.Fatal("expected Allow() == true after reset")
	}
}

func TestCircuitBreaker_RecordFailure_ReopensFromHalfOpen(t *testing.T) {
	now := time.Now()
	cb := newCircuitBreakerWithClock(1, 5*time.Second, func() time.Time { return now })

	cb.RecordFailure() // open
	now = now.Add(6 * time.Second)
	cb.Allow() // half-open

	cb.RecordFailure() // re-open
	if cb.CurrentState() != StateOpen {
		t.Fatalf("expected open after probe failure, got %s", cb.CurrentState())
	}
}

func TestState_String(t *testing.T) {
	cases := []struct {
		state State
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.state.String(); got != tc.want {
			t.Errorf("State(%d).String() = %q, want %q", tc.state, got, tc.want)
		}
	}
}
