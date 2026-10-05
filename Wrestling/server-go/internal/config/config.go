package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	DatabaseURL       string
	CORSOrigins       string
	PublicDir         string
	WebDir            string
	OpenAPIFile       string
	DBMaxConns        int32
	DBMinConns        int32
	DBMaxConnLifetime time.Duration
	DBMaxConnIdleTime time.Duration
	ShutdownTimeout   time.Duration
	SessionTTL        time.Duration
	YouTubeAPIKey     string
}

func FromEnv() Config {
	return Config{
		Port:              stringEnv("PORT", "8080"),
		DatabaseURL:       stringEnv("DATABASE_URL", "postgres://wrestling:wrestling@localhost:5432/wrestling?sslmode=disable"),
		CORSOrigins:       stringEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8080"),
		PublicDir:         stringEnv("PUBLIC_DIR", "../public"),
		WebDir:            stringEnv("WEB_DIR", "../build"),
		OpenAPIFile:       stringEnv("OPENAPI_FILE", ""),
		DBMaxConns:        int32Env("DB_MAX_CONNS", 20),
		DBMinConns:        int32Env("DB_MIN_CONNS", 2),
		DBMaxConnLifetime: durationEnv("DB_MAX_CONN_LIFETIME", 30*time.Minute),
		DBMaxConnIdleTime: durationEnv("DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
		ShutdownTimeout:   durationEnv("SHUTDOWN_TIMEOUT", 10*time.Second),
		SessionTTL:        durationEnv("SESSION_TTL", 30*24*time.Hour),
		YouTubeAPIKey:     stringEnv("YOUTUBE_API_KEY", ""),
	}
}

func stringEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func int32Env(key string, fallback int32) int32 {
	value, err := strconv.ParseInt(os.Getenv(key), 10, 32)
	if err == nil && value > 0 {
		return int32(value)
	}
	return fallback
}
func durationEnv(key string, fallback time.Duration) time.Duration {
	if value, err := time.ParseDuration(os.Getenv(key)); err == nil && value > 0 {
		return value
	}
	return fallback
}
func (c Config) Validate() error {
	if c.DBMinConns > c.DBMaxConns {
		return fmt.Errorf("DB_MIN_CONNS cannot exceed DB_MAX_CONNS")
	}
	return nil
}
