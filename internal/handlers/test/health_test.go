package handler_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radek-zitek-cloud/goedu-theta/internal/handlers"
)

// TestHandleHealth tests the health check endpoint handler with comprehensive validation.
//
// This test verifies:
// - Correct HTTP status code (200 OK)
// - Proper Content-Type header (application/json)
// - Required JSON response fields and their values
// - Cache control headers to prevent caching
// - Memory metrics structure and data types
// - Runtime information accuracy
//
// Testing Strategy:
//   - Unit test approach using httptest.ResponseRecorder
//   - No external dependencies (fast execution)
//   - Comprehensive field validation including nested structures
//   - Memory metrics validation
func TestHandleHealth(t *testing.T) {
	// Set Gin to test mode to avoid debug output during testing
	gin.SetMode(gin.TestMode)

	// Create a test logger that discards output to avoid test noise
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create a new handler instance with the test logger
	h := handlers.NewHandler(logger)

	// Create a new Gin router for testing
	router := gin.New()
	router.GET("/health", h.HandleHealth)

	// Create test HTTP request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder to capture handler output
	w := httptest.NewRecorder()

	// Execute the request through the router
	router.ServeHTTP(w, req)

	// Test HTTP status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Test Content-Type header
	expectedContentType := "application/json; charset=utf-8"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected content type '%s', got '%s'", expectedContentType, contentType)
	}

	// Test cache control headers
	expectedCacheControl := "no-cache, no-store, must-revalidate"
	if cacheControl := w.Header().Get("Cache-Control"); cacheControl != expectedCacheControl {
		t.Errorf("Expected Cache-Control '%s', got '%s'", expectedCacheControl, cacheControl)
	}

	if pragma := w.Header().Get("Pragma"); pragma != "no-cache" {
		t.Errorf("Expected Pragma 'no-cache', got '%s'", pragma)
	}

	if expires := w.Header().Get("Expires"); expires != "0" {
		t.Errorf("Expected Expires '0', got '%s'", expires)
	}

	// Parse JSON response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Test top-level required fields
	if status, exists := response["status"]; !exists || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", status)
	}

	// Test timestamp field format (RFC3339)
	if timestamp, exists := response["timestamp"]; !exists {
		t.Error("Missing required field 'timestamp' in response")
	} else if timestampStr, ok := timestamp.(string); !ok {
		t.Error("Field 'timestamp' should be a string")
	} else {
		// Validate RFC3339 format by parsing
		if _, err := time.Parse(time.RFC3339, timestampStr); err != nil {
			t.Errorf("Field 'timestamp' is not in RFC3339 format: %v", err)
		}
	}

	// Test runtime section
	if runtime, exists := response["runtime"]; !exists {
		t.Error("Missing required field 'runtime' in response")
	} else if runtimeMap, ok := runtime.(map[string]interface{}); !ok {
		t.Error("Field 'runtime' should be a map")
	} else {
		// Test runtime fields
		if goVersion, exists := runtimeMap["go_version"]; !exists {
			t.Error("Missing 'go_version' in runtime section")
		} else if _, ok := goVersion.(string); !ok {
			t.Error("Field 'go_version' should be a string")
		}

		if platform, exists := runtimeMap["platform"]; !exists {
			t.Error("Missing 'platform' in runtime section")
		} else if _, ok := platform.(string); !ok {
			t.Error("Field 'platform' should be a string")
		}

		if cpuCores, exists := runtimeMap["cpu_cores"]; !exists {
			t.Error("Missing 'cpu_cores' in runtime section")
		} else if cores, ok := cpuCores.(float64); !ok || cores <= 0 {
			t.Errorf("Field 'cpu_cores' should be a positive number, got %v", cpuCores)
		}
	}

	// Test memory section
	if memory, exists := response["memory"]; !exists {
		t.Error("Missing required field 'memory' in response")
	} else if memoryMap, ok := memory.(map[string]interface{}); !ok {
		t.Error("Field 'memory' should be a map")
	} else {
		// Test memory fields
		memoryFields := []string{"allocated_bytes", "heap_in_use_bytes", "gc_cycles", "allocated_mb", "next_gc_bytes"}
		for _, field := range memoryFields {
			if value, exists := memoryMap[field]; !exists {
				t.Errorf("Missing '%s' in memory section", field)
			} else if _, ok := value.(float64); !ok {
				t.Errorf("Field '%s' should be a number, got %T", field, value)
			}
		}
	}

	// Test goroutines section
	if goroutines, exists := response["goroutines"]; !exists {
		t.Error("Missing required field 'goroutines' in response")
	} else if goroutinesMap, ok := goroutines.(map[string]interface{}); !ok {
		t.Error("Field 'goroutines' should be a map")
	} else {
		if count, exists := goroutinesMap["count"]; !exists {
			t.Error("Missing 'count' in goroutines section")
		} else if countVal, ok := count.(float64); !ok || countVal <= 0 {
			t.Errorf("Field 'count' should be a positive number, got %v", count)
		}
	}

	// Test service section
	if service, exists := response["service"]; !exists {
		t.Error("Missing required field 'service' in response")
	} else if serviceMap, ok := service.(map[string]interface{}); !ok {
		t.Error("Field 'service' should be a map")
	} else {
		// Test service fields
		serviceFields := map[string]string{
			"name":        "goedu-theta",
			"environment": "development",
			"version":     "1.0.0",
		}

		for field, expected := range serviceFields {
			if value, exists := serviceMap[field]; !exists {
				t.Errorf("Missing '%s' in service section", field)
			} else if value != expected {
				t.Errorf("Field '%s': expected '%s', got '%v'", field, expected, value)
			}
		}
	}
}

