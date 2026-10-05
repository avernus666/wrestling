package main

import (
	"testing"

	"wrestling/internal/config"
)

func TestServerConfigDefaults(t *testing.T) {
	cfg := config.FromEnv()
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}
	if cfg.DBMinConns < 1 || cfg.DBMaxConns < cfg.DBMinConns {
		t.Fatalf("invalid default pool configuration: min=%d max=%d", cfg.DBMinConns, cfg.DBMaxConns)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should validate: %v", err)
	}
}
