package httpserver

import (
  "net/http"
  "net/http/httptest"
  "testing"
  "time"
)

func TestLimiterBlocksAfterLimit(t *testing.T) {
  l := newLimiter(2, time.Minute)
  if ok, _ := l.allow("test"); !ok { t.Fatal("first request blocked") }
  if ok, _ := l.allow("test"); !ok { t.Fatal("second request blocked") }
  if ok, _ := l.allow("test"); ok { t.Fatal("third request should be blocked") }
}

func TestBodyLimitRejectsOversizedPayload(t *testing.T) {
  h := BodyLimit(4, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    b := make([]byte, 8)
    _, err := r.Body.Read(b)
    if err == nil { t.Fatal("expected limited body error") }
    w.WriteHeader(http.StatusNoContent)
  }))
  req := httptest.NewRequest(http.MethodPost, "/", nil)
  req.Body = http.NoBody
  rr := httptest.NewRecorder()
  h.ServeHTTP(rr, req)
}
