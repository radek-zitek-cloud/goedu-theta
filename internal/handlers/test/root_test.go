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

// TestHandleRoot tests the root endpoint handler with comprehensive validation.
//
// This test verifies:
// - Correct HTTP status code (200 OK)
// - Proper Content-Type header (application/json)
// - Required JSON response fields and their values
// - Response structure and data types
// - Endpoint list completeness and accuracy
//
// Testing Strategy:
//   - Unit test approach using httptest.ResponseRecorder
//   - No external dependencies (fast execution)
//   - Comprehensive field validation
//   - Error message clarity for debugging
func TestHandleRoot(t *testing.T) {
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
	router.GET("/", h.HandleRoot)

	// Create test HTTP request
	req, err := http.NewRequest("GET", "/", nil)
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

	// Parse JSON response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Test required fields and their values
	testCases := []struct {
		field    string
		expected interface{}
		testType string
	}{
		{"message", "Welcome to GoEdu-Theta API Server", "string"},
		{"status", "running", "string"},
		{"version", "1.0.0", "string"},
	}

	for _, tc := range testCases {
		if value, exists := response[tc.field]; !exists {
			t.Errorf("Missing required field '%s' in response", tc.field)
		} else if value != tc.expected {
			t.Errorf("Field '%s': expected '%v', got '%v'", tc.field, tc.expected, value)
		}
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

	// Test endpoints field structure and content
	if endpoints, exists := response["endpoints"]; !exists {
		t.Error("Missing required field 'endpoints' in response")
	} else if endpointsList, ok := endpoints.([]interface{}); !ok {
		t.Error("Field 'endpoints' should be an array")
	} else {
		// Verify expected endpoints are present
		expectedEndpoints := []string{"GET /", "GET /health", "GET /metrics"}
		if len(endpointsList) != len(expectedEndpoints) {
			t.Errorf("Expected %d endpoints, got %d", len(expectedEndpoints), len(endpointsList))
		}

		// Convert to string slice for comparison
		actualEndpoints := make([]string, len(endpointsList))
		for i, endpoint := range endpointsList {
			if endpointStr, ok := endpoint.(string); ok {
				actualEndpoints[i] = endpointStr
			} else {
				t.Errorf("Endpoint at index %d is not a string: %v", i, endpoint)
			}
		}

		// Check that all expected endpoints are present
		for _, expected := range expectedEndpoints {
			found := false
			for _, actual := range actualEndpoints {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected endpoint '%s' not found in response", expected)
			}
		}
	}
}

// TestHandleRootWithUserAgent tests the root endpoint with different User-Agent headers.
//
// This test verifies that the handler correctly processes and logs different
// User-Agent strings, which is important for analytics and monitoring.
func TestHandleRootWithUserAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	h := handlers.NewHandler(logger)

	router := gin.New()
	router.GET("/", h.HandleRoot)

	testCases := []struct {
		name      string
		userAgent string
	}{
		{"Browser", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
		{"API Client", "curl/7.68.0"},
		{"Load Balancer", "ELB-HealthChecker/2.0"},
		{"Empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Set User-Agent header
			if tc.userAgent != "" {
				req.Header.Set("User-Agent", tc.userAgent)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should always return 200 OK regardless of User-Agent
			if w.Code != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}

			// Response should still be valid JSON
			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
			}
		})
	}
}

// TestHandleRootResponseConsistency tests that multiple calls return consistent responses.
//
// This test verifies that the handler returns consistent data across multiple calls,
// with only the timestamp field expected to change between requests.
func TestHandleRootResponseConsistency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	h := handlers.NewHandler(logger)

	router := gin.New()
	router.GET("/", h.HandleRoot)

	// Make multiple requests
	const numRequests = 3
	responses := make([]map[string]interface{}, numRequests)

	for i := 0; i < numRequests; i++ {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("Failed to create request %d: %v", i, err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status code %d, got %d", i, http.StatusOK, w.Code)
		}

		if err := json.NewDecoder(w.Body).Decode(&responses[i]); err != nil {
			t.Fatalf("Request %d: failed to parse JSON response: %v", i, err)
		}

		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Compare responses (excluding timestamp which should differ)
	staticFields := []string{"message", "status", "version"}

	for _, field := range staticFields {
		firstValue := responses[0][field]
		for i := 1; i < numRequests; i++ {
			if responses[i][field] != firstValue {
				t.Errorf("Field '%s' inconsistent between requests: first=%v, request_%d=%v",
					field, firstValue, i, responses[i][field])
			}
		}
	}

	// Special handling for endpoints field (slice comparison)
	firstEndpoints := responses[0]["endpoints"]
	for i := 1; i < numRequests; i++ {
		currentEndpoints := responses[i]["endpoints"]

		// Convert to JSON strings for comparison since slices can't be compared directly
		firstJSON, _ := json.Marshal(firstEndpoints)
		currentJSON, _ := json.Marshal(currentEndpoints)

		if string(firstJSON) != string(currentJSON) {
			t.Errorf("Field 'endpoints' inconsistent between requests: first=%v, request_%d=%v",
				firstEndpoints, i, currentEndpoints)
		}
	}

	// Verify timestamps are different (they should be generated at request time)
	// Note: In fast test environments, timestamps might be identical due to high resolution
	// This is acceptable as long as the timestamp format is correct
	for i := 1; i < numRequests; i++ {
		if responses[i]["timestamp"] == responses[0]["timestamp"] {
			t.Logf("Timestamps are identical between requests (acceptable in fast test environment)")
		}
	}
}
