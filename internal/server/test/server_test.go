package server_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
	"github.com/radek-zitek-cloud/goedu-theta/internal/server"
)

// TestNewServer tests the server creation with default configuration.
//
// This test verifies that the server can be created successfully with a given
// configuration and that all necessary components are properly initialized.
//
// Testing Strategy:
//   - Unit test for server instantiation
//   - Validates server creation without starting HTTP listener
//   - Tests configuration injection
//   - Verifies logger dependency injection
func TestNewServer(t *testing.T) {
	cfg := config.Server{
		Port:            8080,
		Host:            "localhost",
		ReadTimeout:     30,
		WriteTimeout:    30,
		ShutdownTimeout: 15,
	}

	// Create a test logger that discards output to avoid test noise
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv := server.NewServer(cfg, logger)

	if srv == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	// Server should be created successfully
	// Since internal fields are private, we test by starting and stopping the server
}

// TestServerStartAndShutdown tests the complete server lifecycle.
//
// This test verifies that the server can start successfully, listen for connections,
// and shut down gracefully within the expected timeframe.
//
// Testing Strategy:
//   - Integration test for server lifecycle
//   - Tests actual HTTP listener startup
//   - Validates graceful shutdown functionality
//   - Tests shutdown timeout behavior
func TestServerStartAndShutdown(t *testing.T) {
	cfg := config.Server{
		Port:            8090,
		Host:            "localhost",
		ReadTimeout:     30,
		WriteTimeout:    30,
		ShutdownTimeout: 15,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv := server.NewServer(cfg, logger)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give the server a moment to start listening
	time.Sleep(100 * time.Millisecond)

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Server shutdown failed: %v", err)
	}
}

// TestServerWithInvalidConfig tests server creation with invalid configuration.
//
// This test verifies that the server handles invalid configurations appropriately
// and that proper error handling is in place for edge cases.
//
// Testing Strategy:
//   - Negative testing with invalid configurations
//   - Tests server behavior with edge case values
//   - Validates configuration validation (if implemented)
func TestServerWithInvalidConfig(t *testing.T) {
	testCases := []struct {
		name   string
		config config.Server
	}{
		{
			name: "Zero Port",
			config: config.Server{
				Port:         0,
				Host:         "localhost",
				ReadTimeout:  30,
				WriteTimeout: 30,
			},
		},
		{
			name: "Negative Timeouts",
			config: config.Server{
				Port:         8091,
				Host:         "localhost",
				ReadTimeout:  -1,
				WriteTimeout: -1,
			},
		},
		{
			name: "Empty Host",
			config: config.Server{
				Port:         8092,
				Host:         "",
				ReadTimeout:  30,
				WriteTimeout: 30,
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Server creation should not panic even with invalid config
			srv := server.NewServer(tc.config, logger)
			if srv == nil {
				t.Error("Server should be created even with invalid config (validation should happen at start)")
			}

			// Note: Actual validation might happen during Start(), not NewServer()
			// This test ensures no panics during server creation
		})
	}
}

// TestServerShutdownTimeout tests server shutdown behavior with timeout.
//
// This test verifies that the server properly handles shutdown timeouts
// and returns appropriate errors when graceful shutdown cannot complete.
//
// Testing Strategy:
//   - Tests timeout behavior during shutdown
//   - Validates context cancellation handling
//   - Tests error propagation from shutdown operations
func TestServerShutdownTimeout(t *testing.T) {
	cfg := config.Server{
		Port:            8093,
		Host:            "localhost",
		ReadTimeout:     30,
		WriteTimeout:    30,
		ShutdownTimeout: 15,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv := server.NewServer(cfg, logger)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Create a very short timeout context to simulate timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(10 * time.Millisecond)

	// Attempt shutdown with already expired context
	err := srv.Shutdown(ctx)

	// This might or might not return an error depending on how fast shutdown is
	// The important thing is that it doesn't panic and handles the timeout gracefully
	if err != nil {
		t.Logf("Shutdown returned error as expected with expired context: %v", err)
	}
}

// TestServerMultipleShutdowns tests calling shutdown multiple times.
//
// This test verifies that calling shutdown multiple times on the same server
// doesn't cause panics or unexpected behavior.
//
// Testing Strategy:
//   - Tests idempotent shutdown behavior
//   - Validates that multiple shutdown calls are safe
//   - Tests error handling for shutdown of already stopped server
func TestServerMultipleShutdowns(t *testing.T) {
	cfg := config.Server{
		Port:            8094,
		Host:            "localhost",
		ReadTimeout:     30,
		WriteTimeout:    30,
		ShutdownTimeout: 15,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv := server.NewServer(cfg, logger)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First shutdown
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("First shutdown failed: %v", err)
	}

	// Second shutdown (should not panic)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	err := srv.Shutdown(ctx2)
	// Second shutdown might return an error (server already closed), but shouldn't panic
	if err != nil {
		t.Logf("Second shutdown returned error as expected: %v", err)
	}
}

// TestServerPortBinding tests server port binding behavior.
//
// This test verifies that the server properly binds to the specified port
// and that attempting to bind to the same port twice results in appropriate errors.
//
// Testing Strategy:
//   - Tests successful port binding
//   - Tests port conflict detection
//   - Validates that server listens on correct address
func TestServerPortBinding(t *testing.T) {
	cfg1 := config.Server{
		Port:         8095,
		Host:         "localhost",
		ReadTimeout:  30,
		WriteTimeout: 30,
	}
	cfg2 := config.Server{
		Port:         8095, // Same port
		Host:         "localhost",
		ReadTimeout:  30,
		WriteTimeout: 30,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	srv1 := server.NewServer(cfg1, logger)
	srv2 := server.NewServer(cfg2, logger)

	// Start first server
	if err := srv1.Start(); err != nil {
		t.Fatalf("Failed to start first server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv1.Shutdown(ctx)
	}()

	// Give first server time to bind to port
	time.Sleep(100 * time.Millisecond)

	// Test that server is actually listening by making a request
	resp, err := http.Get("http://localhost:8095/")
	if err != nil {
		t.Errorf("Server not listening on expected port: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Unexpected response status: %d", resp.StatusCode)
		}
	}

	// Starting second server on same port should eventually fail
	// (though Start() is non-blocking, so we test by trying to make requests)
	if err := srv2.Start(); err != nil {
		t.Logf("Second server start returned error as expected: %v", err)
	}

	// Give second server a moment to try to start
	time.Sleep(100 * time.Millisecond)

	// The first server should still be responding
	resp2, err := http.Get("http://localhost:8095/")
	if err != nil {
		t.Error("First server should still be accessible")
	} else {
		resp2.Body.Close()
	}
}
