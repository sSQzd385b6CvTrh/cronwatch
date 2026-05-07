package healthcheck

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.Handler that serves the health status as JSON.
// It responds with 200 when healthy and 503 when degraded.
func Handler(c *Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := c.Check()

		code := http.StatusOK
		if !status.Healthy {
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)

		if err := json.NewEncoder(w).Encode(status); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})
}
