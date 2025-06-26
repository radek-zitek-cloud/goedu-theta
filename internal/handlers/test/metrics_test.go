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

// TestHandleMetrics tests the metrics endpoint handler with comprehensive validation.
//
// This test verifies:
// - Correct HTTP status code (200 OK)
// - Proper Content-Type header (application/json)
// - Required JSON response fields and their nested structures
// - Cache control headers to prevent caching
// - Memory metrics structure and reasonable values
// - Runtime information accuracy
// - Metadata collection timing
//
// Testing Strategy:
//   - Unit test approach using httptest.ResponseRecorder
//   - No external dependencies (fast execution)
//   - Comprehensive field validation including deeply nested structures
//   - Memory and runtime metrics validation
func TestHandleMetrics(t *testing.T) {
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
	router.GET("/metrics", h.HandleMetrics)

	// Create test HTTP request
	req, err := http.NewRequest("GET", "/metrics", nil)
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

	// Parse JSON response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Test metadata section
	if metadata, exists := response["metadata"]; !exists {
		t.Error("Missing required field 'metadata' in response")
	} else if metadataMap, ok := metadata.(map[string]interface{}); !ok {
		t.Error("Field 'metadata' should be a map")
	} else {
		// Test metadata fields
		if collectedAt, exists := metadataMap["collected_at"]; !exists {
			t.Error("Missing 'collected_at' in metadata section")
		} else if timestampStr, ok := collectedAt.(string); !ok {
			t.Error("Field 'collected_at' should be a string")
		} else {
			// Validate RFC3339 format
			if _, err := time.Parse(time.RFC3339, timestampStr); err != nil {
				t.Errorf("Field 'collected_at' is not in RFC3339 format: %v", err)
			}
		}

		if schemaVersion, exists := metadataMap["schema_version"]; !exists || schemaVersion != "1.0.0" {
			t.Errorf("Expected schema_version '1.0.0', got %v", schemaVersion)
		}

		if collectionDuration, exists := metadataMap["collection_duration_ns"]; !exists {
			t.Error("Missing 'collection_duration_ns' in metadata section")
		} else if _, ok := collectionDuration.(float64); !ok {
			t.Error("Field 'collection_duration_ns' should be a number")
		}
	}

	// Test memory section structure
	testMemorySection(t, response)

	// Test runtime section structure
	testRuntimeSection(t, response)

	// Test application section structure
	testApplicationSection(t, response)

	// Test resources section structure
	testResourcesSection(t, response)
}

// testMemorySection validates the memory metrics section structure and values.
func testMemorySection(t *testing.T, response map[string]interface{}) {
	memory, exists := response["memory"]
	if !exists {
		t.Error("Missing required field 'memory' in response")
		return
	}

	memoryMap, ok := memory.(map[string]interface{})
	if !ok {
		t.Error("Field 'memory' should be a map")
		return
	}

	// Test heap subsection
	if heap, exists := memoryMap["heap"]; !exists {
		t.Error("Missing 'heap' in memory section")
	} else if heapMap, ok := heap.(map[string]interface{}); !ok {
		t.Error("Field 'heap' should be a map")
	} else {
		heapFields := []string{"allocated_bytes", "in_use_bytes", "system_bytes", "objects_count", "allocated_mb"}
		for _, field := range heapFields {
			if value, exists := heapMap[field]; !exists {
				t.Errorf("Missing '%s' in heap section", field)
			} else if _, ok := value.(float64); !ok {
				t.Errorf("Field '%s' should be a number, got %T", field, value)
			}
		}
	}

	// Test gc subsection
	if gc, exists := memoryMap["gc"]; !exists {
		t.Error("Missing 'gc' in memory section")
	} else if gcMap, ok := gc.(map[string]interface{}); !ok {
		t.Error("Field 'gc' should be a map")
	} else {
		gcFields := []string{"cycles_total", "next_target_bytes", "pause_total_ns", "last_pause_ns", "cpu_percent"}
		for _, field := range gcFields {
			if value, exists := gcMap[field]; !exists {
				t.Errorf("Missing '%s' in gc section", field)
			} else if _, ok := value.(float64); !ok {
				t.Errorf("Field '%s' should be a number, got %T", field, value)
			}
		}
	}

	// Test stack subsection
	if stack, exists := memoryMap["stack"]; !exists {
		t.Error("Missing 'stack' in memory section")
	} else if stackMap, ok := stack.(map[string]interface{}); !ok {
		t.Error("Field 'stack' should be a map")
	} else {
		stackFields := []string{"in_use_bytes", "system_bytes"}
		for _, field := range stackFields {
			if value, exists := stackMap[field]; !exists {
				t.Errorf("Missing '%s' in stack section", field)
			} else if _, ok := value.(float64); !ok {
				t.Errorf("Field '%s' should be a number, got %T", field, value)
			}
		}
	}
}

