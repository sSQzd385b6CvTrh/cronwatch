package ratelimit

import (
	"fmt"
	"sync"
)

// PerJobTokenBucket maintains an independent TokenBucket for every
// distinct job name. This lets each cron job consume its own quota
// without being affected by noisy neighbours.
type PerJobTokenBucket struct {
	mu       sync.Mutex
	buckets  map[string]*TokenBucket
	capacity float64
	rate     float64
}

// NewPerJobTokenBucket creates a PerJobTokenBucket where every
// per-job bucket shares the same capacity and refill rate.
func NewPerJobTokenBucket(capacity, ratePerSec float64) (*PerJobTokenBucket, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("ratelimit: per-job token bucket capacity must be positive")
	}
	if ratePerSec <= 0 {
		return nil, fmt.Errorf("ratelimit: per-job token bucket rate must be positive")
	}
	return &PerJobTokenBucket{
		buckets:  make(map[string]*TokenBucket),
		capacity: capacity,
		rate:     ratePerSec,
	}, nil
}

// Allow returns true if the named job has a token available.
// A new bucket (full) is created on the first call for a given job.
func (p *PerJobTokenBucket) Allow(job string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	tb, ok := p.buckets[job]
	if !ok {
		var err error
		tb, err = NewTokenBucket(p.capacity, p.rate)
		if err != nil {
			return false
		}
		p.buckets[job] = tb
	}
	return tb.Allow(job)
}

// Reset removes the bucket for the given job so the next call starts
// fresh. This is a no-op for unknown jobs.
func (p *PerJobTokenBucket) Reset(job string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.buckets, job)
}

// Jobs returns the names of all jobs that currently have a bucket.
func (p *PerJobTokenBucket) Jobs() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	names := make([]string, 0, len(p.buckets))
	for k := range p.buckets {
		names = append(names, k)
	}
	return names
}
