package handlers

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	res := httptest.NewRecorder()
	SecurityHeaders(next).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	for key, want := range map[string]string{"X-Content-Type-Options": "nosniff", "X-Frame-Options": "SAMEORIGIN", "Referrer-Policy": "strict-origin-when-cross-origin"} {
		if got := res.Header().Get(key); got != want {
			t.Fatalf("%s: expected %q, got %q", key, want, got)
		}
	}
}

func TestRequestIDPreservesIncomingID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "client-id")
	RequestID(next).ServeHTTP(res, req)
	if res.Header().Get("X-Request-ID") != "client-id" {
		t.Fatal("request ID was not preserved")
	}
}

func TestRequestIDGeneratesID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	res := httptest.NewRecorder()
	RequestID(next).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := res.Header().Get("X-Request-ID"); len(got) != 32 {
		t.Fatalf("expected generated request ID, got %q", got)
	}
}

func TestRequestLoggerRecordsStatusAndPath(t *testing.T) {
	var buffer bytes.Buffer
	logger := log.New(&buffer, "", 0)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusCreated) })
	res := httptest.NewRecorder()
	RequestLogger(logger, next).ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/test", nil))
	logLine := buffer.String()
	if !strings.Contains(logLine, "method=POST") || !strings.Contains(logLine, "path=/api/test") || !strings.Contains(logLine, "status=201") {
		t.Fatalf("unexpected log: %s", logLine)
	}
}
