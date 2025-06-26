package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler holds dependencies needed by HTTP handlers.
//
// This struct centralizes all dependencies that handlers need, providing a clean
// separation of concerns and making handlers easily testable by allowing dependency injection.
//
// Dependencies:
//   - Logger: Structured logger for request/response logging and debugging
//   - Config: Application configuration (can be added later if needed)
//   - Services: Business logic services (can be added later if needed)
//
// Design Benefits:
//   - Testability: Easy to mock dependencies in unit tests
//   - Separation of Concerns: Handlers focus on HTTP concerns, not infrastructure
//   - Consistency: All handlers use the same dependency injection pattern
//   - Maintainability: Dependencies are explicit and centralized
type Handler struct {
	logger *slog.Logger // Structured logger instance for HTTP request logging
}

// NewHandler creates a new Handler instance with the provided dependencies.
//
// Parameters:
//   - logger: Structured logger for request/response logging
//
// Returns:
//   - *Handler: Configured handler instance ready to handle HTTP requests
//
// Example:
//
//	h := handlers.NewHandler(logger)
//	router.GET("/", h.HandleRoot)
func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// HandleRoot handles GET / requests and provides API discovery information.
//
// This endpoint serves as the main entry point for API consumers, providing:
// - Service identification and version information
// - Current operational status
// - Available endpoints for API discovery
// - Standardized timestamp in RFC3339 format
//
// Use Cases:
//   - API documentation and discovery
//   - Basic connectivity tests
//   - Service version verification
//   - Integration testing setup
//
// Response Format: JSON with HTTP 200 status
// Performance: <50ms response time (no external dependencies)
// Caching: Safe to cache for short periods (30s-1min)
func (h *Handler) HandleRoot(c *gin.Context) {
	// Construct the response payload with comprehensive service information
	// Using gin.H (map[string]interface{}) for flexible JSON serialization
	response := gin.H{
		// Human-readable welcome message for API consumers
		"message": "Welcome to GoEdu-Theta API Server",

		// Current operational status - always "running" if endpoint is accessible
		// This provides immediate feedback that the service is operational
		"status": "running",

		// Semantic version following semver.org conventions
		// TODO: Consider loading from build-time variables or config
		"version": "1.0.0",

		// Current timestamp in UTC using RFC3339 format (ISO 8601 compliant)
		// Useful for: timezone consistency, request timing, cache validation
		"timestamp": time.Now().UTC().Format(time.RFC3339),

		// Complete list of available endpoints for API discovery
		// Helps clients understand the available functionality without documentation
		"endpoints": []string{
			"GET /",        // This endpoint (self-reference)
			"GET /health",  // Health check endpoint
			"GET /metrics", // Metrics endpoint
		},
	}

	// Log the access for debugging and monitoring purposes
	// Captures client information for security analysis and usage patterns
	h.logger.Debug("📡 Root endpoint accessed",
		slog.String("client_ip", c.ClientIP()),               // Client IP for security monitoring
		slog.String("user_agent", c.GetHeader("User-Agent")), // User agent for analytics
	)

	// Send the JSON response with HTTP 200 status
	// Gin automatically sets Content-Type: application/json header
	// The response is serialized to JSON and sent to the client
	c.JSON(http.StatusOK, response)
}
