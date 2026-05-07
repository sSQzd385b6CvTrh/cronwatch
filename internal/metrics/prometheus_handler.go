package metrics

import (
	"net/http"
)

// PrometheusHandler returns an http.Handler that serves Prometheus text-format
// metrics for the given Store. The endpoint responds only to GET requests and
// uses the standard content-type expected by Prometheus scrapers.
func PrometheusHandler(store *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		snap := store.Snapshot()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if err := WritePrometheus(w, snap); err != nil {
			// Headers already sent; best-effort log via stderr is handled by caller.
			_ = err
		}
	})
}
