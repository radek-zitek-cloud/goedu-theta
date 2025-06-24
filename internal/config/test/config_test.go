package test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
)

func TestNewDefaultConfig_ReturnsDefaults(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
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

func TestLoadFromJSONFile_FileNotFound(t *testing.T) {
	cfg := &config.Config{}
	err := config.LoadFromJSONFile("nonexistent.json", cfg)
	if err == nil {
		t.Error("Expected error for missing config file, got nil")
	}
}

func TestOverrideFromEnv_OverridesStringAndBool(t *testing.T) {
	os.Setenv("SLOG_LEVEL", "error")
	os.Setenv("SLOG_ADD_SOURCE", "false")
	cfg := &config.Config{
		Logger: config.Logger{
			Level:     "debug",
			AddSource: true,
		},
	}
	err := config.OverrideFromEnv(".env", cfg)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cfg.Logger.Level != "error" {
		t.Errorf("Expected Logger.Level to be overridden to 'error', got '%s'", cfg.Logger.Level)
	}
	if cfg.Logger.AddSource != false {
		t.Errorf("Expected Logger.AddSource to be overridden to false, got %v", cfg.Logger.AddSource)
	}
}
