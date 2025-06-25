package config_test

import (
	"log/slog"
	"testing"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
)

func TestNewDefaultConfig_Values(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(nil, nil))
	cfg := config.NewDefaultConfig(*logger)
	if cfg.Environment != "development" {
		t.Errorf("Expected default environment 'development', got '%s'", cfg.Environment)
	}
	if cfg.Logger.Level != "debug" {
		t.Errorf("Expected default log level 'debug', got '%s'", cfg.Logger.Level)
	}
	if cfg.Logger.Format != "text" {
		t.Errorf("Expected default log format 'text', got '%s'", cfg.Logger.Format)
	}
	if cfg.Logger.Output != "stdout" {
		t.Errorf("Expected default log output 'stdout', got '%s'", cfg.Logger.Output)
	}
	if !cfg.Logger.AddSource {
		t.Error("Expected AddSource to be true by default")
	}
	// Test Server defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default server port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected default server host 'localhost', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.ReadTimeout != 30 {
		t.Errorf("Expected default read timeout 30, got %d", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 30 {
		t.Errorf("Expected default write timeout 30, got %d", cfg.Server.WriteTimeout)
	}
	if cfg.Server.ShutdownTimeout != 15 {
		t.Errorf("Expected default shutdown timeout 15, got %d", cfg.Server.ShutdownTimeout)
	}
}
