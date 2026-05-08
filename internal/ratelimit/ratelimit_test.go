package ratelimit

import (
	"testing"
	"time"
)

type fixedClock struct{ t time.Time }

func (f *fixedClock) Now() time.Time { return f.t }

func limiterWithClock(cooldown time.Duration, clk *fixedClock) *Limiter {
	l := New(cooldown)
	l.clock = clk.Now
	return l
}

func TestAllow_FirstCallAlwaysAllowed(t *testing.T) {
	clk := &fixedClock{t: time.Now()}
	l := limiterWithClock(5*time.Minute, clk)
	if !l.Allow() {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_BlockedWithinCooldown(t *testing.T) {
	clk := &fixedClock{t: time.Now()}
	l := limiterWithClock(5*time.Minute, clk)
	l.Allow()
	clk.t = clk.t.Add(1 * time.Minute)
	if l.Allow() {
		t.Fatal("expected call within cooldown to be blocked")
	}
}

func TestAllow_PermittedAfterCooldown(t *testing.T) {
	clk := &fixedClock{t: time.Now()}
	l := limiterWithClock(5*time.Minute, clk)
	l.Allow()
	clk.t = clk.t.Add(6 * time.Minute)
	if !l.Allow() {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestAllow_ZeroCooldown_AlwaysAllows(t *testing.T) {
	l := New(0)
	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Fatalf("iteration %d: expected zero-cooldown to always allow", i)
		}
	}
}

func TestRemaining_BeforeFirstAllow(t *testing.T) {
	l := New(5 * time.Minute)
	if r := l.Remaining(); r != 0 {
		t.Fatalf("expected 0 remaining before first allow, got %v", r)
	}
}

func TestRemaining_WithinCooldown(t *testing.T) {
	clk := &fixedClock{t: time.Now()}
	l := limiterWithClock(5*time.Minute, clk)
	l.Allow()
	clk.t = clk.t.Add(2 * time.Minute)
	r := l.Remaining()
	if r <= 0 || r > 3*time.Minute+time.Second {
		t.Fatalf("unexpected remaining duration: %v", r)
	}
}

func TestNextAllowedAt_ZeroBeforeFirstAllow(t *testing.T) {
	l := New(5 * time.Minute)
	if !l.NextAllowedAt().IsZero() {
		t.Fatal("expected zero time before first allow")
	}
}

func TestNextAllowedAt_SetAfterAllow(t *testing.T) {
	clk := &fixedClock{t: time.Now()}
	l := limiterWithClock(5*time.Minute, clk)
	l.Allow()
	next := l.NextAllowedAt()
	if next.IsZero() {
		t.Fatal("expected non-zero NextAllowedAt after allow")
	}
	expected := clk.t.Add(5 * time.Minute)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}
