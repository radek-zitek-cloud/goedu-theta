package main

import (
	"log/slog"
	"os"
)

// Setting up the slog logger for initial logging
func setupBoostrapLogger() *slog.Logger {

	// Set different options based on the environment
	var handlerOptions *slog.HandlerOptions

	// Determine if we are in production mode using the environment variable
	var isProduction = os.Getenv("ENVIRONMENT") == "production"

	if isProduction {
		// Initialize the slog logger with production settings
		// In production, we log errors and above, and include source information
		// but do not log debug or info messages to avoid cluttering logs
		// This is useful for performance and to reduce log noise in production environments
		// The source information will help in debugging issues when they arise
		// The log level is set to Error to ensure that only critical issues are logged
		// This helps in maintaining a clean log output while still capturing necessary information
		handlerOptions = &slog.HandlerOptions{
			AddSource: true, // Include source file and line number in logs
			Level:     slog.LevelError, // Default Error log level
		}
	} else {
		// Initialize the slog logger with development settings
		// In development, we log debug and above, which includes debug, info, warning
		// and error messages. This is useful for development and debugging purposes.
		handlerOptions = &slog.HandlerOptions{
			AddSource: true, // Include source file and line number in logs
			Level:     slog.LevelDebug, // Default Debug log level
		}
	}

	// Create a new slog logger with the specified handler options
	// The logger will output to standard output (os.Stdout) with the specified handler options
	// This allows us to see the logs in the console, which is useful for both development
	// and production environments. In production, the logs will be more concise,
	// while in development, they will provide more detailed information.
	logger := slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	slog.SetDefault(logger)

	logger.Debug("👢 Bootstrap Logger initialized",
		slog.String("environment", os.Getenv("ENVIRONMENT")))

	return logger
}

// This is the main entry point for the server application.
func main() {
	// Initialize the slog bootstrap logger
	logger := setupBoostrapLogger()

	logger.Debug("1️⃣ About to start configuration load",)
}