// BurstLimiter implements a simple burst-then-cooldown rate limiting strategy.
//
// Unlike the sliding window limiter (which tracks event counts over a rolling
// time window), BurstLimiter is designed for alert suppression scenarios where
// a short flurry of notifications is acceptable but sustained noise should be
// silenced until a quiet period has passed.
//
// Behaviour summary:
//
//	1. Up to Burst consecutive events are allowed immediately.
//	2. Once the burst is exhausted all subsequent events are blocked.
//	3. After CooldownDuration has elapsed since the last allowed event the
//	   burst counter is fully replenished and events are permitted again.
//
// Thread safety: all methods are safe for concurrent use.
package ratelimit
