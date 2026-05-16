package ratelimit

import (
	"testing"
)

func TestNewPerJobTokenBucket_InvalidCapacity(t *testing.T) {
	_, err := NewPerJobTokenBucket(0, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewPerJobTokenBucket_InvalidRate(t *testing.T) {
	_, err := NewPerJobTokenBucket(5, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPerJobTokenBucket_FirstCallAllowed(t *testing.T) {
	p, _ := NewPerJobTokenBucket(5, 1)
	if !p.Allow("backup") {
		t.Fatal("first call should be allowed")
	}
}

func TestPerJobTokenBucket_IndependentBuckets(t *testing.T) {
	p, _ := NewPerJobTokenBucket(1, 1)
	p.Allow("job-a") // exhausts job-a
	if !p.Allow("job-b") {
		t.Fatal("job-b should be unaffected by job-a exhaustion")
	}
	if p.Allow("job-a") {
		t.Fatal("job-a should still be denied")
	}
}

func TestPerJobTokenBucket_Reset_AllowsImmediately(t *testing.T) {
	p, _ := NewPerJobTokenBucket(1, 1)
	p.Allow("job") // exhaust
	if p.Allow("job") {
		t.Fatal("should be denied before reset")
	}
	p.Reset("job")
	if !p.Allow("job") {
		t.Fatal("should be allowed after reset")
	}
}

func TestPerJobTokenBucket_Reset_UnknownJobIsNoop(t *testing.T) {
	p, _ := NewPerJobTokenBucket(5, 1)
	p.Reset("nonexistent") // must not panic
}

func TestPerJobTokenBucket_Jobs_TracksCreatedBuckets(t *testing.T) {
	p, _ := NewPerJobTokenBucket(5, 1)
	p.Allow("alpha")
	p.Allow("beta")
	jobs := p.Jobs()
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestPerJobTokenBucket_Jobs_EmptyInitially(t *testing.T) {
	p, _ := NewPerJobTokenBucket(5, 1)
	if len(p.Jobs()) != 0 {
		t.Fatal("expected no jobs initially")
	}
}
