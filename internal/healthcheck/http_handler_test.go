package healthcheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_GET_Healthy(t *testing.T) {
	c := New()
	h := Handler(c)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var s Status
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !s.Healthy {
		t.Fatal("expected healthy status in body")
	}
}

func TestHandler_GET_Degraded(t *testing.T) {
	c := New()
	c.SetComponent("watcher", "poll loop stalled")
	h := Handler(c)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var s Status
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if s.Healthy {
		t.Fatal("expected unhealthy status in body")
	}
	if s.Details["watcher"] != "poll loop stalled" {
		t.Fatalf("unexpected detail: %v", s.Details)
	}
}

func TestHandler_NonGET_Returns405(t *testing.T) {
	c := New()
	h := Handler(c)

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/healthz", nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s: expected 405, got %d", method, rec.Code)
		}
	}
}

func TestHandler_ContentType(t *testing.T) {
	c := New()
	h := Handler(c)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}
}
