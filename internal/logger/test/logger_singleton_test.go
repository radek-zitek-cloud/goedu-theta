package test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
	"github.com/radek-zitek-cloud/goedu-theta/internal/logger"
)

// Note: Due to Go's package scoping, we cannot reset the logger singleton from outside the logger package.
// These tests focus on observable behavior, not internal state.

func Test_InitializeBootstrapLogger_SetsDefaultLogger(t *testing.T) {
	os.Setenv("ENVIRONMENT", "development")
	log := logger.InitializeBootstrapLogger()
	if log == nil {
		t.Fatal("Expected non-nil logger instance")
	}
	if slog.Default() != log {
		t.Error("Expected slog.Default to be set to the bootstrap logger")
	}
}

func Test_ConfigureLogger_ChangesLoggerConfig(t *testing.T) {
	cfg := config.Logger{
		Level:     "error",
		Format:    "json",
		AddSource: true,
		Output:    "stdout",
	}
	logger.ConfigureLogger(cfg)
	log := logger.GetLogger()
	if log == nil {
		t.Fatal("Expected non-nil logger after configuration")
	}
	log.Error("test error log") // Should not panic
}

func Test_GetLogger_InitializesIfNil(t *testing.T) {
	log := logger.GetLogger()
	if log == nil {
		t.Fatal("Expected GetLogger to return a logger instance")
	}
}

func Test_Logger_IsSingleton(t *testing.T) {
	log1 := logger.InitializeBootstrapLogger()
	log2 := logger.GetLogger()
	if log1 != log2 {
		t.Error("Expected logger to be singleton (same instance)")
	}
}
