package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempLogger(t *testing.T, max int) (*Logger, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	l, err := New(path, max)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return l, path
}

func TestLog_WritesToFile(t *testing.T) {
	l, path := tempLogger(t, 10)
	l.Log(KindRun, "backup", "")
	_ = l.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var e Entry
	if err := json.Unmarshal(data[:len(data)-1], &e); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if e.Kind != KindRun {
		t.Errorf("kind = %q, want %q", e.Kind, KindRun)
	}
	if e.Job != "backup" {
		t.Errorf("job = %q, want \"backup\"", e.Job)
	}
}

func TestRecent_Order(t *testing.T) {
	l, _ := tempLogger(t, 100)
	for _, kind := range []EventKind{KindRegistered, KindRun, KindDrift, KindMissed} {
		l.Log(kind, "job", "")
	}
	entries := l.Recent(4)
	if len(entries) != 4 {
		t.Fatalf("len = %d, want 4", len(entries))
	}
	expected := []EventKind{KindRegistered, KindRun, KindDrift, KindMissed}
	for i, e := range entries {
		if e.Kind != expected[i] {
			t.Errorf("entries[%d].Kind = %q, want %q", i, e.Kind, expected[i])
		}
	}
}

func TestRecent_FewerThanRequested(t *testing.T) {
	l, _ := tempLogger(t, 100)
	l.Log(KindRun, "job", "")
	if got := l.Recent(50); len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestRingBuffer_Wraps(t *testing.T) {
	const max = 5
	l, _ := tempLogger(t, max)
	for i := 0; i < 8; i++ {
		l.Log(KindRun, "j", "")
	}
	if got := l.Recent(max); len(got) != max {
		t.Errorf("len = %d, want %d", len(got), max)
	}
}

func TestLog_TimestampIsUTC(t *testing.T) {
	l, _ := tempLogger(t, 10)
	before := time.Now().UTC()
	l.Log(KindDrift, "job", "5s drift")
	after := time.Now().UTC()

	entries := l.Recent(1)
	ts := entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in [%v, %v]", ts, before, after)
	}
	if ts.Location() != time.UTC {
		t.Errorf("timestamp not UTC")
	}
}

func TestNew_DefaultMax(t *testing.T) {
	l, _ := tempLogger(t, 0)
	if l.maxEntries != 1000 {
		t.Errorf("maxEntries = %d, want 1000", l.maxEntries)
	}
}
