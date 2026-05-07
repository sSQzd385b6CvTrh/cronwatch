package metrics

import (
	"testing"
	"time"
)

func TestNewRuntimeCollector_DefaultInterval(t *testing.T) {
	rc := NewRuntimeCollector(0)
	if rc.interval != 30*time.Second {
		t.Fatalf("expected default interval 30s, got %v", rc.interval)
	}
}

func TestRuntimeCollector_SamplePopulatesFields(t *testing.T) {
	rc := NewRuntimeCollector(time.Minute)
	rc.sample()

	snap := rc.Latest()

	if snap.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if snap.GoRoutines <= 0 {
		t.Errorf("expected positive goroutine count, got %d", snap.GoRoutines)
	}
	if snap.HeapAllocMB <= 0 {
		t.Errorf("expected positive HeapAllocMB, got %f", snap.HeapAllocMB)
	}
	if snap.HeapSysMB <= 0 {
		t.Errorf("expected positive HeapSysMB, got %f", snap.HeapSysMB)
	}
}

func TestRuntimeCollector_StartStop(t *testing.T) {
	rc := NewRuntimeCollector(10 * time.Millisecond)
	rc.Start()

	// Allow at least one tick beyond the initial sample.
	time.Sleep(50 * time.Millisecond)
	rc.Stop()

	snap := rc.Latest()
	if snap.Timestamp.IsZero() {
		t.Error("expected a collected snapshot after Start")
	}
}

func TestRuntimeCollector_LatestUpdatesOverTime(t *testing.T) {
	rc := NewRuntimeCollector(10 * time.Millisecond)
	rc.sample()
	first := rc.Latest()

	time.Sleep(15 * time.Millisecond)
	rc.sample()
	second := rc.Latest()

	if !second.Timestamp.After(first.Timestamp) {
		t.Errorf("expected second snapshot to be newer: first=%v second=%v",
			first.Timestamp, second.Timestamp)
	}
}

func TestRuntimeCollector_GCCyclesNonNegative(t *testing.T) {
	rc := NewRuntimeCollector(time.Minute)
	rc.sample()
	snap := rc.Latest()
	// GC may not have run yet in a short test, but the field must be valid.
	if snap.GCCycles < 0 {
		t.Errorf("unexpected negative GC cycle count: %d", snap.GCCycles)
	}
}
