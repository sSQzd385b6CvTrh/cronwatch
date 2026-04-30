package schedule

import (
	"testing"
	"time"
)

func TestParse_Valid(t *testing.T) {
	tests := []struct {
		expr string
	}{
		{"* * * * *"},
		{"0 * * * *"},
		{"30 6 * * 1"},
		{"0,30 8-17 * * 1-5"},
		{"15 14 1 * *"},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			_, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.expr, err)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	tests := []string{
		"",
		"* * * *",
		"60 * * * *",
		"* 25 * * *",
		"* * 32 * *",
		"* * * 13 *",
		"* * * * 7",
		"abc * * * *",
	}
	for _, expr := range tests {
		t.Run(expr, func(t *testing.T) {
			_, err := Parse(expr)
			if err == nil {
				t.Fatalf("expected error for %q but got nil", expr)
			}
		})
	}
}

func TestNextAfter_EveryMinute(t *testing.T) {
	c, err := Parse("* * * * *")
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	next := c.NextAfter(base)
	expected := time.Date(2024, 1, 15, 10, 1, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("NextAfter got %v, want %v", next, expected)
	}
}

func TestNextAfter_HourlyAtZero(t *testing.T) {
	c, err := Parse("0 * * * *")
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	next := c.NextAfter(base)
	expected := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("NextAfter got %v, want %v", next, expected)
	}
}

func TestNextAfter_WeekdayConstraint(t *testing.T) {
	// Monday only
	c, err := Parse("0 9 * * 1")
	if err != nil {
		t.Fatal(err)
	}
	// 2024-01-15 is a Monday
	base := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	next := c.NextAfter(base)
	// next Monday
	expected := time.Date(2024, 1, 22, 9, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("NextAfter got %v, want %v", next, expected)
	}
}
