package audit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestLogger(t *testing.T) *Logger {
	t.Helper()
	l, _ := tempLogger(t, 200)
	return l
}

func TestHandler_GET_ReturnsJSON(t *testing.T) {
	l := newTestLogger(t)
	l.Log(KindRun, "backup", "")
	l.Log(KindDrift, "backup", "3s")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	Handler(l).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["count"].(float64) != 2 {
		t.Errorf("count = %v, want 2", body["count"])
	}
}

func TestHandler_NonGET_Returns405(t *testing.T) {
	l := newTestLogger(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/audit", nil)
	Handler(l).ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHandler_KindFilter(t *testing.T) {
	l := newTestLogger(t)
	l.Log(KindRun, "j", "")
	l.Log(KindMissed, "j", "")
	l.Log(KindRun, "j", "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit?kind=run", nil)
	Handler(l).ServeHTTP(rec, req)

	var body map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["count"].(float64) != 2 {
		t.Errorf("count = %v, want 2", body["count"])
	}
}

func TestHandler_NParam(t *testing.T) {
	l := newTestLogger(t)
	for i := 0; i < 20; i++ {
		l.Log(KindRun, "j", "")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit?n=5", nil)
	Handler(l).ServeHTTP(rec, req)

	var body map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["count"].(float64) != 5 {
		t.Errorf("count = %v, want 5", body["count"])
	}
}

func TestHandler_ContentType(t *testing.T) {
	l := newTestLogger(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	Handler(l).ServeHTTP(rec, req)
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
