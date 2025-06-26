package handlers

import (
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// HandleHealth handles GET /health requests and provides comprehensive health status.
//
// This endpoint implements a comprehensive health check following industry best practices
// for service health monitoring. It provides detailed information about the service's
// operational status, resource utilization, and runtime characteristics.
//
// Health Check Categories:
//   - Service Status: Basic operational status
//   - Resource Usage: Memory, goroutines, and system resources
//   - Timing Information: Startup time and current timestamp
//   - Runtime Metrics: Go version, OS information
//
// Use Cases:
//   - Load balancer health checks
//   - Kubernetes liveness/readiness probes
//   - Service discovery registration
//   - Monitoring and alerting systems
//   - Automated deployment validation
//   - Performance monitoring and capacity planning
//
// Response Format: JSON with HTTP 200 status (always successful)
// Performance: <10ms response time (local operations only)
// Caching: Should NOT be cached (real-time health status)
//
// Standards Compliance:
//   - Follows RFC 7234 for non-cacheable responses
//   - Compatible with Kubernetes health check patterns
//   - Aligns with microservice health check best practices
func (h *Handler) HandleHealth(c *gin.Context) {
	// Capture current memory statistics for resource monitoring
	// This is a snapshot of Go runtime memory usage at request time
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Construct comprehensive health status response
	// Each field provides specific operational insights for monitoring systems
	response := gin.H{
		// Primary health indicator - always "healthy" if endpoint is reachable
		// This is the main field that load balancers and orchestrators check
		"status": "healthy",

		// Current UTC timestamp in RFC3339 format for request timing analysis
		// Helps identify stale responses and measure response times
		"timestamp": time.Now().UTC().Format(time.RFC3339),

		// Runtime environment information for debugging and compatibility verification
		"runtime": gin.H{
			// Go version used to build the service - critical for compatibility tracking
			"go_version": runtime.Version(),

			// Operating system and architecture - important for deployment verification
			"platform": runtime.GOOS + "/" + runtime.GOARCH,

			// Number of logical CPUs available to the process
			// Useful for understanding available parallelism and resource allocation
			"cpu_cores": runtime.NumCPU(),
		},

		// Memory usage statistics from Go runtime - essential for resource monitoring
		"memory": gin.H{
			// Current heap memory allocated to objects (bytes)
			// This is memory actively used by the application
			"allocated_bytes": memStats.Alloc,

			// Total number of bytes allocated and still in use
			// Different from Alloc due to garbage collection timing
			"heap_in_use_bytes": memStats.HeapInuse,

			// Number of completed garbage collection cycles
			// High values may indicate memory pressure or allocation patterns
			"gc_cycles": memStats.NumGC,

			// Human-readable memory allocation for easier interpretation
			// Provides quick visual assessment without byte-to-MB conversion
			"allocated_mb": bToMb(memStats.Alloc),

			// Next garbage collection target size (bytes)
			// Helps understand GC behavior and memory management efficiency
			"next_gc_bytes": memStats.NextGC,
		},

		// Go runtime concurrency metrics - crucial for performance monitoring
		"goroutines": gin.H{
			// Current number of goroutines in the Go runtime
			// High values may indicate goroutine leaks or high concurrency
			"count": runtime.NumGoroutine(),

			// Note: This count includes system goroutines (GC, scheduler, etc.)
			// Typical baseline is 2-5 goroutines for minimal applications
		},

		// Service-specific operational metrics
		"service": gin.H{
			// Service identifier for multi-service environments
			"name": "goedu-theta",

			// Environment indicator - useful for distinguishing deployment stages
			// TODO: Consider loading from configuration or environment variables
			"environment": "development",

			// Service version for deployment tracking and compatibility verification
			// TODO: Consider loading from build-time variables or git tags
			"version": "1.0.0",
		},
	}

	// Log health check access with performance metrics for monitoring
	// This helps track health check frequency and identify potential issues
	h.logger.Debug("🏥 Health check endpoint accessed",
		slog.String("client_ip", c.ClientIP()),               // Client IP for access pattern analysis
		slog.Int("goroutines", runtime.NumGoroutine()),       // Current goroutine count
		slog.Uint64("memory_mb", bToMb(memStats.Alloc)),      // Current memory usage in MB
		slog.Uint64("gc_cycles", uint64(memStats.NumGC)),     // GC cycle count
		slog.String("user_agent", c.GetHeader("User-Agent")), // User agent for client identification
	)

	// Set cache control headers to prevent health status caching
	// Health checks should always return current status, not cached values
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1
	c.Header("Pragma", "no-cache")                                   // HTTP 1.0 compatibility
	c.Header("Expires", "0")                                         // Proxy cache prevention

	// Send the JSON response with HTTP 200 status
	// Always returns 200 OK - service unavailability is indicated by no response
	c.JSON(http.StatusOK, response)
}

// bToMb converts bytes to megabytes for human-readable memory reporting.
//
// This utility function converts raw byte values from runtime.MemStats to
// megabytes for easier interpretation in monitoring dashboards and logs.
//
// Parameters:
//   - b: Memory size in bytes (uint64)
//
// Returns:
//   - Memory size in megabytes (uint64)
//
// Calculation: Uses binary megabytes (1 MB = 1024 * 1024 bytes)
// Note: This is binary MB (MiB), not decimal MB used by storage vendors
//
// Example:
//
//	bytes := uint64(52428800)  // 50 MB
//	mb := bToMb(bytes)         // Returns: 50
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
