package audit

import (
	"testing"
	"time"
)

func makeEntry(msg string) Entry {
	return Entry{Kind: "test", Message: msg, At: time.Now()}
}

func TestRingBuffer_EmptyRecent(t *testing.T) {
	rb := newRingBuffer(5)
	if got := rb.recent(3); got != nil {
		t.Fatalf("expected nil for empty buffer, got %v", got)
	}
}

func TestRingBuffer_PushAndRecent(t *testing.T) {
	rb := newRingBuffer(5)
	for _, msg := range []string{"a", "b", "c"} {
		rb.push(makeEntry(msg))
	}
	got := rb.recent(3)
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if got[0].Message != "a" || got[1].Message != "b" || got[2].Message != "c" {
		t.Fatalf("unexpected order: %v", got)
	}
}

func TestRingBuffer_Wraps(t *testing.T) {
	rb := newRingBuffer(3)
	for _, msg := range []string{"a", "b", "c", "d", "e"} {
		rb.push(makeEntry(msg))
	}
	// Buffer holds only the last 3: c, d, e
	got := rb.recent(3)
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if got[0].Message != "c" || got[1].Message != "d" || got[2].Message != "e" {
		t.Fatalf("unexpected entries after wrap: %v", got)
	}
}

func TestRingBuffer_RecentFewerThanStored(t *testing.T) {
	rb := newRingBuffer(10)
	for _, msg := range []string{"a", "b", "c", "d", "e"} {
		rb.push(makeEntry(msg))
	}
	got := rb.recent(2)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	// Should be the 2 most recent: d, e
	if got[0].Message != "d" || got[1].Message != "e" {
		t.Fatalf("expected last 2 entries, got %v", got)
	}
}

func TestRingBuffer_Len(t *testing.T) {
	rb := newRingBuffer(4)
	if rb.len() != 0 {
		t.Fatalf("expected len 0, got %d", rb.len())
	}
	rb.push(makeEntry("x"))
	rb.push(makeEntry("y"))
	if rb.len() != 2 {
		t.Fatalf("expected len 2, got %d", rb.len())
	}
	// Fill and overflow
	rb.push(makeEntry("z"))
	rb.push(makeEntry("w"))
	rb.push(makeEntry("v")) // overflows
	if rb.len() != 4 {
		t.Fatalf("expected len 4 (cap), got %d", rb.len())
	}
}

func TestRingBuffer_ZeroCapacityClamped(t *testing.T) {
	rb := newRingBuffer(0)
	if rb.cap != 1 {
		t.Fatalf("expected capacity clamped to 1, got %d", rb.cap)
	}
	rb.push(makeEntry("only"))
	got := rb.recent(1)
	if len(got) != 1 || got[0].Message != "only" {
		t.Fatalf("unexpected result: %v", got)
	}
}
