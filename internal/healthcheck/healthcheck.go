// Package healthcheck exposes a simple liveness/readiness probe for cronwatch.
package healthcheck

import (
	"sync"
	"time"
)

// Status represents the overall health of the daemon.
type Status struct {
	Healthy   bool      `json:"healthy"`
	StartedAt time.Time `json:"started_at"`
	CheckedAt time.Time `json:"checked_at"`
	Details   map[string]string `json:"details,omitempty"`
}

// Checker tracks component health and produces a Status snapshot.
type Checker struct {
	mu        sync.RWMutex
	startedAt time.Time
	components map[string]string
}

// New returns a Checker that records the current time as the start time.
func New() *Checker {
	return &Checker{
		startedAt:  time.Now().UTC(),
		components: make(map[string]string),
	}
}

// SetComponent records a named component's health note.
// An empty message means healthy; any non-empty message is treated as degraded.
func (c *Checker) SetComponent(name, message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if message == "" {
		delete(c.components, name)
	} else {
		c.components[name] = message
	}
}

// Check returns the current Status.
func (c *Checker) Check() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()

	details := make(map[string]string, len(c.components))
	for k, v := range c.components {
		details[k] = v
	}

	health := len(c.components) == 0
	if len(details) == 0 {
		details = nil
	}

	return Status{
		Healthy:   health,
		StartedAt: c.startedAt,
		CheckedAt: time.Now().UTC(),
		Details:   details,
	}
}
