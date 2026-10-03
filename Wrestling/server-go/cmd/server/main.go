package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wrestling/internal/adapters/postgres"
	"wrestling/internal/config"
	"wrestling/internal/db"
	"wrestling/internal/handlers"
	"wrestling/internal/httpserver"
	"wrestling/internal/observability"
	"wrestling/internal/services"
)

func main() {
	logger := log.New(os.Stdout, "wrestling ", log.LstdFlags|log.LUTC)
	cfg := config.FromEnv()
	if err := cfg.Validate(); err != nil {
		logger.Fatal(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	pool, err := db.NewPool(ctx, db.Config{URL: cfg.DatabaseURL, MaxConns: int32(cfg.DBMaxConns), MinConns: int32(cfg.DBMinConns), MaxConnLifetime: cfg.DBMaxConnLifetime, MaxConnIdleTime: cfg.DBMaxConnIdleTime})
	if err != nil {
		logger.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	migrationCtx, migrationCancel := context.WithTimeout(ctx, 30*time.Second)
	defer migrationCancel()
	if err := db.Migrate(migrationCtx, pool); err != nil {
		logger.Fatalf("database migration failed: %v", err)
	}
	repo := postgres.NewContentRepository(pool)
	contentService := services.NewContentService(repo)
	elementService := services.NewElementService(repo.Elements())
	userRepo := postgres.NewUserRepository(pool)
	authService := services.NewAuthService(userRepo, cfg.SessionTTL)
	userService := services.NewUserService(userRepo)
	youtubeVerifier := services.NewYouTubeVerifier(cfg.YouTubeAPIKey)
	h := handlers.New(contentService, elementService, authService, userService, youtubeVerifier)
	metrics := observability.NewMetrics()
	router := httpserver.NewRouter(httpserver.RouterConfig{Handler: h, Pool: pool, PublicDir: cfg.PublicDir, WebDir: cfg.WebDir, OpenAPIFile: cfg.OpenAPIFile, Metrics: metrics})
	server := &http.Server{Addr: ":" + cfg.Port, Handler: httpserver.Stack(logger, metrics, cfg.CORSOrigins, router), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Printf("server shutdown failed: %v", err)
		}
	}()
	logger.Printf("Wrestling API listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("server failed: %v", err)
	}
}