// TestHandleHealthCacheHeaders tests that health endpoint sets proper cache headers.
//
// Health checks should never be cached as they represent real-time status.
func TestHandleHealthCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	h := handlers.NewHandler(logger)

	router := gin.New()
	router.GET("/health", h.HandleHealth)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Test all cache-related headers
	expectedHeaders := map[string]string{
		"Cache-Control": "no-cache, no-store, must-revalidate",
		"Pragma":        "no-cache",
		"Expires":       "0",
	}

	for header, expected := range expectedHeaders {
		if actual := w.Header().Get(header); actual != expected {
			t.Errorf("Header '%s': expected '%s', got '%s'", header, expected, actual)
		}
	}
}

// TestHandleHealthWithDifferentClients tests health endpoint with various client types.
//
// This verifies that the health check works consistently regardless of the client.
func TestHandleHealthWithDifferentClients(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	h := handlers.NewHandler(logger)

	router := gin.New()
	router.GET("/health", h.HandleHealth)

	testCases := []struct {
		name      string
		userAgent string
		clientIP  string
	}{
		{"Load Balancer", "ELB-HealthChecker/2.0", "10.0.0.1"},
		{"Kubernetes", "kube-probe/1.22", "172.16.0.1"},
		{"Monitoring", "Prometheus/2.30.0", "192.168.1.100"},
		{"Manual Check", "curl/7.68.0", "127.0.0.1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/health", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tc.userAgent != "" {
				req.Header.Set("User-Agent", tc.userAgent)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should always return 200 OK
			if w.Code != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}

			// Response should be valid JSON with health status
			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
			}

			if status := response["status"]; status != "healthy" {
				t.Errorf("Expected status 'healthy', got %v", status)
			}
		})
	}
}

// TestHandleHealthMemoryMetrics tests that memory metrics are reasonable.
//
// This test validates that the memory metrics returned by the health endpoint
// contain sensible values that indicate the application is functioning properly.
func TestHandleHealthMemoryMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	h := handlers.NewHandler(logger)

	router := gin.New()
	router.GET("/health", h.HandleHealth)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	memory, ok := response["memory"].(map[string]interface{})
	if !ok {
		t.Fatal("Memory section missing or invalid")
	}

	// Test that allocated memory is positive
	if allocated, exists := memory["allocated_bytes"]; exists {
		if allocatedVal, ok := allocated.(float64); ok && allocatedVal <= 0 {
			t.Error("Allocated bytes should be positive")
		}
	}

	// Test that heap in use is positive
	if heapInUse, exists := memory["heap_in_use_bytes"]; exists {
		if heapVal, ok := heapInUse.(float64); ok && heapVal <= 0 {
			t.Error("Heap in use bytes should be positive")
		}
	}

	// Test that allocated MB is consistent with allocated bytes
	if allocatedBytes, ok1 := memory["allocated_bytes"].(float64); ok1 {
		if allocatedMB, ok2 := memory["allocated_mb"].(float64); ok2 {
			expectedMB := allocatedBytes / 1024 / 1024
			// Allow for small rounding differences
			if allocatedMB < expectedMB-1 || allocatedMB > expectedMB+1 {
				t.Errorf("Allocated MB (%f) not consistent with allocated bytes (%f)", allocatedMB, allocatedBytes)
			}
		}
	}

	// Test that GC cycles is non-negative
	if gcCycles, exists := memory["gc_cycles"]; exists {
		if gcVal, ok := gcCycles.(float64); ok && gcVal < 0 {
			t.Error("GC cycles should be non-negative")
		}
	}
}

// TestHandleHealthResponseTime tests that health endpoint responds quickly.
//
// Health checks should be fast to avoid timeouts in load balancers and monitoring systems.
func TestHandleHealthResponseTime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	h := handlers.NewHandler(logger)

	router := gin.New()
	router.GET("/health", h.HandleHealth)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Measure response time
	start := time.Now()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	duration := time.Since(start)

	// Health checks should be very fast (under 50ms even in tests)
	maxDuration := 50 * time.Millisecond
	if duration > maxDuration {
		t.Errorf("Health check took too long: %v (max: %v)", duration, maxDuration)
	}

	// Should still return proper response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}
