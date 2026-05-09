package ratelimit

import (
	"testing"
	"time"
)

func adaptiveWithClock(min, max time.Duration, threshold int, clk func() time.Time) *AdaptiveLimiter {
	return newAdaptiveLimiterWithClock(min, max, threshold, clk)
}

func TestAdaptiveLimiter_FirstCallAllowed(t *testing.T) {
	now := time.Unix(1_000, 0)
	al := adaptiveWithClock(time.Second, 32*time.Second, 3, func() time.Time { return now })
	if !al.Allow("job") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAdaptiveLimiter_BlockedWithinCooldown(t *testing.T) {
	now := time.Unix(1_000, 0)
	al := adaptiveWithClock(4*time.Second, 32*time.Second, 5, func() time.Time { return now })
	al.Allow("job") // seed lastAllowed
	now = now.Add(time.Second) // still within cooldown
	if al.Allow("job") {
		t.Fatal("expected call to be blocked within cooldown")
	}
}

func TestAdaptiveLimiter_PermittedAfterCooldown(t *testing.T) {
	now := time.Unix(1_000, 0)
	al := adaptiveWithClock(2*time.Second, 32*time.Second, 10, func() time.Time { return now })
	al.Allow("job")
	now = now.Add(3 * time.Second)
	if !al.Allow("job") {
		t.Fatal("expected call to be allowed after cooldown elapsed")
	}
}

func TestAdaptiveLimiter_ScalesUpAfterThreshold(t *testing.T) {
	now := time.Unix(0, 0)
	al := adaptiveWithClock(time.Second, 32*time.Second, 2, func() time.Time { return now })

	// Two consecutive allowed calls should trigger scale-up.
	al.Allow("job")             // consecutive=1
	now = now.Add(2 * time.Second) // past cooldown
	al.Allow("job")             // consecutive=2 → scale up, reset

	if al.CurrentCooldown() <= time.Second {
		t.Fatalf("expected cooldown to increase beyond min, got %v", al.CurrentCooldown())
	}
}

func TestAdaptiveLimiter_ScalesDownOnBlock(t *testing.T) {
	now := time.Unix(0, 0)
	// Start with an elevated cooldown.
	al := adaptiveWithClock(time.Second, 32*time.Second, 10, func() time.Time { return now })
	al.Allow("job") // seed
	// Force internal cooldown up.
	al.current = 16 * time.Second

	// A blocked call should scale down toward min.
	now = now.Add(time.Millisecond) // still blocked
	al.Allow("job")

	if al.CurrentCooldown() >= 16*time.Second {
		t.Fatalf("expected cooldown to decrease, got %v", al.CurrentCooldown())
	}
}

func TestAdaptiveLimiter_CooldownCappedAtMax(t *testing.T) {
	now := time.Unix(0, 0)
	al := adaptiveWithClock(time.Second, 4*time.Second, 1, func() time.Time { return now })

	// Every allowed call with threshold=1 doubles the cooldown.
	for i := 0; i < 10; i++ {
		now = now.Add(al.CurrentCooldown() + time.Millisecond)
		al.Allow("job")
	}

	if al.CurrentCooldown() > 4*time.Second {
		t.Fatalf("cooldown exceeded max: %v", al.CurrentCooldown())
	}
}

func TestNewAdaptiveLimiter_InvalidArgs(t *testing.T) {
	_, err := NewAdaptiveLimiter(0, time.Second, 1)
	if err == nil {
		t.Fatal("expected error for zero minCooldown")
	}
	_, err = NewAdaptiveLimiter(time.Second, time.Millisecond, 1)
	if err == nil {
		t.Fatal("expected error when max < min")
	}
}
