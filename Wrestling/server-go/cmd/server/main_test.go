package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"context"
	"wrestling/internal/handlers"
	"wrestling/internal/models"
	"wrestling/internal/ports"
	"wrestling/internal/services"
)

type routerRepo struct{}

func (routerRepo) All(context.Context, ports.ElementQuery) ([]models.Element, error) {
	return []models.Element{{ID: 1, Name: "Pull Up", Category: "стойка", Difficulty: "easy", Image: "/images/emblem.svg"}}, nil
}
func (routerRepo) ByID(context.Context, int) (models.Element, bool, error) {
	return models.Element{ID: 1, Name: "Pull Up", Category: "стойка", Difficulty: "easy", Image: "/images/emblem.svg"}, true, nil
}
func (routerRepo) ByCategory(context.Context, string) ([]models.Element, error) {
	return []models.Element{{ID: 1, Name: "Pull Up"}}, nil
}
func (routerRepo) Stats(context.Context) (models.Stats, error) {
	return models.Stats{Total: 1}, nil
}

type routerContent struct{}

func (routerContent) Elements() ports.ElementRepository          { return routerRepo{} }
func (routerContent) Bars(context.Context) ([]models.Bar, error) { return []models.Bar{}, nil }
func (routerContent) Base(context.Context) ([]models.BaseBlock, error) {
	return []models.BaseBlock{}, nil
}
func (routerContent) Safety(context.Context) ([]models.SafetyItem, error) {
	return []models.SafetyItem{}, nil
}

func testRouter(t *testing.T) http.Handler {
	t.Helper()
	repo := routerContent{}
	h := handlers.New(services.NewContentService(repo), services.NewElementService(repo.Elements()))
	ready := func(w http.ResponseWriter, r *http.Request) { readyWithoutDB(w, r) }
	return newRouter(h, ready, filepath.Join("..", "..", "..", "public"), filepath.Join("..", "..", "..", "public"))
}

func readyWithoutDB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"status":"ready","service":"wrestling-api"}`))
}

func TestRouterSmoke(t *testing.T) {
	router := testRouter(t)
	for _, path := range []string{"/api", "/api/health", "/api/ready", "/api/openapi.yaml", "/api/elements", "/api/elements/1", "/api/elements/category/стойка", "/api/bars", "/api/base", "/api/safety", "/api/stats"} {
		t.Run(path, func(t *testing.T) {
			res := httptest.NewRecorder()
			router.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d body=%q", res.Code, res.Body.String())
			}
		})
	}
}

func TestHealthAndReady(t *testing.T) {
	for name, handler := range map[string]http.HandlerFunc{"health": health, "api": apiInfo} {
		t.Run(name, func(t *testing.T) {
			res := httptest.NewRecorder()
			handler(res, httptest.NewRequest(http.MethodGet, "/api/"+name, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", res.Code)
			}
		})
	}
}

func TestEnvOr(t *testing.T) {
	if got := envOr("WRESTLING_MISSING_TEST_VALUE", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}
