package ratelimit_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/ratelimit"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAllow_FirstCallAlwaysAllowed(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	if !l.Allow("backup") {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestAllow_BlockedWithinCooldown(t *testing.T) {
	base := time.Now()
	l := ratelimit.New(5 * time.Minute)
	l.(*ratelimit.Limiter) // ensure concrete type if needed — use exported field via helper

	// Use the package-internal clock injection via a test-only constructor.
	l2 := ratelimitNewWithClock(5*time.Minute, fixedClock(base))

	if !l2.Allow("job") {
		t.Fatal("first call should be allowed")
	}
	if l2.Allow("job") {
		t.Fatal("second call within cooldown should be blocked")
	}
}

func TestAllow_PermittedAfterCooldown(t *testing.T) {
	base := time.Now()
	calls := 0
	clock := func() time.Time {
		calls++
		if calls <= 2 {
			return base
		}
		return base.Add(10 * time.Minute)
	}
	l := ratelimitNewWithClock(5*time.Minute, clock)

	l.Allow("job") // call 1 — allowed, records time
	l.Allow("job") // call 2 — blocked (same instant)
	if !l.Allow("job") { // call 3 — 10 min later, should pass
		t.Fatal("expected alert after cooldown elapsed")
	}
}

func TestAllow_ZeroCooldown_AlwaysAllows(t *testing.T) {
	l := ratelimit.New(0)
	for i := 0; i < 5; i++ {
		if !l.Allow("job") {
			t.Fatalf("zero cooldown: call %d should always be allowed", i)
		}
	}
}

func TestReset_ClearsState(t *testing.T) {
	base := time.Now()
	l := ratelimitNewWithClock(5*time.Minute, fixedClock(base))

	l.Allow("job") // record
	l.Reset("job")
	if !l.Allow("job") {
		t.Fatal("expected allow after explicit reset")
	}
}

func TestResetAll_ClearsAllJobs(t *testing.T) {
	base := time.Now()
	l := ratelimitNewWithClock(5*time.Minute, fixedClock(base))

	l.Allow("a")
	l.Allow("b")
	l.ResetAll()

	if !l.Allow("a") || !l.Allow("b") {
		t.Fatal("expected both jobs to be allowed after ResetAll")
	}
}

func TestAllow_IndependentPerJob(t *testing.T) {
	base := time.Now()
	l := ratelimitNewWithClock(5*time.Minute, fixedClock(base))

	l.Allow("a")
	if !l.Allow("b") {
		t.Fatal("different job should not be rate-limited by another job's state")
	}
}

// ratelimitNewWithClock exposes clock injection for tests via the
// unexported now field — accessed through a test helper in the same module.
func ratelimitNewWithClock(d time.Duration, clock func() time.Time) *ratelimit.Limiter {
	l := ratelimit.New(d)
	ratelimit.SetClock(l, clock)
	return l
}
