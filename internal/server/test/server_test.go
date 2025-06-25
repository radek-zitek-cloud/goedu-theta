package server_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
	"github.com/radek-zitek-cloud/goedu-theta/internal/server"
)

// TestNewServer tests the server creation with default configuration.
func TestNewServer(t *testing.T) {
	cfg := config.Server{
		Port:            8080,
		Host:            "localhost",
		ReadTimeout:     30,
		WriteTimeout:    30,
		ShutdownTimeout: 15,
	}

	// Create a test logger that discards output
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv := server.NewServer(cfg, logger)

	if srv == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	// Since config and router are private, we can only test the server creation
	// and verify it works by testing actual HTTP requests
}

// TestHandleRoot tests the root endpoint handler.
func TestHandleRoot(t *testing.T) {
	cfg := config.Server{Port: 8082, Host: "localhost", ReadTimeout: 30, WriteTimeout: 30}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv := server.NewServer(cfg, logger)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Make HTTP request to the server
	resp, err := http.Get("http://localhost:8082/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check response content type
	expectedContentType := "application/json; charset=utf-8"
	if contentType := resp.Header.Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, contentType)
	}

	// Parse and validate JSON response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	// Check required fields
	if message, exists := response["message"]; !exists || message != "Welcome to GoEdu-Theta API Server" {
		t.Errorf("Expected welcome message, got %v", message)
	}

	if status, exists := response["status"]; !exists || status != "running" {
		t.Errorf("Expected status 'running', got %v", status)
	}

	if endpoints, exists := response["endpoints"]; !exists {
		t.Error("Expected endpoints field in response")
	} else if endpointsList, ok := endpoints.([]interface{}); !ok || len(endpointsList) != 3 {
		t.Errorf("Expected 3 endpoints, got %v", endpoints)
	}
}

// TestHandleHealth tests the health check endpoint handler.
func TestHandleHealth(t *testing.T) {
	cfg := config.Server{Port: 8083, Host: "localhost", ReadTimeout: 30, WriteTimeout: 30}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv := server.NewServer(cfg, logger)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Make HTTP request to the server
	resp, err := http.Get("http://localhost:8083/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Parse and validate JSON response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	// Check required fields
	if status, exists := response["status"]; !exists || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", status)
	}

	if checks, exists := response["checks"]; !exists {
		t.Error("Expected checks field in response")
	} else if checksMap, ok := checks.(map[string]interface{}); !ok {
		t.Error("Expected checks to be a map")
	} else {
		expectedChecks := []string{"database", "memory", "disk"}
		for _, check := range expectedChecks {
			if value, exists := checksMap[check]; !exists || value != "ok" {
				t.Errorf("Expected check %s to be 'ok', got %v", check, value)
			}
		}
	}
}

// TestHandleMetrics tests the metrics endpoint handler.
func TestHandleMetrics(t *testing.T) {
	cfg := config.Server{Port: 8084, Host: "localhost", ReadTimeout: 30, WriteTimeout: 30}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv := server.NewServer(cfg, logger)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Make HTTP request to the server
	resp, err := http.Get("http://localhost:8084/metrics")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Parse and validate JSON response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	// Check required fields
	if status, exists := response["status"]; !exists || status != "ok" {
		t.Errorf("Expected status 'ok', got %v", status)
	}

	if metrics, exists := response["metrics"]; !exists {
		t.Error("Expected metrics field in response")
	} else if metricsMap, ok := metrics.(map[string]interface{}); !ok {
		t.Error("Expected metrics to be a map")
	} else {
		expectedMetrics := []string{"http_requests_total", "http_request_duration", "active_connections", "memory_usage_bytes", "goroutines_count"}
		for _, metric := range expectedMetrics {
			if _, exists := metricsMap[metric]; !exists {
				t.Errorf("Expected metric %s to exist", metric)
			}
		}
	}

	if buildInfo, exists := response["build_info"]; !exists {
		t.Error("Expected build_info field in response")
	} else if buildInfoMap, ok := buildInfo.(map[string]interface{}); !ok {
		t.Error("Expected build_info to be a map")
	} else {
		expectedBuildInfo := []string{"version", "commit", "built_at"}
		for _, info := range expectedBuildInfo {
			if _, exists := buildInfoMap[info]; !exists {
				t.Errorf("Expected build_info %s to exist", info)
			}
		}
	}
}

// TestServerShutdown tests graceful server shutdown.
func TestServerShutdown(t *testing.T) {
	cfg := config.Server{Port: 8085, Host: "localhost", ReadTimeout: 30, WriteTimeout: 30}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv := server.NewServer(cfg, logger)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Server shutdown failed: %v", err)
	}
}

// TestInvalidRoute tests handling of non-existent routes.
func TestInvalidRoute(t *testing.T) {
	cfg := config.Server{Port: 8086, Host: "localhost", ReadTimeout: 30, WriteTimeout: 30}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv := server.NewServer(cfg, logger)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Make HTTP request to the server
	resp, err := http.Get("http://localhost:8086/nonexistent")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check that we get a 404 status code
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d for invalid route, got %d", http.StatusNotFound, resp.StatusCode)
	}
}
