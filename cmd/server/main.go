package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
	"github.com/radek-zitek-cloud/goedu-theta/internal/logger"
	"github.com/radek-zitek-cloud/goedu-theta/internal/server"
	"github.com/radek-zitek-cloud/goedu-theta/internal/database"
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

	// Initialize MongoDB connection
	slog.Info("🍃 Initializing MongoDB connection...")

	dbManager, err := database.NewMongoDBManager(cfg.Database, logger.GetLogger())
	if err != nil {
		slog.Error("❌ Failed to initialize MongoDB connection", slog.Any("error", err))
		return
	}
	defer func() {
		if err := dbManager.Close(); err != nil {
			slog.Error("❌ Failed to close MongoDB connection", slog.Any("error", err))
		} else {
			slog.Info("🍃 MongoDB connection closed successfully")
		}
	}()

	slog.Info("🍃 MongoDB connection established successfully")



	// This is a placeholder for future application startup code.
	slog.Info("🚀 Server is starting up...")

	// Create the HTTP server instance
	httpServer := server.NewServer(cfg.Server, logger.GetLogger())

	// Start the HTTP server
	if err := httpServer.Start(); err != nil {
		slog.Error("❌ Failed to start HTTP server",
			slog.Any("error", err),
		)
		return
	}

	slog.Info("🪛 HTTP server started successfully",
		slog.String("address", cfg.Server.Host),
		slog.Int("port", cfg.Server.Port),
	)

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-quit
	slog.Info("🛑 Shutdown signal received, initiating graceful shutdown...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// Shutdown the HTTP server gracefully
	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("❌ Error during server shutdown",
			slog.Any("error", err),
		)
		return
	}

	slog.Info("✅ Server shutdown completed successfully")
}
