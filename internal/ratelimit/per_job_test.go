package ratelimit

import (
	"sort"
	"testing"
	"time"
)

func newPerJobWithClock(cooldown time.Duration, clk *fixedClock) *PerJobLimiter {
	p := NewPerJobLimiter(cooldown)
	p.clock = clk.Now
	return p
}

func TestReset_AllowsImmediately(t *testing.T) {
	clk := &fixedClock{t: time.Now()}
	p := newPerJobWithClock(10*time.Minute, clk)
	p.Allow("backup")
	p.Reset("backup")
	if !p.Allow("backup") {
		t.Fatal("expected allow after reset")
	}
}

func TestReset_UnknownJobIsNoop(t *testing.T) {
	p := NewPerJobLimiter(5 * time.Minute)
	p.Reset("nonexistent") // must not panic
}

func TestActiveJobs_Empty(t *testing.T) {
	p := NewPerJobLimiter(5 * time.Minute)
	if jobs := p.ActiveJobs(); len(jobs) != 0 {
		t.Fatalf("expected no active jobs, got %v", jobs)
	}
}

func TestActiveJobs_TracksThrottled(t *testing.T) {
	clk := &fixedClock{t: time.Now()}
	p := newPerJobWithClock(10*time.Minute, clk)
	p.Allow("job-a")
	p.Allow("job-b")

	jobs := p.ActiveJobs()
	sort.Strings(jobs)
	if len(jobs) != 2 || jobs[0] != "job-a" || jobs[1] != "job-b" {
		t.Fatalf("unexpected active jobs: %v", jobs)
	}
}

func TestActiveJobs_ExcludesExpired(t *testing.T) {
	clk := &fixedClock{t: time.Now()}
	p := newPerJobWithClock(5*time.Minute, clk)
	p.Allow("job-a")
	p.Allow("job-b")

	// advance past cooldown for job-a only by resetting then re-allowing after expiry
	clk.t = clk.t.Add(6 * time.Minute)
	p.Allow("job-b") // refreshes job-b timestamp

	// job-a's last-allowed is now expired; job-b was just refreshed
	// Re-check: job-a cooldown expired so not in ActiveJobs
	jobs := p.ActiveJobs()
	for _, j := range jobs {
		if j == "job-a" {
			t.Fatal("job-a should not be in active jobs after cooldown expired")
		}
	}
}

func TestPerJobLimiter_IndependentLimits(t *testing.T) {
	clk := &fixedClock{t: time.Now()}
	p := newPerJobWithClock(5*time.Minute, clk)

	if !p.Allow("alpha") {
		t.Fatal("first allow for alpha")
	}
	if !p.Allow("beta") {
		t.Fatal("first allow for beta should be independent")
	}
	// both now throttled
	if p.Allow("alpha") {
		t.Fatal("alpha should be throttled")
	}
	if p.Allow("beta") {
		t.Fatal("beta should be throttled")
	}
}
