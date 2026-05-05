package tracker_test

import (
	"testing"
	"time"

	"github.com/example/cronwatch/internal/tracker"
)

func mustTracker(t *testing.T, threshold time.Duration) *tracker.Tracker {
	t.Helper()
	tr := tracker.New(threshold)
	if err := tr.Register("job1", "* * * * *"); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return tr
}

func TestRegister_InvalidExpr(t *testing.T) {
	tr := tracker.New(time.Minute)
	if err := tr.Register("bad", "invalid cron"); err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

func TestRecordRun_NoDrift(t *testing.T) {
	tr := mustTracker(t, 30*time.Second)
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if _, err := tr.RecordRun("job1", base); err != nil {
		t.Fatalf("first RecordRun: %v", err)
	}
	// exactly one minute later — no drift
	next := base.Add(time.Minute)
	alert, err := tr.RecordRun("job1", next)
	if err != nil {
		t.Fatalf("second RecordRun: %v", err)
	}
	if alert != nil {
		t.Errorf("unexpected drift alert: %v", alert)
	}
}

func TestRecordRun_DriftDetected(t *testing.T) {
	tr := mustTracker(t, 30*time.Second)
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if _, err := tr.RecordRun("job1", base); err != nil {
		t.Fatal(err)
	}
	// 90 seconds late — exceeds 30 s threshold
	late := base.Add(time.Minute + 90*time.Second)
	alert, err := tr.RecordRun("job1", late)
	if err != nil {
		t.Fatal(err)
	}
	if alert == nil {
		t.Fatal("expected drift alert but got nil")
	}
	if alert.Drift <= 30*time.Second {
		t.Errorf("drift %v should exceed threshold", alert.Drift)
	}
}

func TestRecordRun_UnknownJob(t *testing.T) {
	tr := tracker.New(time.Minute)
	_, err := tr.RecordRun("ghost", time.Now())
	if err == nil {
		t.Fatal("expected error for unknown job")
	}
}

func TestCheckMissed(t *testing.T) {
	tr := mustTracker(t, 10*time.Second)
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if _, err := tr.RecordRun("job1", base); err != nil {
		t.Fatal(err)
	}
	// Simulate 5 minutes passing with no run
	now := base.Add(5 * time.Minute)
	alerts := tr.CheckMissed(now)
	if len(alerts) == 0 {
		t.Fatal("expected at least one missed alert")
	}
	if alerts[0].Job != "job1" {
		t.Errorf("unexpected job in alert: %q", alerts[0].Job)
	}
}

func TestAlertStrings(t *testing.T) {
	now := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	da := tracker.DriftAlert{
		Job:      "myjob",
		Expected: now,
		Actual:   now.Add(2 * time.Minute),
		Drift:    2 * time.Minute,
	}
	if da.String() == "" {
		t.Error("DriftAlert.String() returned empty string")
	}
	ma := tracker.MissedAlert{
		Job:      "myjob",
		Expected: now,
		MissedAt: now.Add(3 * time.Minute),
	}
	if ma.String() == "" {
		t.Error("MissedAlert.String() returned empty string")
	}
}
