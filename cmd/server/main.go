package main

import (
	"log/slog"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
	"github.com/radek-zitek-cloud/goedu-theta/internal/logger"
)

// main is the entry point for the GoEdu-Theta server application.
//
// Responsibilities:
//   - Initializes a bootstrap logger for early-stage logging (before config is loaded)
//   - Loads the application configuration from JSON files and environment variables
//   - Reconfigures the logger based on loaded configuration
//   - Provides detailed debug/error logging for each step
//
// Error Handling:
//   - If configuration loading fails, logs the error and exits gracefully
//
// Usage:
//
//	Run the compiled binary. The application expects configuration files in the 'configs/' directory
//	and optionally a .env file in the project root.
//
// Example:
//
//	$ go run cmd/server/main.go
//
// Complexity:
//
//	Time: O(1) (all operations are constant time except for file I/O)
//	Space: O(1) (config struct is small)
func main() {
	// Initialize the slog bootstrap logger for early logging.
	// This logger uses default settings and is replaced after config is loaded.
	logger.InitializeBootstrapLogger()

	// Log the start of the configuration loading process.
	slog.Debug("🔠 About to start configuration load")

	// Load the application configuration from JSON files and environment variables.
	// This function merges base, environment-specific, and local config files,
	// then overrides with environment variables and .env file values.
	cfg, err := config.NewConfig()
	if err != nil {
		// Log the error and exit if configuration loading fails.
		slog.Error("🔠 Error loading configuration",
			slog.Any("error", err),
		)
		return
	}

	// Log the loaded configuration for debugging purposes.
	// This is useful for verifying that all config values are as expected.
	slog.Debug("🔠 Configuration loaded successfully",
		slog.Any("config", cfg),
	)

	// Reconfigure the logger with the loaded configuration.
	// This allows log level, format, and other options to be set via config.
	logger.ConfigureLogger(cfg.Logger)

	// Confirm that the logger has been reconfigured.
	slog.Debug("🔠 Logger configured successfully",
		slog.Any("logger", cfg.Logger),
	)

	// TODO (2025-06-24): Start the main server logic here (HTTP server, gRPC, etc.)
	// This is a placeholder for future application startup code.
}
