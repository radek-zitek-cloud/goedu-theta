package config

import (
	"log/slog"
)

// NewDefaultConfig returns a Config struct with default values.
//
// Parameters:
//   - logger: slog.Logger for debug logging during initialization
//
// Returns:
//   - *Config: pointer to a Config struct with default values
//
// Example:
//
//	cfg := config.NewDefaultConfig(logger)
//
// Complexity:
//
//	Time: O(1), Space: O(1)
func NewDefaultConfig(logger slog.Logger) *Config {
	logger.Debug("🔠 Initializing default configuration")
	return &Config{
		Environment: "development", // Default to development environment
		Logger: Logger{
			Level:     "debug",  // Default log level
			Format:    "text",   // Default log format
			Output:    "stdout", // Default output to standard output
			AddSource: true,     // Include source file and line number in logs
		},
		Server: Server{
			Port:            6910,        // Default HTTP port
			Host:            "localhost", // Default bind address
			ReadTimeout:     30,          // 30 seconds read timeout
			WriteTimeout:    30,          // 30 seconds write timeout
			ShutdownTimeout: 15,          // 15 seconds graceful shutdown timeout
		},
	}
}
