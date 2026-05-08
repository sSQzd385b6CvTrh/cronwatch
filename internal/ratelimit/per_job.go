package ratelimit

import (
	"sync"
	"time"
)

// PerJobLimiter maintains an independent rate-limiter for each named job.
// This prevents a single noisy job from consuming the shared alert budget.
type PerJobLimiter struct {
	mu       sync.Mutex
	limiters map[string]*Limiter
	cooldown time.Duration
	clock    func() time.Time
}

// NewPerJobLimiter returns a PerJobLimiter where each job is governed by
// the given cooldown duration.
func NewPerJobLimiter(cooldown time.Duration) *PerJobLimiter {
	return &PerJobLimiter{
		limiters: make(map[string]*Limiter),
		cooldown: cooldown,
		clock:    time.Now,
	}
}

// Allow reports whether an alert for jobName should be delivered.
// The first call for any job is always allowed.
func (p *PerJobLimiter) Allow(jobName string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	l, ok := p.limiters[jobName]
	if !ok {
		l = &Limiter{
			cooldown: p.cooldown,
			clock:    p.clock,
		}
		p.limiters[jobName] = l
	}
	return l.Allow()
}

// Reset clears the rate-limit state for jobName, allowing the next call to
// Allow to pass immediately. This is useful when a job recovers.
func (p *PerJobLimiter) Reset(jobName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.limiters, jobName)
}

// ActiveJobs returns the names of jobs that are currently being throttled
// (i.e. their cooldown has not yet expired).
func (p *PerJobLimiter) ActiveJobs() []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := p.clock()
	var out []string
	for name, l := range p.limiters {
		if !l.lastAllowed.IsZero() && now.Sub(l.lastAllowed) < p.cooldown {
			out = append(out, name)
		}
	}
	return out
}
