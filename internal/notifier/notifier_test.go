package notifier_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/notifier"
)

// captureSink records every alert sent to it.
type captureSink struct {
	alerts []notifier.Alert
	forceErr bool
}

func (c *captureSink) Send(a notifier.Alert) error {
	if c.forceErr {
		return errors.New("sink failure")
	}
	c.alerts = append(c.alerts, a)
	return nil
}

func TestNotify_DeliveredToAllSinks(t *testing.T) {
	s1, s2 := &captureSink{}, &captureSink{}
	n := notifier.New(s1, s2)

	a := notifier.Alert{JobName: "backup", Level: notifier.LevelWarn, Message: "drift detected"}
	if err := n.Notify(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s1.alerts) != 1 || len(s2.alerts) != 1 {
		t.Fatalf("expected 1 alert per sink, got s1=%d s2=%d", len(s1.alerts), len(s2.alerts))
	}
	if s1.alerts[0].JobName != "backup" {
		t.Errorf("unexpected job name: %s", s1.alerts[0].JobName)
	}
}

func TestNotify_TimestampBackfilled(t *testing.T) {
	s := &captureSink{}
	n := notifier.New(s)

	before := time.Now()
	_ = n.Notify(notifier.Alert{JobName: "job", Level: notifier.LevelInfo, Message: "ok"})
	after := time.Now()

	got := s.alerts[0].OccuredAt
	if got.Before(before) || got.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", got, before, after)
	}
}

func TestNotify_SinkErrorReturned(t *testing.T) {
	s := &captureSink{forceErr: true}
	n := notifier.New(s)

	err := n.Notify(notifier.Alert{JobName: "job", Level: notifier.LevelError, Message: "missed"})
	if err == nil {
		t.Fatal("expected error from failing sink, got nil")
	}
}

func TestLogSink_Output(t *testing.T) {
	var buf bytes.Buffer
	sink := notifier.NewLogSink(&buf)

	a := notifier.Alert{
		JobName:   "cleanup",
		Level:     notifier.LevelError,
		Message:   "missed run",
		OccuredAt: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}
	if err := sink.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"cleanup", "ERROR", "missed run", "2024-06-01"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing expected substring %q", out, want)
		}
	}
}
