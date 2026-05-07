package healthcheck

import (
	"testing"
	"time"
)

func TestNew_InitiallyHealthy(t *testing.T) {
	c := New()
	s := c.Check()
	if !s.Healthy {
		t.Fatal("expected healthy on creation")
	}
	if s.Details != nil {
		t.Fatalf("expected nil details, got %v", s.Details)
	}
}

func TestSetComponent_Degraded(t *testing.T) {
	c := New()
	c.SetComponent("tracker", "no jobs registered")
	s := c.Check()
	if s.Healthy {
		t.Fatal("expected unhealthy after degraded component")
	}
	if s.Details["tracker"] != "no jobs registered" {
		t.Fatalf("unexpected detail: %v", s.Details)
	}
}

func TestSetComponent_ClearRestoresHealth(t *testing.T) {
	c := New()
	c.SetComponent("notifier", "sink unavailable")
	c.SetComponent("notifier", "") // clear
	s := c.Check()
	if !s.Healthy {
		t.Fatal("expected healthy after clearing component")
	}
}

func TestCheck_TimestampsPopulated(t *testing.T) {
	before := time.Now().UTC()
	c := New()
	s := c.Check()
	after := time.Now().UTC()

	if s.StartedAt.Before(before) || s.StartedAt.After(after) {
		t.Fatalf("StartedAt out of range: %v", s.StartedAt)
	}
	if s.CheckedAt.Before(before) || s.CheckedAt.After(after) {
		t.Fatalf("CheckedAt out of range: %v", s.CheckedAt)
	}
}

func TestMultipleComponents_AllMustClearForHealthy(t *testing.T) {
	c := New()
	c.SetComponent("a", "err a")
	c.SetComponent("b", "err b")
	c.SetComponent("a", "") // clear a only
	s := c.Check()
	if s.Healthy {
		t.Fatal("expected unhealthy while b is still degraded")
	}
	c.SetComponent("b", "")
	s = c.Check()
	if !s.Healthy {
		t.Fatal("expected healthy after both components cleared")
	}
}
