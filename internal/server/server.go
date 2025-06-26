package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
	"github.com/radek-zitek-cloud/goedu-theta/internal/handlers"
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
	// Set Gin mode to release to disable debug output and improve performance
	// This prevents Gin from printing debug information to stdout in production
	// For now setting to debug mode for local development
	gin.SetMode(gin.DebugMode)

	// Create a new Gin router instance without any default middleware
	// gin.New() creates a bare router, unlike gin.Default() which includes logger and recovery
	router := gin.New()

	// Add custom middleware stack in order of execution:
	// 1. Custom slog-based logging middleware for structured logging
	router.Use(ginLoggerMiddleware(logger))
	// 2. Recovery middleware to handle panics gracefully and return 500 errors
	router.Use(gin.Recovery())

	// Create the underlying HTTP server with configuration-driven timeouts
	// These timeouts prevent resource exhaustion from slow or malicious clients
	httpServer := &http.Server{
		// Combine host and port into a network address string (e.g., "localhost:8080")
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: router, // Use Gin router as the HTTP handler

		// ReadTimeout: Maximum duration for reading the entire request (including body)
		// Prevents slow clients from holding connections open indefinitely
		ReadTimeout: time.Duration(cfg.ReadTimeout) * time.Second,

		// WriteTimeout: Maximum duration before timing out writes of the response
		// Prevents slow clients from causing goroutine/memory leaks
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	// Create the Server struct instance with all necessary components
	// This encapsulates the HTTP server, router, config, and logger in one place
	server := &Server{
		router: router,     // Gin router for handling HTTP requests
		server: httpServer, // Standard library HTTP server with timeouts
		config: cfg,        // Configuration settings for server behavior
		logger: logger,     // Structured logger for debugging and monitoring
	}

	// Initialize all HTTP routes and their handlers
	// This must be called after the router is created but before starting the server
	server.setupRoutes()

	// Log successful server creation with key configuration details
	// This helps with debugging and verifying correct configuration
	logger.Debug("🚀 HTTP server created",
		slog.String("addr", httpServer.Addr),        // Network address (host:port)
		slog.Int("read_timeout", cfg.ReadTimeout),   // Read timeout in seconds
		slog.Int("write_timeout", cfg.WriteTimeout), // Write timeout in seconds
	)

	// Return the fully configured and ready-to-start server instance
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
	// Log server startup with the network address for debugging and monitoring
	s.logger.Info("🚀 Starting HTTP server",
		slog.String("addr", s.server.Addr), // Shows exactly where the server will listen
	)

	// Start the HTTP server in a separate goroutine to make this method non-blocking
	// This allows the calling code to continue execution while the server runs
	go func() {
		// Attempt to start the server and listen for incoming connections
		// ListenAndServe blocks until the server is shut down or encounters an error
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Only log errors that aren't from normal server shutdown
			// http.ErrServerClosed is returned when Shutdown() is called, which is expected
			s.logger.Error("❌ HTTP server failed to start",
				slog.String("error", err.Error()), // Detailed error message for debugging
			)
		}
		// If we reach here, the server has stopped (either gracefully or due to error)
	}()

	// Return immediately without waiting for the server to start listening
	// The caller can use other mechanisms to verify the server is ready if needed
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
	// Log the beginning of the graceful shutdown process
	s.logger.Info("🛑 Shutting down HTTP server gracefully...")

	// Attempt to gracefully shutdown the server within the provided context timeout
	// This will:
	// 1. Stop accepting new connections
	// 2. Close idle connections
	// 3. Wait for active connections to finish their current requests
	// 4. Return an error if the context timeout is exceeded
	if err := s.server.Shutdown(ctx); err != nil {
		// Log any errors that occur during shutdown (e.g., timeout, force close needed)
		s.logger.Error("❌ Error during server shutdown",
			slog.String("error", err.Error()), // Detailed error for debugging
		)
		// Return the error to the caller so they can decide how to handle it
		return err
	}

	// Log successful completion of graceful shutdown
	s.logger.Info("✅ HTTP server shutdown completed")

	// Return nil to indicate successful graceful shutdown
	return nil
}

