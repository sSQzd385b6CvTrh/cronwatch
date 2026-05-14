// QuotaLimiter implements a fixed-window rate limit.
//
// # Behaviour
//
// The limiter divides time into non-overlapping windows of a fixed duration
// (e.g. 1 hour). Within each window up to max events are allowed. Once the
// limit is reached every subsequent call to Allow returns false until the next
// window begins.
//
// This differs from SlidingWindow (which rolls from the first event) and
// BurstLimiter (which uses a token-bucket model). Use QuotaLimiter when you
// need hard per-hour / per-day caps that reset on a predictable boundary.
//
// # Example
//
//	ql, err := ratelimit.NewQuotaLimiter(100, time.Hour)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if ql.Allow("job-backup") {
//		// send alert
//	}
//	fmt.Println(ql.Remaining()) // events left this hour
package ratelimit
