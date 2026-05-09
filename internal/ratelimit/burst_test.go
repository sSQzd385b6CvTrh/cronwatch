package ratelimit

import (
	"testing"
	"time"
)

func burstWithClock(burst int, cooldown time.Duration, clk func() time.Time) *BurstLimiter {
	return newBurstLimiterWithClock(burst, cooldown, clk)
}

func TestBurstLimiter_FirstCallAllowed(t *testing.T) {
	bl := NewBurstLimiter(3, time.Minute)
	if !bl.Allow() {
		t.Fatal("expected first call to be allowed")
	}
}

func TestBurstLimiter_AllowsUpToBurst(t *testing.T) {
	bl := NewBurstLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !bl.Allow() {
			t.Fatalf("expected call %d to be allowed", i+1)
		}
	}
	if bl.Allow() {
		t.Fatal("expected call beyond burst to be blocked")
	}
}

func TestBurstLimiter_BlockedDuringCooldown(t *testing.T) {
	now := time.Now()
	bl := burstWithClock(2, 5*time.Second, func() time.Time { return now })
	bl.Allow()
	bl.Allow()

	// Still within cooldown window.
	now = now.Add(3 * time.Second)
	if bl.Allow() {
		t.Fatal("expected call to be blocked during cooldown")
	}
}

func TestBurstLimiter_AllowedAfterCooldown(t *testing.T) {
	now := time.Now()
	bl := burstWithClock(2, 5*time.Second, func() time.Time { return now })
	bl.Allow()
	bl.Allow()

	// Advance past cooldown.
	now = now.Add(6 * time.Second)
	if !bl.Allow() {
		t.Fatal("expected call to be allowed after cooldown")
	}
}

func TestBurstLimiter_CooldownReplenishesFully(t *testing.T) {
	now := time.Now()
	bl := burstWithClock(3, 5*time.Second, func() time.Time { return now })
	for i := 0; i < 3; i++ {
		bl.Allow()
	}
	now = now.Add(6 * time.Second)

	for i := 0; i < 3; i++ {
		if !bl.Allow() {
			t.Fatalf("expected replenished burst call %d to be allowed", i+1)
		}
	}
}

func TestBurstLimiter_Remaining(t *testing.T) {
	bl := NewBurstLimiter(4, time.Minute)
	if bl.Remaining() != 4 {
		t.Fatalf("expected remaining=4, got %d", bl.Remaining())
	}
	bl.Allow()
	if bl.Remaining() != 3 {
		t.Fatalf("expected remaining=3, got %d", bl.Remaining())
	}
}

func TestBurstLimiter_Reset(t *testing.T) {
	bl := NewBurstLimiter(2, time.Hour)
	bl.Allow()
	bl.Allow()
	if bl.Allow() {
		t.Fatal("expected blocked before reset")
	}
	bl.Reset()
	if !bl.Allow() {
		t.Fatal("expected allowed after reset")
	}
}

func TestBurstLimiter_MinBurstEnforced(t *testing.T) {
	bl := NewBurstLimiter(0, time.Minute)
	if !bl.Allow() {
		t.Fatal("expected at least one event to be allowed (min burst=1)")
	}
}
