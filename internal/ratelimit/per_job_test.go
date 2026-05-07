package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestReset_AllowsImmediately(t *testing.T) {
	rl := New(10 * time.Minute)
	rl.Allow("job-a") // consume the first free slot

	if rl.Allow("job-a") {
		t.Fatal("expected second call to be blocked")
	}

	rl.Reset("job-a")

	if !rl.Allow("job-a") {
		t.Fatal("expected allow after reset")
	}
}

func TestReset_UnknownJobIsNoop(t *testing.T) {
	rl := New(5 * time.Minute)
	// resetting an unseen job should not panic
	rl.Reset("nonexistent")

	if !rl.Allow("nonexistent") {
		t.Fatal("first call on unknown job should be allowed")
	}
}

func TestActiveJobs_Empty(t *testing.T) {
	rl := New(5 * time.Minute)
	if got := rl.ActiveJobs(); len(got) != 0 {
		t.Fatalf("expected 0 active jobs, got %d", len(got))
	}
}

func TestActiveJobs_TracksThrottled(t *testing.T) {
	rl := New(10 * time.Minute)
	rl.Allow("job-x") // first call allowed, records timestamp
	rl.Allow("job-x") // second call blocked — job-x is now throttled
	rl.Allow("job-y")
	rl.Allow("job-y")

	active := rl.ActiveJobs()
	if len(active) != 2 {
		t.Fatalf("expected 2 active jobs, got %d: %v", len(active), active)
	}
}

func TestActiveJobs_ExcludesExpired(t *testing.T) {
	clock := &fixedClock{t: time.Now()}
	rl := newWithClock(1*time.Millisecond, clock.now)

	rl.Allow("job-z")
	clock.t = clock.t.Add(10 * time.Millisecond) // advance past cooldown

	active := rl.ActiveJobs()
	if len(active) != 0 {
		t.Fatalf("expected 0 active jobs after expiry, got %d", len(active))
	}
}

func TestActiveJobs_Concurrent(t *testing.T) {
	rl := New(10 * time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			job := "job"
			rl.Allow(job)
			_ = rl.ActiveJobs()
		}(i)
	}
	wg.Wait()
}
