package notifier

import (
	"errors"
	"testing"
	"time"
)

// fakeSink records every alert it receives and can be configured to error.
type fakeSink struct {
	alerts []Alert
	errOn  int // return error on the nth call (1-based); 0 = never
	calls  int
}

func (f *fakeSink) Send(a Alert) error {
	f.calls++
	if f.errOn > 0 && f.calls == f.errOn {
		return errors.New("fake sink error")
	}
	f.alerts = append(f.alerts, a)
	return nil
}

func TestNotify_DeliveredToAllSinks_Multi(t *testing.T) {
	s1 := &fakeSink{}
	s2 := &fakeSink{}
	n := New(s1, s2)

	a := Alert{JobName: "job-a", Kind: KindMissed, Scheduled: time.Now(), DetectedAt: time.Now()}
	if err := n.Notify(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s1.alerts) != 1 || len(s2.alerts) != 1 {
		t.Errorf("expected each sink to receive 1 alert")
	}
}

func TestNotify_ContinuesAfterSinkError(t *testing.T) {
	s1 := &fakeSink{errOn: 1}
	s2 := &fakeSink{}
	n := New(s1, s2)

	a := Alert{JobName: "job-b", Kind: KindDrift, Scheduled: time.Now(), DetectedAt: time.Now()}
	err := n.Notify(a)
	if err == nil {
		t.Fatal("expected aggregated error")
	}
	// s2 must still have received the alert despite s1 failing
	if len(s2.alerts) != 1 {
		t.Errorf("s2 should have received alert even after s1 error")
	}
}
