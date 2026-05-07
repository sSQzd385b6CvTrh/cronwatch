package audit

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const defaultRecentN = 100

// Handler returns an http.Handler that serves recent audit entries as JSON.
//
// Query parameters:
//
//	n   – number of entries to return (default 100, max 1000)
//	kind – filter by EventKind (optional)
func Handler(l *Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		n := defaultRecentN
		if s := r.URL.Query().Get("n"); s != "" {
			if v, err := strconv.Atoi(s); err == nil && v > 0 {
				if v > 1000 {
					v = 1000
				}
				n = v
			}
		}

		kindFilter := EventKind(r.URL.Query().Get("kind"))

		entries := l.Recent(n)
		if kindFilter != "" {
			filtered := entries[:0]
			for _, e := range entries {
				if e.Kind == kindFilter {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   len(entries),
			"entries": entries,
		})
	})
}
