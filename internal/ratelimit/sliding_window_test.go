package ratelimit

import (
	"testing"
	"time"
)

func slidingWithClock(max int, window time.Duration, now func() time.Time) *SlidingWindow {
	return newSlidingWindowWithClock(max, window, now)
}

func TestSlidingWindow_FirstCallAllowed(t *testing.T) {
	sw := NewSlidingWindow(3, time.Minute)
	if !sw.Allow() {
		t.Fatal("expected first call to be allowed")
	}
}

func TestSlidingWindow_AllowsUpToMax(t *testing.T) {
	sw := NewSlidingWindow(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !sw.Allow() {
			t.Fatalf("call %d should be allowed", i+1)
		}
	}
	if sw.Allow() {
		t.Fatal("4th call should be blocked")
	}
}

func TestSlidingWindow_AllowsAfterWindowExpires(t *testing.T) {
	now := time.Unix(1_000, 0)
	clock := func() time.Time { return now }

	sw := slidingWithClock(2, time.Second*10, clock)
	sw.Allow()
	sw.Allow()

	if sw.Allow() {
		t.Fatal("should be blocked before window expires")
	}

	// Advance past the window.
	now = now.Add(11 * time.Second)
	if !sw.Allow() {
		t.Fatal("should be allowed after window expires")
	}
}

func TestSlidingWindow_CountReflectsWindow(t *testing.T) {
	now := time.Unix(1_000, 0)
	clock := func() time.Time { return now }

	sw := slidingWithClock(10, time.Second*10, clock)
	sw.Allow()
	sw.Allow()

	if got := sw.Count(); got != 2 {
		t.Fatalf("expected count 2, got %d", got)
	}

	now = now.Add(11 * time.Second)
	if got := sw.Count(); got != 0 {
		t.Fatalf("expected count 0 after expiry, got %d", got)
	}
}

func TestSlidingWindow_Reset(t *testing.T) {
	sw := NewSlidingWindow(2, time.Minute)
	sw.Allow()
	sw.Allow()

	if sw.Allow() {
		t.Fatal("should be blocked before reset")
	}
	sw.Reset()
	if !sw.Allow() {
		t.Fatal("should be allowed after reset")
	}
}

func TestSlidingWindow_ZeroMaxClamped(t *testing.T) {
	// max=0 is clamped to 1 — one event should be allowed, second blocked.
	sw := NewSlidingWindow(0, time.Minute)
	if !sw.Allow() {
		t.Fatal("first call should be allowed even with zero max")
	}
	if sw.Allow() {
		t.Fatal("second call should be blocked")
	}
}

func TestSlidingWindow_ConcurrentAccess(t *testing.T) {
	sw := NewSlidingWindow(100, time.Minute)
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				sw.Allow()
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 20; i++ {
		<-done
	}
}
