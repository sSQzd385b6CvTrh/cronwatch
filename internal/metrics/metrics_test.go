package metrics

import (
	"testing"
	"time"
)

func TestNew_ZeroValues(t *testing.T) {
	c := New()
	s := c.Snapshot()
	if s.JobsTracked != 0 || s.RunsRecorded != 0 || s.AlertsTotal != 0 || s.MissedTotal != 0 {
		t.Fatalf("expected zero snapshot, got %+v", s)
	}
	if len(s.DriftSamples) != 0 {
		t.Fatalf("expected no drift samples")
	}
}

func TestIncCounters(t *testing.T) {
	c := New()
	c.SetJobsTracked(3)
	c.IncRuns()
	c.IncRuns()
	c.IncAlerts()
	c.IncMissed()
	c.IncMissed()

	s := c.Snapshot()
	if s.JobsTracked != 3 {
		t.Errorf("jobs tracked: want 3, got %d", s.JobsTracked)
	}
	if s.RunsRecorded != 2 {
		t.Errorf("runs: want 2, got %d", s.RunsRecorded)
	}
	if s.AlertsTotal != 1 {
		t.Errorf("alerts: want 1, got %d", s.AlertsTotal)
	}
	if s.MissedTotal != 2 {
		t.Errorf("missed: want 2, got %d", s.MissedTotal)
	}
}

func TestRecordDrift_AppendedAndCopied(t *testing.T) {
	c := New()
	c.RecordDrift("backup", 5*time.Second)
	c.RecordDrift("cleanup", 10*time.Second)

	s := c.Snapshot()
	if len(s.DriftSamples) != 2 {
		t.Fatalf("want 2 samples, got %d", len(s.DriftSamples))
	}
	if s.DriftSamples[0].Job != "backup" {
		t.Errorf("unexpected job name: %s", s.DriftSamples[0].Job)
	}
	if s.DriftSamples[1].Drift != 10*time.Second {
		t.Errorf("unexpected drift: %v", s.DriftSamples[1].Drift)
	}
}

func TestRecordDrift_CappedAt500(t *testing.T) {
	c := New()
	for i := 0; i < 600; i++ {
		c.RecordDrift("job", time.Duration(i)*time.Millisecond)
	}
	s := c.Snapshot()
	if len(s.DriftSamples) != 500 {
		t.Errorf("want 500 samples, got %d", len(s.DriftSamples))
	}
	// oldest should have been evicted; first remaining drift = 100ms
	if s.DriftSamples[0].Drift != 100*time.Millisecond {
		t.Errorf("expected eviction of oldest; first drift = %v", s.DriftSamples[0].Drift)
	}
}

func TestSnapshot_IsolatedCopy(t *testing.T) {
	c := New()
	c.RecordDrift("job", time.Second)
	s1 := c.Snapshot()
	c.RecordDrift("job", 2*time.Second)
	s2 := c.Snapshot()

	if len(s1.DriftSamples) != 1 {
		t.Errorf("s1 should have 1 sample, got %d", len(s1.DriftSamples))
	}
	if len(s2.DriftSamples) != 2 {
		t.Errorf("s2 should have 2 samples, got %d", len(s2.DriftSamples))
	}
}
