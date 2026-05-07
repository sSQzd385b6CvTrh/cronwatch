// Package healthcheck provides liveness and readiness probing for the
// cronwatch daemon.
//
// Components across the application can register degraded states via
// SetComponent; the HTTP handler exposes the aggregated status at a
// configurable endpoint (typically /healthz).
//
// A 200 response indicates all components are healthy.
// A 503 response indicates at least one component has reported a problem.
package healthcheck
