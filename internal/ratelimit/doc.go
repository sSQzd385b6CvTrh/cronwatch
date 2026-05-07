// Package ratelimit provides a per-job cooldown gate that prevents
// alert storms when a cron job repeatedly drifts or misses its schedule.
//
// Usage:
//
//	rl := ratelimit.New(5 * time.Minute)
//	if rl.Allow("backup-job") {
//	    notifier.Send(alert)
//	}
//
// The limiter is safe for concurrent use. Each job name is tracked
// independently; suppressing alerts for one job does not affect others.
package ratelimit
