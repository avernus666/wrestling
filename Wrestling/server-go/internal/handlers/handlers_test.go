package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"wrestling/internal/models"
	"wrestling/internal/ports"
	"wrestling/internal/services"
)

type fakeContentRepository struct{ elements ports.ElementRepository }

func (f fakeContentRepository) Elements() ports.ElementRepository { return f.elements }
func (f fakeContentRepository) Bars(context.Context) ([]models.Bar, error) {
	return []models.Bar{{ID: 1, Name: "Турник"}}, nil
}
func (f fakeContentRepository) Base(context.Context) ([]models.BaseBlock, error) {
	return []models.BaseBlock{{Title: "База"}}, nil
}
func (f fakeContentRepository) Safety(context.Context) ([]models.SafetyItem, error) {
	return []models.SafetyItem{{ID: 1, Name: "Разминка"}}, nil
}

type fakeElementRepository struct{}

func (fakeElementRepository) All(context.Context, ports.ElementQuery) ([]models.Element, error) {
	return []models.Element{{ID: 1, Name: "Pull Up", Category: "стойка", Difficulty: "easy"}, {ID: 2, Name: "Muscle Up", Category: "партер", Difficulty: "hard"}}, nil
}
func (fakeElementRepository) ByID(_ context.Context, id int) (models.Element, bool, error) {
	for _, item := range []models.Element{{ID: 1, Name: "Pull Up", Category: "стойка", Difficulty: "easy"}, {ID: 2, Name: "Muscle Up", Category: "партер", Difficulty: "hard"}} {
		if item.ID == id {
			return item, true, nil
		}
	}
	return models.Element{}, false, nil
}
func (fakeElementRepository) ByCategory(context.Context, string) ([]models.Element, error) {
	return []models.Element{{ID: 1, Name: "Pull Up", Category: "стойка", Difficulty: "easy"}}, nil
}
func (fakeElementRepository) Stats(context.Context) (models.Stats, error) {
	return models.Stats{Total: 2, ByCategory: map[string]int{"стойка": 1, "партер": 1}, ByDifficulty: map[string]int{"easy": 1, "hard": 1}, Categories: []string{"партер", "стойка"}, Difficulties: []string{"easy", "hard"}}, nil
}

func newTestHandler() *Handler {
	repo := fakeContentRepository{elements: fakeElementRepository{}}
	return New(services.NewContentService(repo), services.NewElementService(repo.Elements()))
}

func TestGetElements(t *testing.T) {
	res := httptest.NewRecorder()
	newTestHandler().GetElements(res, httptest.NewRequest(http.MethodGet, "/api/elements", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var items []models.Element
	if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestGetElementValidationAndNotFound(t *testing.T) {
	for _, tt := range []struct {
		path string
		want int
	}{
		{"/api/elements/nope", http.StatusBadRequest},
		{"/api/elements/999", http.StatusNotFound},
		{"/api/elements/1", http.StatusOK},
	} {
		t.Run(tt.path, func(t *testing.T) {
			res := httptest.NewRecorder()
			newTestHandler().GetElement(res, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if res.Code != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, res.Code)
			}
		})
	}
}

func TestContentHandlers(t *testing.T) {
	for _, tt := range []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"bars", newTestHandler().GetBars},
		{"base", newTestHandler().GetBase},
		{"safety", newTestHandler().GetSafety},
		{"stats", newTestHandler().GetStats},
	} {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			tt.handler(res, httptest.NewRequest(http.MethodGet, "/api/"+tt.name, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", res.Code)
			}
		})
	}
}

func TestEnableCORS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	EnableCORS(next).ServeHTTP(res, req)
	if res.Header.Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatal("missing CORS header")
	}
	if res.Code != http.StatusTeapot {
		t.Fatalf("expected downstream response, got %d", res.Code)
	}
}
