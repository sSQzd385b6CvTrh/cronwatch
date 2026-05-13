// Package ratelimit provides several rate-limiting and circuit-breaking
// primitives used by cronwatch to suppress alert storms and protect downstream
// notification sinks.
//
// CircuitBreaker
//
// A CircuitBreaker protects an outbound call (e.g. a webhook or email sink)
// from being hammered when the downstream service is unhealthy.
//
// States:
//
//   - Closed  — normal operation; all calls proceed.
//   - Open    — downstream is considered unhealthy; calls are blocked until
//               the cooldown period elapses.
//   - HalfOpen — one probe call is allowed through to test recovery; further
//               calls are blocked until the probe outcome is recorded.
//
// Typical usage:
//
//	cb, _ := ratelimit.NewCircuitBreaker(5, 30*time.Second)
//
//	if !cb.Allow() {
//	    return errors.New("circuit open")
//	}
//	if err := sink.Send(ctx, alert); err != nil {
//	    cb.RecordFailure()
//	    return err
//	}
//	cb.RecordSuccess()
package ratelimit
