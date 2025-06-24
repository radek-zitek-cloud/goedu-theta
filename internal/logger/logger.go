package logger

import (
	"log/slog"
	"os"
	"sync"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
)

// Config holds logger configuration
type Config struct {
	Level     string
	Format    string
	AddSource bool
	Output    string
}

var (
	instance *slog.Logger
	mu       sync.RWMutex
)

// InitializeBootstrapLogger creates and sets the singleton logger with default bootstrap settings.
// Should be called at program startup before config is loaded.
func InitializeBootstrapLogger() *slog.Logger {
	mu.Lock()
	defer mu.Unlock()

	var handlerOptions *slog.HandlerOptions
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	if isProduction {
		handlerOptions = &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelError,
		}
	} else {
		handlerOptions = &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelDebug,
		}
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	instance = logger
	slog.SetDefault(logger)
	logger.Debug("👢 Bootstrap Logger initialized", slog.String("environment", os.Getenv("ENVIRONMENT")))
	return logger
}

// ConfigureLogger reconfigures the singleton logger with the provided config.
// This should be called after config is loaded.
func ConfigureLogger(config config.Logger) {
	mu.Lock()
	defer mu.Unlock()

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

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	instance = logger
	slog.SetDefault(logger)
	logger.Debug("🔄 Logger reconfigured", slog.String("level", config.Level), slog.String("format", config.Format))
}

// GetLogger returns the singleton logger instance.
func GetLogger() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if instance == nil {
		// If not initialized, initialize with bootstrap logger
		return InitializeBootstrapLogger()
	}
	return instance
}
