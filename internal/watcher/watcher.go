// Package watcher polls the tracker for missed cron jobs and fires alerts
// through the notifier when a scheduled run has not been recorded within
// the expected window.
package watcher

import (
	"context"
	"log"
	"time"

	"github.com/example/cronwatch/internal/notifier"
	"github.com/example/cronwatch/internal/tracker"
)

// Watcher periodically checks every registered job to see whether its last
// recorded run is overdue relative to its cron schedule.
type Watcher struct {
	tracker  *tracker.Tracker
	notifier *notifier.Notifier
	interval time.Duration
	grace    time.Duration
}

// New creates a Watcher that ticks every interval and allows jobs a grace
// period beyond their scheduled time before raising a missed-run alert.
func New(t *tracker.Tracker, n *notifier.Notifier, interval, grace time.Duration) *Watcher {
	return &Watcher{
		tracker:  t,
		notifier: n,
		interval: interval,
		grace:    grace,
	}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			w.check(now)
		}
	}
}

// check inspects every job known to the tracker at the given wall-clock time.
func (w *Watcher) check(now time.Time) {
	missed := w.tracker.MissedJobs(now, w.grace)
	for _, name := range missed {
		msg := notifier.Alert{
			Job:       name,
			Kind:      notifier.KindMissed,
			Message:   "job did not run within the expected window",
			Timestamp: now,
		}
		if err := w.notifier.Notify(ctx, msg); err != nil {
			log.Printf("watcher: failed to send missed-run alert for %q: %v", name, err)
		}
	}
}
