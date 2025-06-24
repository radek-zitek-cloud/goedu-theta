package main

import (
	"log/slog"
	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
	"github.com/radek-zitek-cloud/goedu-theta/internal/logger"
)



// This is the main entry point for the server application.
func main() {
	// Initialize the slog bootstrap logger
	logger.InitializeBootstrapLogger()

	slog.Debug("🔠 About to start configuration load",)

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("🔠 Error loading configuration",
			slog.Any("error", err),
		)
		return
	}

	slog.Debug("🔠 Configuration loaded successfully",
		slog.Any("config", cfg),
	)

	logger.ConfigureLogger(cfg.Logger)
	slog.Debug("🔠 Logger configured successfully",
		slog.Any("logger", cfg.Logger),
	)
}