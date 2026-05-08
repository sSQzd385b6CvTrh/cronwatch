// Package ratelimit provides rate-limiting primitives used by cronwatch to
// suppress alert storms when many jobs drift or are missed simultaneously.
//
// # Limiter (token-bucket / cooldown)
//
// The Limiter type enforces a per-key cooldown: once an alert fires for a
// given key it is silenced until the cooldown duration elapses.
//
// # PerJobLimiter
//
// PerJobLimiter wraps Limiter with per-job-name keying so callers do not need
// to manage key strings themselves.
//
// # SlidingWindow
//
// SlidingWindow tracks event timestamps within a rolling time window and
// rejects new events once the maximum count within that window is reached.
// Unlike the cooldown-based Limiter, SlidingWindow is useful when you want to
// cap the total number of alerts fired across all jobs within a short burst
// window (e.g. no more than 10 alerts per minute globally).
//
// Example:
//
//	global := ratelimit.NewSlidingWindow(10, time.Minute)
//	if global.Allow() {
//		notifier.Notify(ctx, alert)
//	}
package ratelimit
