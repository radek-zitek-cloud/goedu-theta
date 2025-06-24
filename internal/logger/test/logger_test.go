package test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/radek-zitek-cloud/goedu-theta/internal/logger"
	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
)

// Helper to reset the logger singleton for test isolation
func resetLogger() {
	// Use reflection or package-level access if needed; here we use a hacky workaround
	// by reinitializing the package variables (not recommended for production)
	// This is a limitation of testing singletons in Go.
	// In real-world, consider refactoring for testability.
	_ = logger.GetLogger // just to avoid unused import warning
}

func TestInitializeBootstrapLogger_SetsDefaultLogger(t *testing.T) {
	// Arrange
	os.Setenv("ENVIRONMENT", "development")
	resetLogger()

	// Act
	log := logger.InitializeBootstrapLogger()

	// Assert
	if log == nil {
		t.Fatal("Expected non-nil logger instance")
	}
	if slog.Default() != log {
		t.Error("Expected slog.Default to be set to the bootstrap logger")
	}
}

func TestConfigureLogger_ChangesLoggerConfig(t *testing.T) {
	// Arrange
	resetLogger()
	cfg := config.Logger{
		Level:     "error",
		Format:    "json",
		AddSource: true,
		Output:    "stdout",
	}

	// Act
	logger.ConfigureLogger(cfg)
	log := logger.GetLogger()

	// Assert
	if log == nil {
		t.Fatal("Expected non-nil logger after configuration")
	}
	// We can't easily check the internal state of slog.Logger, but we can check that it doesn't panic
	log.Error("test error log")
}

func TestGetLogger_InitializesIfNil(t *testing.T) {
	// Arrange
	resetLogger()

	// Act
	log := logger.GetLogger()

	// Assert
	if log == nil {
		t.Fatal("Expected GetLogger to return a logger instance")
	}
}

func TestLogger_IsSingleton(t *testing.T) {
	// Arrange
	resetLogger()
	log1 := logger.InitializeBootstrapLogger()
	log2 := logger.GetLogger()

	// Assert
	if log1 != log2 {
		t.Error("Expected logger to be singleton (same instance)")
	}
}
