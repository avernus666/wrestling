package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"strconv"
	"strings"
	"wrestling/internal/models"

	"wrestling/internal/services"
)

type ContentService interface {
	Community(context.Context, string, string, int, int) ([]models.CommunityElement, error)
	CreateCommunity(context.Context, models.CommunityElement) error
	Comments(context.Context, string, int) ([]models.CommunityComment, error)
	CreateComment(context.Context, models.CommunityComment) error
	Report(context.Context, models.CommunityReport, string) error
	Bars(context.Context) ([]models.Bar, error)
	Base(context.Context) ([]models.BaseBlock, error)
	Safety(context.Context) ([]models.SafetyItem, error)
}

type ElementService interface {
	List(context.Context, services.ElementFilter) ([]models.Element, error)
	Get(context.Context, int) (models.Element, bool, error)
	ByCategory(context.Context, string) ([]models.Element, error)
	Stats(context.Context) (models.Stats, error)
}

type Handler struct {
	content  ContentService
	elements ElementService
	auth     AuthService
	users    UserService
	youtube  interface {
		Verify(context.Context, string) (string, error)
	}
}

func New(content ContentService, elements ElementService, deps ...any) *Handler {
	h := &Handler{content: content, elements: elements}
	for _, dep := range deps {
		switch value := dep.(type) {
		case AuthService:
			h.auth = value
		case UserService:
			h.users = value
		case interface {
			Verify(context.Context, string) (string, error)
		}:
			h.youtube = value
		}
	}
	return h
}

func (h *Handler) GetElements(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	items, err := h.elements.List(r.Context(), services.ElementFilter{
		Category: r.URL.Query().Get("category"), Subcategory: r.URL.Query().Get("subcategory"),
		Difficulty: r.URL.Query().Get("difficulty"), Search: r.URL.Query().Get("search"),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Не удалось загрузить каталог")
		return
	}
	respondJSON(w, items)
}

func (h *Handler) GetElement(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/elements/"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Некорректный ID элемента")
		return
	}
	item, ok, err := h.elements.Get(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Не удалось загрузить элемент")
		return
	}
	if !ok {
		respondError(w, http.StatusNotFound, "Элемент не найден")
		return
	}
	respondJSON(w, item)
}

func (h *Handler) GetElementsByCategory(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	category := strings.TrimPrefix(r.URL.Path, "/api/elements/category/")
	items, err := h.elements.ByCategory(r.Context(), category)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Не удалось загрузить категорию")
		return
	}
	respondJSON(w, items)
}

func (h *Handler) GetBars(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	items, err := h.content.Bars(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Не удалось загрузить оборудование")
		return
	}
	respondJSON(w, items)
}

func (h *Handler) GetBase(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	items, err := h.content.Base(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Не удалось загрузить базу")
		return
	}
	respondJSON(w, items)
}

func (h *Handler) GetSafety(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	items, err := h.content.Safety(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Не удалось загрузить раздел безопасности")
		return
	}
	respondJSON(w, items)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	stats, err := h.elements.Stats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Не удалось получить статистику")
		return
	}
	respondJSON(w, stats)
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	respondError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
	return false
}

func respondJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func EnableCORS(next http.Handler) http.Handler { return EnableCORSWithOrigins("*", next) }

func EnableCORSWithOrigins(origins string, next http.Handler) http.Handler {
	allowed := parseOrigins(origins)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (contains(allowed, "*") || contains(allowed, origin)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func parseOrigins(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			result = append(result, origin)
		}
	}
	return result
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (h *Handler) Community(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	items, err := h.content.Community(r.Context(), category, search, 24, 0)
	if err != nil {
		respondError(w, 500, "Не удалось загрузить пользовательские элементы")
		return
	}
	respondJSON(w, items)
}
