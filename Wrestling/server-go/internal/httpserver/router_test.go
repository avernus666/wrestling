package httpserver

import (
  "context"
  "net/http"
  "net/http/httptest"
  "testing"
)

type testPing struct{ err error }
func (p testPing) Ping(context.Context) error { return p.err }

func TestHealthAndMethodGuards(t *testing.T) {
  h := NewRouter(RouterConfig{Pool: testPing{}, PublicDir: "../../public", WebDir: "../../public"})
  rr := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
  h.ServeHTTP(rr, req)
  if rr.Code != http.StatusOK { t.Fatalf("health status=%d", rr.Code) }

  rr = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPost, "/api/health", nil)
  h.ServeHTTP(rr, req)
  if rr.Code != http.StatusMethodNotAllowed { t.Fatalf("health POST status=%d", rr.Code) }
}

func TestReadyUsesDependency(t *testing.T) {
  h := NewRouter(RouterConfig{Pool: testPing{}, PublicDir: "../../public", WebDir: "../../public"})
  rr := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/api/ready", nil)
  h.ServeHTTP(rr, req)
  if rr.Code != http.StatusOK { t.Fatalf("ready status=%d", rr.Code) }
}
