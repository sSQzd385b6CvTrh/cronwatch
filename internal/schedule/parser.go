package schedule

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronExpr represents a parsed cron expression with five fields.
type CronExpr struct {
	Raw     string
	Minute  []int
	Hour    []int
	Day     []int
	Month   []int
	Weekday []int
}

// Parse parses a standard 5-field cron expression string.
func Parse(expr string) (*CronExpr, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid cron expression %q: expected 5 fields, got %d", expr, len(fields))
	}

	ranges := []struct{ min, max int }{
		{0, 59}, // minute
		{0, 23}, // hour
		{1, 31}, // day
		{1, 12}, // month
		{0, 6},  // weekday
	}

	c := &CronExpr{Raw: expr}
	ptrs := []*[]int{&c.Minute, &c.Hour, &c.Day, &c.Month, &c.Weekday}

	for i, field := range fields {
		vals, err := parseField(field, ranges[i].min, ranges[i].max)
		if err != nil {
			return nil, fmt.Errorf("field %d: %w", i+1, err)
		}
		*ptrs[i] = vals
	}

	return c, nil
}

// NextAfter returns the next scheduled time after t.
func (c *CronExpr) NextAfter(t time.Time) time.Time {
	t = t.Add(time.Minute).Truncate(time.Minute)
	for i := 0; i < 366*24*60; i++ {
		if contains(c.Month, int(t.Month())) &&
			contains(c.Day, t.Day()) &&
			contains(c.Weekday, int(t.Weekday())) &&
			contains(c.Hour, t.Hour()) &&
			contains(c.Minute, t.Minute()) {
			return t
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}
}

func parseField(field string, min, max int) ([]int, error) {
	if field == "*" {
		vals := make([]int, max-min+1)
		for i := range vals {
			vals[i] = min + i
		}
		return vals, nil
	}

	var result []int
	for _, part := range strings.Split(field, ",") {
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			lo, err1 := strconv.Atoi(bounds[0])
			hi, err2 := strconv.Atoi(bounds[1])
			if err1 != nil || err2 != nil || lo > hi || lo < min || hi > max {
				return nil, fmt.Errorf("invalid range %q", part)
			}
			for v := lo; v <= hi; v++ {
				result = append(result, v)
			}
		} else {
			v, err := strconv.Atoi(part)
			if err != nil || v < min || v > max {
				return nil, fmt.Errorf("invalid value %q", part)
			}
			result = append(result, v)
		}
	}
	return result, nil
}

func contains(vals []int, v int) bool {
	for _, x := range vals {
		if x == v {
			return true
		}
	}
	return false
}
