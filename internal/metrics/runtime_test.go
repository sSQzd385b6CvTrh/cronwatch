package metrics

import (
	"testing"
	"time"
)

func TestNewRuntimeCollector_DefaultInterval(t *testing.T) {
	rc := NewRuntimeCollector(0)
	if rc.interval != defaultRuntimeInterval {
		t.Fatalf("expected default interval %v, got %v", defaultRuntimeInterval, rc.interval)
	}
}

func TestRuntimeCollector_SamplePopulatesFields(t *testing.T) {
	rc := NewRuntimeCollector(time.Hour) // long interval — won't tick during test
	rc.sample()
	s := rc.Latest()

	if s.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if s.Goroutines <= 0 {
		t.Errorf("expected goroutines > 0, got %d", s.Goroutines)
	}
	if s.HeapAllocMB <= 0 {
		t.Errorf("expected HeapAllocMB > 0, got %f", s.HeapAllocMB)
	}
	if s.HeapSysMB <= 0 {
		t.Errorf("expected HeapSysMB > 0, got %f", s.HeapSysMB)
	}
}

func TestRuntimeCollector_StartStop(t *testing.T) {
	rc := NewRuntimeCollector(10 * time.Millisecond)
	rc.Start()
	time.Sleep(35 * time.Millisecond)
	rc.Stop()

	s := rc.Latest()
	if s.Timestamp.IsZero() {
		t.Error("expected sample after Start")
	}
}

func TestRuntimeCollector_LatestUpdatesOverTime(t *testing.T) {
	rc := NewRuntimeCollector(10 * time.Millisecond)
	rc.Start()
	defer rc.Stop()

	time.Sleep(5 * time.Millisecond)
	first := rc.Latest()

	time.Sleep(25 * time.Millisecond)
	second := rc.Latest()

	if !second.Timestamp.After(first.Timestamp) {
		t.Errorf("expected second sample to be newer: first=%v second=%v",
			first.Timestamp, second.Timestamp)
	}
}

func TestRuntimeCollector_GCCyclesNonNegative(t *testing.T) {
	rc := NewRuntimeCollector(time.Hour)
	rc.sample()
	s := rc.Latest()
	if s.GCCycles < 0 {
		t.Errorf("GCCycles should be non-negative, got %d", s.GCCycles)
	}
}
