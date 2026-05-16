// Package ratelimit — token_bucket.go
//
// TokenBucket implements the classic token-bucket algorithm for rate
// limiting cron-job alert notifications.
//
// # How it works
//
// A bucket starts full (capacity tokens). Each successful Allow call
// consumes one token. Tokens refill continuously at a fixed rate
// (tokens-per-second). Once the bucket is empty, Allow returns false
// until enough time has elapsed for at least one token to accumulate.
//
// # Comparison with other limiters
//
//   - Unlike the sliding-window limiter, bursts up to the full
//     capacity are always permitted as long as tokens are available.
//   - Unlike the burst limiter, the refill is continuous rather than
//     resetting after a fixed cooldown window.
//
// # Example
//
//	tb, err := ratelimit.NewTokenBucket(10, 1) // cap=10, 1 token/sec
//	if err != nil { … }
//	if tb.Allow("backup-job") {
//	    // send alert
//	}
package ratelimit
