package test

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
}
