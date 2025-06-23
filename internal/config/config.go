package config

import (
	"os"
	"log/slog"
)

type Config struct {
	// Environment variable to determine if the application is in production mode
	Environment string `json:"environment" yaml:"environment"`
	Logger      Logger `json:"logger" yaml:"logger"`
}

type Logger struct {
	// Level indicates the log level (e.g., Debug, Info, Error)
	Level string `json:"level" yaml:"level"`
	// Format indicates the log format (e.g., JSON, Text)
	Format string `json:"format" yaml:"format"`
	// Output indicates where to output the logs (e.g., stdout, file)
	Output string `json:"output" yaml:"output"`
	// AddSource indicates whether to include source file and line number in logs
	AddSource bool `json:"add_source" yaml:"add_source"`
}

func NewDefaultConfig(logger slog.Logger) *Config {
	logger.Debug("🔠 Initializing default configuration")
	return &Config{
		Environment: "development", // Default to development environment
		Logger: Logger{
			Level:      "debug",    // Default log level
			Format:     "text",     // Default log format
			Output:     "stdout",   // Default output to standard output
			AddSource:  true,       // Include source file and line number in logs
		},
	}
}

func NewConfig(logger slog.Logger) *Config {
	logger.Debug("🔠 Loading configuration")

	// Setup new default configuration
	config := NewDefaultConfig(logger)

	// Load environment variables and override default configuration
	var environment string = os.Getenv("ENVIRONMENT")
	logger.Debug("🔠 Environment variable loaded",
		slog.String("environment", environment),
	)
	// Validate the environment variable and set the environment
	// If the environment variable is not set or invalid, default to "development"
	switch environment {
	case "development", "test", "staging", "production":
		logger.Debug("🔠 Valid environment detected",
			slog.String("environment", environment),
		)
	default:
		logger.Warn("🔠 Invalid or unset environment variable, defaulting to development",
			slog.String("environment", environment),
		)
		environment = "development"
	}
	logger.Debug("🔠 Setting environment",
		slog.String("environment", environment),
	)

	// Log the loaded configuration
	logger.Debug("🔠 Configuration loaded",
		slog.String("environment", config.Environment),
		slog.String("log_level", config.Logger.Level),
		slog.String("log_format", config.Logger.Format),
		slog.String("log_output", config.Logger.Output),
		slog.Bool("log_add_source", config.Logger.AddSource),
	)

	return config
}