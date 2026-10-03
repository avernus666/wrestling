package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"wrestling/internal/handlers"
	"wrestling/internal/observability"
)

type RouterConfig struct {
	Handler                        *handlers.Handler
	Pool                           Pinger
	PublicDir, WebDir, OpenAPIFile string
	Metrics                        *observability.Metrics
}
type Pinger interface{ Ping(context.Context) error }

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", apiInfo)
	mux.HandleFunc("/api/health", health)
	mux.HandleFunc("/api/ready", func(w http.ResponseWriter, r *http.Request) { ready(w, r, cfg.Pool) })
	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) { cfg.Metrics.Handler(w, r) })
	mux.HandleFunc("/api/openapi.json", func(w http.ResponseWriter, r *http.Request) { openAPI(w, r, cfg.OpenAPIFile) })
	mux.HandleFunc("/api/openapi.yaml", func(w http.ResponseWriter, r *http.Request) { openAPI(w, r, cfg.OpenAPIFile) })
	mux.HandleFunc("/api/elements", cfg.Handler.GetElements)
	mux.HandleFunc("/api/elements/", elementRoute(cfg.Handler))
	mux.HandleFunc("/api/bars", cfg.Handler.GetBars)
	mux.HandleFunc("/api/base", cfg.Handler.GetBase)
	mux.HandleFunc("/api/safety", cfg.Handler.GetSafety)
	mux.HandleFunc("/api/stats", cfg.Handler.GetStats)
	mux.HandleFunc("/api/auth/register", cfg.Handler.Register)
	mux.HandleFunc("/api/auth/login", cfg.Handler.Login)
	mux.HandleFunc("/api/auth/me", cfg.Handler.Me)
	mux.HandleFunc("/api/auth/logout", cfg.Handler.Logout)
	mux.HandleFunc("/api/progress", progressRoute(cfg.Handler))
	mux.HandleFunc("/api/workouts", cfg.Handler.CreateWorkout)
	mux.HandleFunc("/api/workouts/", workoutRoute(cfg.Handler))
	mux.HandleFunc("/api/history", cfg.Handler.History)
	mux.HandleFunc("/api/analytics", cfg.Handler.Analytics)
	mux.HandleFunc("/api/skills", skillRoute(cfg.Handler))
	mux.HandleFunc("/api/community", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			cfg.Handler.CreateCommunity(w, r)
			return
		}
		cfg.Handler.Community(w, r)
	})
	mux.HandleFunc("/api/community/", communityRoute(cfg.Handler))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(filepath.Join(cfg.PublicDir, "images")))))
	mux.Handle("/", spaHandler(cfg.WebDir))
	return mux
}

func elementRoute(h *handlers.Handler) http.HandlerFunc {
	const prefix = "/api/elements/category/"
	return func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > len(prefix) && r.URL.Path[:len(prefix)] == prefix {
			h.GetElementsByCategory(w, r)
			return
		}
		h.GetElement(w, r)
	}
}
func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok", "service": "wrestling-api"})
}
func ready(w http.ResponseWriter, r *http.Request, p Pinger) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := p.Ping(ctx); err != nil {
		respond(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "service": "wrestling-api"})
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ready", "service": "wrestling-api"})
}
func openAPI(w http.ResponseWriter, r *http.Request, configured string) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	candidates := []string{configured, "../openapi.yaml", "../../openapi.yaml", "../../../openapi.yaml"}
	for _, file := range candidates {
		if file == "" {
			continue
		}
		if data, err := os.ReadFile(file); err == nil {
			w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
			_, _ = w.Write(data)
			return
		}
	}
	http.Error(w, "OpenAPI specification is unavailable", http.StatusServiceUnavailable)
}
func apiInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	respond(w, http.StatusOK, map[string]any{"name": "Wrestling API", "version": "1.0.0", "endpoints": map[string]string{"elements": "/api/elements", "element": "/api/elements/{id}", "category": "/api/elements/category/{cat}", "bars": "/api/bars", "base": "/api/base", "safety": "/api/safety", "stats": "/api/stats", "metrics": "/api/metrics"}})
}
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func methodNotAllowed(w http.ResponseWriter, method string) {
	w.Header().Set("Allow", method)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
func spaHandler(webDir string) http.Handler {
	fs := http.FileServer(http.Dir(webDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			methodNotAllowed(w, http.MethodGet)
			return
		}
		relative := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		target := filepath.Join(webDir, relative)
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		r.URL.Path = "/"
		fs.ServeHTTP(w, r)
	})
}

func progressRoute(h *handlers.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetProgress(w, r)
			return
		}
		if r.Method == http.MethodPut {
			h.SetProgress(w, r)
			return
		}
		methodNotAllowed(w, http.MethodGet)
	}
}
func workoutRoute(h *handlers.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CompleteWorkout(w, r)
			return
		}
		methodNotAllowed(w, http.MethodPost)
	}
}
func skillRoute(h *handlers.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.Skills(w, r)
			return
		}
		if r.Method == http.MethodPut {
			h.SetSkill(w, r)
			return
		}
		methodNotAllowed(w, http.MethodGet)
	}
}

func communityRoute(h *handlers.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/report") {
			h.ReportCommunity(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/comments") {
			h.CommunityComments(w, r)
			return
		}
		methodNotAllowed(w, http.MethodGet)
	}
}