// setupRoutes configures all HTTP routes and their handlers.
//
// This method defines the complete routing table for the HTTP server, establishing
// the API contract and endpoint structure. Each route is carefully designed for
// specific use cases in production environments:
//
// Route Design Principles:
//   - Consistent JSON response format across all endpoints
//   - Proper HTTP status codes following REST conventions
//   - Fast response times for health checks (<100ms)
//   - Structured logging for observability
//   - No external dependencies for core endpoints
//
// Endpoints:
//   - GET /: Root endpoint with API information and service details
//   - GET /health: Health check for load balancers and monitoring
//   - GET /metrics: Application metrics for observability platforms
//
// Security Considerations:
//   - All endpoints are read-only (GET methods only)
//   - No sensitive information exposed in responses
//   - Rate limiting should be applied at reverse proxy level
func (s *Server) setupRoutes() {
	// Create a handler instance with the logger dependency
	// This centralizes all HTTP handler dependencies in one place
	h := handlers.NewHandler(s.logger)

	// Root endpoint - provides basic API information and version details
	// Used for: API discovery, version checking, basic connectivity tests
	// Expected response time: <50ms
	// Dependencies: None (static response)
	s.router.GET("/", h.HandleRoot)

	// Health check endpoint - indicates if the service is operational
	// Used by: Load balancers, monitoring systems, container orchestrators
	// Expected response time: <100ms
	// Dependencies: Should not depend on external services for basic health
	// Note: Complex health checks (DB connectivity) should be separate endpoint
	s.router.GET("/health", h.HandleHealth)

	// Metrics endpoint - exposes application performance and usage statistics
	// Used by: Monitoring systems (Prometheus, Grafana), observability platforms
	// Expected response time: <200ms
	// Dependencies: Internal metrics collection only
	// TODO: Consider implementing Prometheus-compatible format (/metrics with text/plain)
	s.router.GET("/metrics", h.HandleMetrics)

	// Log the completion of route setup for debugging and operational visibility
	// This helps with troubleshooting startup issues and configuration verification
	s.logger.Debug("🛤️  HTTP routes configured",
		slog.Int("route_count", 3), // Track number of registered routes
	)
}

// ginLoggerMiddleware creates a Gin middleware that logs HTTP requests using structured logging.
//
// This middleware provides comprehensive HTTP request/response logging that integrates
// seamlessly with the application's slog-based logging system. It replaces Gin's
// default logger to ensure consistent log formatting and structured data.
//
// Features:
//   - Structured logging with key-value pairs for easy parsing
//   - Request timing and performance metrics
//   - Client identification and security tracking
//   - HTTP status code and error tracking
//   - Query parameter logging for debugging
//   - Consistent log levels based on response status
//
// Log Levels:
//   - Info: Successful requests (2xx status codes)
//   - Warn: Client errors (4xx status codes)
//   - Error: Server errors (5xx status codes)
//
// Parameters:
//   - logger: slog.Logger instance for request logging
//
// Returns:
//   - gin.HandlerFunc: Middleware function compatible with Gin router
//
// Performance Impact: Minimal (<1ms overhead per request)
func ginLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record request start time for latency calculation
		// High precision timing for performance monitoring
		start := time.Now()

		// Capture request path for logging (before any modifications)
		path := c.Request.URL.Path

		// Capture raw query string for complete request reconstruction
		raw := c.Request.URL.RawQuery

		// Continue processing the request through the middleware chain
		// This is where the actual request handling occurs
		c.Next()

		// Calculate request processing duration for performance metrics
		// This includes time spent in all middleware and handlers
		latency := time.Since(start)

		// Extract client information for security and analytics
		clientIP := c.ClientIP()        // Real client IP (handles proxies/load balancers)
		method := c.Request.Method      // HTTP method (GET, POST, etc.)
		statusCode := c.Writer.Status() // HTTP response status code

		// Reconstruct full request path including query parameters
		// This provides complete request URL for debugging and analytics
		if raw != "" {
			path = path + "?" + raw
		}

		// Determine appropriate log level based on HTTP response status code
		// This follows standard HTTP status code semantics for operational logging:
		// - 2xx Success: Info level (normal operations)
		// - 3xx Redirection: Info level (normal operations)
		// - 4xx Client Error: Warn level (client issues, not server problems)
		// - 5xx Server Error: Error level (server issues requiring attention)
		logLevel := slog.LevelInfo
		if statusCode >= 400 && statusCode < 500 {
			// Client errors (bad requests, unauthorized, not found, etc.)
			// These are warnings because they indicate client issues, not server problems
			logLevel = slog.LevelWarn
		} else if statusCode >= 500 {
			// Server errors (internal errors, bad gateway, service unavailable, etc.)
			// These are errors because they indicate server-side problems requiring investigation
			logLevel = slog.LevelError
		}

		// Log the HTTP request with comprehensive structured data
		// This creates a single log entry per request with all relevant information
		// for debugging, monitoring, and security analysis
		logger.Log(context.Background(), logLevel, "🌐 HTTP Request",
			slog.String("method", method),                        // HTTP method for request classification
			slog.String("path", path),                            // Full request path with query params
			slog.Int("status", statusCode),                       // HTTP status code for response analysis
			slog.String("client_ip", clientIP),                   // Client IP for security and analytics
			slog.Duration("latency", latency),                    // Request processing time for performance monitoring
			slog.String("user_agent", c.GetHeader("User-Agent")), // User agent for client identification and analytics
		)
	}
}