// testRuntimeSection validates the runtime metrics section structure and values.
func testRuntimeSection(t *testing.T, response map[string]interface{}) {
	runtime, exists := response["runtime"]
	if !exists {
		t.Error("Missing required field 'runtime' in response")
		return
	}

	runtimeMap, ok := runtime.(map[string]interface{})
	if !ok {
		t.Error("Field 'runtime' should be a map")
		return
	}

	// Test goroutines subsection
	if goroutines, exists := runtimeMap["goroutines"]; !exists {
		t.Error("Missing 'goroutines' in runtime section")
	} else if goroutinesMap, ok := goroutines.(map[string]interface{}); !ok {
		t.Error("Field 'goroutines' should be a map")
	} else {
		if count, exists := goroutinesMap["count"]; !exists {
			t.Error("Missing 'count' in goroutines section")
		} else if countVal, ok := count.(float64); !ok || countVal <= 0 {
			t.Errorf("Field 'count' should be a positive number, got %v", count)
		}

		if note, exists := goroutinesMap["note"]; !exists || note != "includes_system_goroutines" {
			t.Errorf("Expected note 'includes_system_goroutines', got %v", note)
		}
	}

	// Test environment subsection
	if environment, exists := runtimeMap["environment"]; !exists {
		t.Error("Missing 'environment' in runtime section")
	} else if envMap, ok := environment.(map[string]interface{}); !ok {
		t.Error("Field 'environment' should be a map")
	} else {
		if goVersion, exists := envMap["go_version"]; !exists {
			t.Error("Missing 'go_version' in environment section")
		} else if _, ok := goVersion.(string); !ok {
			t.Error("Field 'go_version' should be a string")
		}

		if platform, exists := envMap["platform"]; !exists {
			t.Error("Missing 'platform' in environment section")
		} else if _, ok := platform.(string); !ok {
			t.Error("Field 'platform' should be a string")
		}

		if cpuCores, exists := envMap["cpu_cores"]; !exists {
			t.Error("Missing 'cpu_cores' in environment section")
		} else if cores, ok := cpuCores.(float64); !ok || cores <= 0 {
			t.Errorf("Field 'cpu_cores' should be a positive number, got %v", cpuCores)
		}

		if maxProcs, exists := envMap["max_procs"]; !exists {
			t.Error("Missing 'max_procs' in environment section")
		} else if procs, ok := maxProcs.(float64); !ok || procs <= 0 {
			t.Errorf("Field 'max_procs' should be a positive number, got %v", maxProcs)
		}
	}
}

