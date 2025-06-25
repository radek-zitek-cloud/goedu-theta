package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
)

// Server represents the HTTP server instance.
//
// This struct encapsulates the Gin router, HTTP server, configuration,
// and logger for a complete web server implementation.
type Server struct {
	router *gin.Engine   // Gin HTTP router
	server *http.Server  // Standard library HTTP server
	config config.Server // Server configuration
	logger *slog.Logger  // Structured logger instance
}

// NewServer creates a new HTTP server instance with Gin router.
//
// Parameters:
//   - cfg: Server configuration struct containing port, timeouts, etc.
//   - logger: Structured logger for server operations
//
// Returns:
//   - *Server: Configured server instance ready to start
//
// Example:
//
//	server := NewServer(config.Server{Port: 8080}, logger)
//
// Complexity:
//   - Time: O(1), Space: O(1)
func NewServer(cfg config.Server, logger *slog.Logger) *Server {
	// Set Gin mode based on environment (production mode disables debug logging)
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router with recovery middleware
	router := gin.New()

	// Add logging middleware that integrates with slog
	router.Use(ginLoggerMiddleware(logger))
	router.Use(gin.Recovery())

	// Create HTTP server with configured timeouts
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	server := &Server{
		router: router,
		server: httpServer,
		config: cfg,
		logger: logger,
	}

	// Setup routes
	server.setupRoutes()

	logger.Debug("🚀 HTTP server created",
		slog.String("addr", httpServer.Addr),
		slog.Int("read_timeout", cfg.ReadTimeout),
		slog.Int("write_timeout", cfg.WriteTimeout),
	)

	return server
}

// Start starts the HTTP server in a non-blocking manner.
//
// The server will start listening on the configured address and port.
// This method is non-blocking and returns immediately.
//
// Returns:
//   - error: Any error that occurred during server startup
//
// Example:
//
//	if err := server.Start(); err != nil {
//	    log.Fatal("Failed to start server:", err)
//	}
func (s *Server) Start() error {
	s.logger.Info("🚀 Starting HTTP server",
		slog.String("addr", s.server.Addr),
	)

	// Start server in goroutine so it doesn't block
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("❌ HTTP server failed to start",
				slog.String("error", err.Error()),
			)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the HTTP server.
//
// This method waits for existing connections to finish processing
// within the configured shutdown timeout period.
//
// Parameters:
//   - ctx: Context for shutdown operation (with timeout)
//
// Returns:
//   - error: Any error that occurred during shutdown
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
//	defer cancel()
//	if err := server.Shutdown(ctx); err != nil {
//	    log.Error("Server shutdown error:", err)
//	}
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("🛑 Shutting down HTTP server gracefully...")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("❌ Error during server shutdown",
			slog.String("error", err.Error()),
		)
		return err
	}

	s.logger.Info("✅ HTTP server shutdown completed")
	return nil
}

// setupRoutes configures all HTTP routes and their handlers.
//
// This method sets up the REST endpoints for the application:
//   - GET /: Root endpoint with welcome message
//   - GET /health: Health check endpoint
//   - GET /metrics: Metrics endpoint (placeholder)
//
// All routes include proper HTTP status codes, JSON responses,
// and structured logging.
func (s *Server) setupRoutes() {
	// Root endpoint - Welcome message
	s.router.GET("/", s.handleRoot)

	// Health check endpoint
	s.router.GET("/health", s.handleHealth)

	// Metrics endpoint (placeholder)
	s.router.GET("/metrics", s.handleMetrics)

	s.logger.Debug("🛤️  HTTP routes configured",
		slog.Int("route_count", 3),
	)
}

// handleRoot handles GET / requests.
//
// Returns a welcome message with basic application information.
// This endpoint can be used to verify the server is running.
func (s *Server) handleRoot(c *gin.Context) {
	response := gin.H{
		"message":   "Welcome to GoEdu-Theta API Server",
		"status":    "running",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"endpoints": []string{
			"GET /",
			"GET /health",
			"GET /metrics",
		},
	}

	s.logger.Debug("📡 Root endpoint accessed",
		slog.String("client_ip", c.ClientIP()),
		slog.String("user_agent", c.GetHeader("User-Agent")),
	)

	c.JSON(http.StatusOK, response)
}

// handleHealth handles GET /health requests.
//
// Returns the health status of the application and its dependencies.
// This endpoint is typically used by load balancers and monitoring systems.
func (s *Server) handleHealth(c *gin.Context) {
	response := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(time.Now()).String(), // TODO: Track actual uptime
		"checks": gin.H{
			"database": "ok", // TODO: Implement actual database health check
			"memory":   "ok", // TODO: Implement memory usage check
			"disk":     "ok", // TODO: Implement disk space check
		},
	}

	s.logger.Debug("🏥 Health endpoint accessed",
		slog.String("client_ip", c.ClientIP()),
	)

	c.JSON(http.StatusOK, response)
}

// handleMetrics handles GET /metrics requests.
//
// Returns application metrics in a structured format.
// This endpoint provides operational metrics for monitoring and observability.
func (s *Server) handleMetrics(c *gin.Context) {
	response := gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"metrics": gin.H{
			"http_requests_total":   0,   // TODO: Implement request counter
			"http_request_duration": 0.0, // TODO: Implement request duration tracking
			"active_connections":    0,   // TODO: Implement connection tracking
			"memory_usage_bytes":    0,   // TODO: Implement memory usage tracking
			"goroutines_count":      0,   // TODO: Implement goroutine counting
		},
		"build_info": gin.H{
			"version":  "1.0.0",
			"commit":   "unknown", // TODO: Add build-time commit hash
			"built_at": "unknown", // TODO: Add build-time timestamp
		},
	}

	s.logger.Debug("📊 Metrics endpoint accessed",
		slog.String("client_ip", c.ClientIP()),
	)

	c.JSON(http.StatusOK, response)
}

// ginLoggerMiddleware creates a Gin middleware that logs HTTP requests using slog.
//
// This middleware integrates Gin's request logging with the application's
// structured logging system, providing consistent log format and levels.
//
// Parameters:
//   - logger: slog.Logger instance for request logging
//
// Returns:
//   - gin.HandlerFunc: Middleware function for Gin router
func ginLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details after processing
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// Determine log level based on status code
		logLevel := slog.LevelInfo
		if statusCode >= 400 && statusCode < 500 {
			logLevel = slog.LevelWarn
		} else if statusCode >= 500 {
			logLevel = slog.LevelError
		}

		logger.Log(context.Background(), logLevel, "🌐 HTTP Request",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.String("client_ip", clientIP),
			slog.Duration("latency", latency),
			slog.String("user_agent", c.GetHeader("User-Agent")),
		)
	}
}
