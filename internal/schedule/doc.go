// Package schedule provides parsing and scheduling utilities for standard
// 5-field cron expressions used by cronwatch to determine expected execution
// times and detect drift or missed runs.
//
// Supported cron field syntax:
//
//	*          – every value in the allowed range
//	v          – a single integer value
//	v1,v2,...  – a comma-separated list of values
//	v1-v2      – an inclusive range of values
//
// Field order and allowed ranges:
//
//	Minute   0-59
//	Hour     0-23
//	Day      1-31
//	Month    1-12
//	Weekday  0-6  (0 = Sunday)
//
// Example usage:
//
//	expr, err := schedule.Parse("30 6 * * 1-5")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	next := expr.NextAfter(time.Now())
//	fmt.Println("Next run:", next)
package schedule
