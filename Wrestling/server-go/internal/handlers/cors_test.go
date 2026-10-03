package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnableCORSWithOriginsAllowsConfiguredOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	EnableCORSWithOrigins("https://example.com, https://admin.example.com", next).ServeHTTP(res, req)
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("expected configured origin, got %q", got)
	}
}

func TestEnableCORSWithOriginsRejectsUnknownOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example")
	EnableCORSWithOrigins("https://example.com", next).ServeHTTP(res, req)
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected CORS origin %q", got)
	}
}
