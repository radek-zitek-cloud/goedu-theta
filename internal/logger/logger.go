package logger

import (
	"log/slog"
	"os"
	"sync"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
)

var (
	instance *slog.Logger // Singleton instance of the logger
	mu       sync.RWMutex // Mutex to ensure thread-safe logger operations
)

// InitializeBootstrapLogger creates and sets the singleton logger with default bootstrap settings.
//
// This function should be called at program startup before the application configuration is loaded.
// It sets up a logger suitable for early-stage logging, using environment-based defaults.
//
// Returns:
//
//	*slog.Logger: The initialized bootstrap logger instance.
//
// Example:
//
//	logger.InitializeBootstrapLogger()
//
// Complexity:
//
//	Time: O(1), Space: O(1)
func InitializeBootstrapLogger() *slog.Logger {
	mu.Lock()
	defer mu.Unlock()

	// Determine log level and options based on environment
	var handlerOptions *slog.HandlerOptions
	isProduction := os.Getenv("ENVIRONMENT") == "production"

	if isProduction {
		// In production, log only errors and above for performance and clarity
		handlerOptions = &slog.HandlerOptions{
			AddSource: false, // Do not include source info in production by default
			Level:     slog.LevelError,
		}
	} else {
		// In development, log debug and above for maximum visibility
		handlerOptions = &slog.HandlerOptions{
			AddSource: false, // Can be set to true if needed
			Level:     slog.LevelDebug,
		}
	}

	// Create a new logger with text output to stdout
	logger := slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	instance = logger
	slog.SetDefault(logger) // Set as the global default logger

	// Log initialization for traceability
	logger.Debug("👢 Bootstrap Logger initialized", slog.String("environment", os.Getenv("ENVIRONMENT")))
	return logger
}

// ConfigureLogger reconfigures the singleton logger with the provided configuration.
//
// This function should be called after the application configuration is loaded.
// It updates the logger's log level, format, and source inclusion based on the config.
//
// Parameters:
//   - config (config.Logger): The logger configuration struct loaded from config files/env.
//
// Example:
//
//	logger.ConfigureLogger(cfg.Logger)
//
// Complexity:
//
//	Time: O(1), Space: O(1)
func ConfigureLogger(config config.Logger) {
	mu.Lock()
	defer mu.Unlock()

	// Map string log level to slog.Level
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Set handler options based on config
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	// Select handler type (JSON or text) based on config
	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Create and set the new logger instance
	logger := slog.New(handler)
	instance = logger
	slog.SetDefault(logger)

	// Log reconfiguration for traceability
	logger.Debug("🔄 Logger reconfigured", slog.String("level", config.Level), slog.String("format", config.Format))
}

// GetLogger returns the singleton logger instance for use throughout the application.
//
// If the logger has not been initialized, it will initialize a bootstrap logger.
//
// Returns:
//
//	*slog.Logger: The singleton logger instance.
//
// Example:
//
//	log := logger.GetLogger()
//
// Complexity:
//
//	Time: O(1), Space: O(1)
func GetLogger() *slog.Logger {
	mu.RLock()
	if instance != nil {
		// If already initialized, return the instance
		defer mu.RUnlock()
		return instance
	}
	mu.RUnlock() // Release the read lock before acquiring the write lock

	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		// If not initialized, initialize with bootstrap logger
		return InitializeBootstrapLogger()
	}
	return instance
}
