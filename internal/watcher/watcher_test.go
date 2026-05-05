package watcher_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/notifier"
	"github.com/example/cronwatch/internal/tracker"
	"github.com/example/cronwatch/internal/watcher"
)

// captureSink records every alert delivered to it.
type captureSink struct {
	mu     sync.Mutex
	alerts []notifier.Alert
}

func (c *captureSink) Send(_ context.Context, a notifier.Alert) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.alerts = append(c.alerts, a)
	return nil
}

func (c *captureSink) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.alerts)
}

func mustTracker(t *testing.T) *tracker.Tracker {
	t.Helper()
	tr, err := tracker.New()
	if err != nil {
		t.Fatalf("tracker.New: %v", err)
	}
	return tr
}

func TestWatcher_DetectsMissedJob(t *testing.T) {
	tr := mustTracker(t)

	// Register a job that runs every minute.
	if err := tr.Register("cleanup", "* * * * *"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	sink := &captureSink{}
	n := notifier.New(sink)

	// Use a very short interval so the test completes quickly.
	w := watcher.New(tr, n, 20*time.Millisecond, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	go w.Run(ctx)
	<-ctx.Done()

	// The job was never recorded, so at least one missed-run alert must have fired.
	if sink.count() == 0 {
		t.Error("expected at least one missed-run alert, got none")
	}
}

func TestWatcher_NoAlertWhenJobRan(t *testing.T) {
	tr := mustTracker(t)

	if err := tr.Register("heartbeat", "* * * * *"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Record a run right now so the job is considered current.
	if err := tr.RecordRun("heartbeat", time.Now()); err != nil {
		t.Fatalf("RecordRun: %v", err)
	}

	sink := &captureSink{}
	n := notifier.New(sink)

	// Grace period larger than the test window — no alert should fire.
	w := watcher.New(tr, n, 20*time.Millisecond, 5*time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	go w.Run(ctx)
	<-ctx.Done()

	if sink.count() != 0 {
		t.Errorf("expected no alerts, got %d", sink.count())
	}
}
