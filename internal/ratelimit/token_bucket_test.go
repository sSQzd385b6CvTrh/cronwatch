package ratelimit

import (
	"testing"
	"time"
)

func tokenBucketWithClock(cap, rate float64, clock func() time.Time) *TokenBucket {
	tb, err := newTokenBucketWithClock(cap, rate, clock)
	if err != nil {
		panic(err)
	}
	return tb
}

func TestNewTokenBucket_InvalidCapacity(t *testing.T) {
	_, err := NewTokenBucket(0, 1)
	if err == nil {
		t.Fatal("expected error for zero capacity")
	}
}

func TestNewTokenBucket_InvalidRate(t *testing.T) {
	_, err := NewTokenBucket(5, -1)
	if err == nil {
		t.Fatal("expected error for negative rate")
	}
}

func TestTokenBucket_FirstCallAllowed(t *testing.T) {
	now := time.Now()
	tb := tokenBucketWithClock(5, 1, func() time.Time { return now })
	if !tb.Allow("job") {
		t.Fatal("first call should be allowed")
	}
}

func TestTokenBucket_ExhaustsCapacity(t *testing.T) {
	now := time.Now()
	tb := tokenBucketWithClock(3, 1, func() time.Time { return now })
	for i := 0; i < 3; i++ {
		if !tb.Allow("job") {
			t.Fatalf("call %d should be allowed", i+1)
		}
	}
	if tb.Allow("job") {
		t.Fatal("4th call should be denied — bucket empty")
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	now := time.Now()
	tb := tokenBucketWithClock(2, 1, func() time.Time { return now })
	tb.Allow("job")
	tb.Allow("job")
	if tb.Allow("job") {
		t.Fatal("should be denied when empty")
	}

	// advance clock by 2 seconds → 2 new tokens
	now = now.Add(2 * time.Second)
	if !tb.Allow("job") {
		t.Fatal("should be allowed after refill")
	}
}

func TestTokenBucket_CapNotExceeded(t *testing.T) {
	now := time.Now()
	tb := tokenBucketWithClock(3, 10, func() time.Time { return now })
	// advance by 100 seconds — without cap, tokens would be 1000
	now = now.Add(100 * time.Second)
	tb.Allow("job") // trigger refill
	if tb.Available() > 3 {
		t.Fatalf("available tokens %d exceed capacity 3", tb.Available())
	}
}

func TestTokenBucket_Available_DecreasesOnAllow(t *testing.T) {
	now := time.Now()
	tb := tokenBucketWithClock(5, 1, func() time.Time { return now })
	before := tb.Available()
	tb.Allow("job")
	after := tb.Available()
	if after != before-1 {
		t.Fatalf("expected available %d, got %d", before-1, after)
	}
}
