package ratelimit

import (
	"testing"
	"time"
)

func quotaWithClock(max int, window time.Duration, now func() time.Time) *QuotaLimiter {
	ql, err := newQuotaLimiterWithClock(max, window, now)
	if err != nil {
		panic(err)
	}
	return ql
}

func TestNewQuotaLimiter_InvalidMax(t *testing.T) {
	_, err := NewQuotaLimiter(0, time.Hour)
	if err == nil {
		t.Fatal("expected error for max=0")
	}
}

func TestNewQuotaLimiter_InvalidWindow(t *testing.T) {
	_, err := NewQuotaLimiter(5, 0)
	if err == nil {
		t.Fatal("expected error for window=0")
	}
}

func TestQuotaLimiter_FirstCallAllowed(t *testing.T) {
	clock := time.Now()
	ql := quotaWithClock(3, time.Hour, func() time.Time { return clock })
	if !ql.Allow("job") {
		t.Fatal("first call should be allowed")
	}
}

func TestQuotaLimiter_AllowsUpToMax(t *testing.T) {
	clock := time.Now()
	ql := quotaWithClock(3, time.Hour, func() time.Time { return clock })
	for i := 0; i < 3; i++ {
		if !ql.Allow("job") {
			t.Fatalf("call %d should be allowed", i+1)
		}
	}
	if ql.Allow("job") {
		t.Fatal("4th call should be blocked")
	}
}

func TestQuotaLimiter_ResetsOnNewWindow(t *testing.T) {
	clock := time.Now().Truncate(time.Hour)
	ql := quotaWithClock(2, time.Hour, func() time.Time { return clock })
	ql.Allow("job")
	ql.Allow("job")
	if ql.Allow("job") {
		t.Fatal("should be blocked in first window")
	}

	// advance past the window boundary
	clock = clock.Add(time.Hour)
	if !ql.Allow("job") {
		t.Fatal("should be allowed in new window")
	}
}

func TestQuotaLimiter_Remaining(t *testing.T) {
	clock := time.Now()
	ql := quotaWithClock(5, time.Hour, func() time.Time { return clock })
	if r := ql.Remaining(); r != 5 {
		t.Fatalf("expected 5 remaining, got %d", r)
	}
	ql.Allow("job")
	ql.Allow("job")
	if r := ql.Remaining(); r != 3 {
		t.Fatalf("expected 3 remaining, got %d", r)
	}
}

func TestQuotaLimiter_Reset(t *testing.T) {
	clock := time.Now()
	ql := quotaWithClock(2, time.Hour, func() time.Time { return clock })
	ql.Allow("job")
	ql.Allow("job")
	ql.Reset()
	if !ql.Allow("job") {
		t.Fatal("should be allowed after manual reset")
	}
}

func TestQuotaLimiter_RemainingAfterWindowExpires(t *testing.T) {
	clock := time.Now().Truncate(time.Hour)
	ql := quotaWithClock(4, time.Hour, func() time.Time { return clock })
	ql.Allow("job")
	ql.Allow("job")
	clock = clock.Add(time.Hour)
	if r := ql.Remaining(); r != 4 {
		t.Fatalf("expected full quota after window reset, got %d", r)
	}
}