// testApplicationSection validates the application metrics section structure.
func testApplicationSection(t *testing.T, response map[string]interface{}) {
	application, exists := response["application"]
	if !exists {
		t.Error("Missing required field 'application' in response")
		return
	}

	appMap, ok := application.(map[string]interface{})
	if !ok {
		t.Error("Field 'application' should be a map")
		return
	}

	// Test service subsection
	if service, exists := appMap["service"]; !exists {
		t.Error("Missing 'service' in application section")
	} else if serviceMap, ok := service.(map[string]interface{}); !ok {
		t.Error("Field 'service' should be a map")
	} else {
		serviceFields := map[string]string{
			"name":        "goedu-theta",
			"version":     "1.0.0",
			"environment": "development",
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

// testResourcesSection validates the resources metrics section structure.
func testResourcesSection(t *testing.T, response map[string]interface{}) {
	resources, exists := response["resources"]
	if !exists {
		t.Error("Missing required field 'resources' in response")
		return
	}

	resourcesMap, ok := resources.(map[string]interface{})
	if !ok {
		t.Error("Field 'resources' should be a map")
		return
	}

	// Test memory_pressure subsection
	if memoryPressure, exists := resourcesMap["memory_pressure"]; !exists {
		t.Error("Missing 'memory_pressure' in resources section")
	} else if pressureMap, ok := memoryPressure.(map[string]interface{}); !ok {
		t.Error("Field 'memory_pressure' should be a map")
	} else {
		if utilization, exists := pressureMap["heap_utilization_percent"]; !exists {
			t.Error("Missing 'heap_utilization_percent' in memory_pressure section")
		} else if utilizationVal, ok := utilization.(float64); !ok || utilizationVal < 0 || utilizationVal > 100 {
			t.Errorf("Field 'heap_utilization_percent' should be between 0-100, got %v", utilization)
		}
	}

	// Test performance subsection
	if performance, exists := resourcesMap["performance"]; !exists {
		t.Error("Missing 'performance' in resources section")
	} else if perfMap, ok := performance.(map[string]interface{}); !ok {
		t.Error("Field 'performance' should be a map")
	} else {
		if gcOverhead, exists := perfMap["gc_overhead_percent"]; !exists {
			t.Error("Missing 'gc_overhead_percent' in performance section")
		} else if _, ok := gcOverhead.(float64); !ok {
			t.Error("Field 'gc_overhead_percent' should be a number")
		}
	}
}

// TestHandleMetricsCacheHeaders tests that metrics endpoint sets proper cache headers.
//
// Metrics should never be cached as they represent real-time application state.
func TestHandleMetricsCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	h := handlers.NewHandler(logger)

	router := gin.New()
	router.GET("/metrics", h.HandleMetrics)

	req, err := http.NewRequest("GET", "/metrics", nil)
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

// TestHandleMetricsMemoryConsistency tests memory metric consistency.
//
// This verifies that the memory metrics are internally consistent and reasonable.
func TestHandleMetricsMemoryConsistency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	h := handlers.NewHandler(logger)

	router := gin.New()
	router.GET("/metrics", h.HandleMetrics)

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	memory := response["memory"].(map[string]interface{})
	heap := memory["heap"].(map[string]interface{})

	// Test MB conversion consistency
	if allocatedBytes, ok1 := heap["allocated_bytes"].(float64); ok1 {
		if allocatedMB, ok2 := heap["allocated_mb"].(float64); ok2 {
			expectedMB := allocatedBytes / 1024 / 1024
			// Allow for small rounding differences
			if allocatedMB < expectedMB-1 || allocatedMB > expectedMB+1 {
				t.Errorf("Allocated MB (%f) not consistent with allocated bytes (%f)", allocatedMB, allocatedBytes)
			}
		}
	}

	// Test that heap in use <= heap system
	if heapInUse, ok1 := heap["in_use_bytes"].(float64); ok1 {
		if heapSystem, ok2 := heap["system_bytes"].(float64); ok2 {
			if heapInUse > heapSystem {
				t.Errorf("Heap in use (%f) should not exceed heap system (%f)", heapInUse, heapSystem)
			}
		}
	}
}

// TestHandleMetricsCollectionTiming tests that collection timing is reasonable.
//
// Metrics collection should be fast to minimize impact on the application.
func TestHandleMetricsCollectionTiming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	h := handlers.NewHandler(logger)

	router := gin.New()
	router.GET("/metrics", h.HandleMetrics)

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Measure total response time
	start := time.Now()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	totalDuration := time.Since(start)

	// Metrics collection should be reasonably fast
	maxDuration := 100 * time.Millisecond
	if totalDuration > maxDuration {
		t.Errorf("Metrics collection took too long: %v (max: %v)", totalDuration, maxDuration)
	}

	// Parse response to check internal collection timing
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	metadata := response["metadata"].(map[string]interface{})
	collectionDurationNs := metadata["collection_duration_ns"].(float64)
	collectionDuration := time.Duration(collectionDurationNs) * time.Nanosecond

	// Internal collection should be very fast
	maxInternalDuration := 50 * time.Millisecond
	if collectionDuration > maxInternalDuration {
		t.Errorf("Internal collection took too long: %v (max: %v)", collectionDuration, maxInternalDuration)
	}

	// Collection duration should be positive
	if collectionDuration <= 0 {
		t.Error("Collection duration should be positive")
	}
}

// TestHandleMetricsWithMonitoringClients tests metrics endpoint with various monitoring clients.
//
// This verifies that the metrics endpoint works consistently for different monitoring systems.
func TestHandleMetricsWithMonitoringClients(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	h := handlers.NewHandler(logger)

	router := gin.New()
	router.GET("/metrics", h.HandleMetrics)

	testCases := []struct {
		name      string
		userAgent string
	}{
		{"Prometheus", "Prometheus/2.30.0"},
		{"Grafana", "Grafana/8.2.0"},
		{"DataDog", "datadog-agent/7.32.0"},
		{"Custom Monitor", "MyMonitor/1.0.0"},
		{"Load Balancer", "ELB-HealthChecker/2.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/metrics", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("User-Agent", tc.userAgent)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should always return 200 OK
			if w.Code != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}

			// Response should be valid JSON with metrics
			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
			}

			// Should have all major sections
			requiredSections := []string{"metadata", "memory", "runtime", "application", "resources"}
			for _, section := range requiredSections {
				if _, exists := response[section]; !exists {
					t.Errorf("Missing required section '%s' in response", section)
				}
			}
		})
	}
}
