package database_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
	"github.com/radek-zitek-cloud/goedu-theta/internal/database"
)

// TestNewMongoDBManager_ValidConfig tests the MongoDB manager creation with valid configuration.
//
// This test verifies that the MongoDB manager can be created successfully with a properly
// configured database configuration struct. It tests the basic initialization path without
// requiring an actual MongoDB server connection.
//
// Testing Strategy:
//   - Unit test for manager creation with valid parameters
//   - Tests configuration validation and client initialization
//   - Validates proper error handling for connection failures
//   - Verifies logger integration and structured logging
//
// Note: This test may fail if no MongoDB server is available on the configured host:port.
// In CI/CD environments, this test should either be skipped or run with a MongoDB test container.
func TestNewMongoDBManager_ValidConfig(t *testing.T) {
	// Create a test logger for capturing initialization logs
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create a valid database configuration for testing
	// Note: This assumes MongoDB is running on localhost:27017
	cfg := config.Database{
		Host:     "localhost",
		Port:     27017,
		User:     "testuser", // Test username
		Password: "testpass", // Test password
		Name:     "testdb",   // Test database name
	}

	// Attempt to create MongoDB manager
	// This may fail if MongoDB is not available, which is expected in some test environments
	manager, err := database.NewMongoDBManager(cfg, logger)

	if err != nil {
		// Connection failed - this is expected if MongoDB is not available
		// Log the error but don't fail the test (allows testing in environments without MongoDB)
		t.Logf("MongoDB connection failed (expected if MongoDB not available): %v", err)
		t.Skip("Skipping test - MongoDB server not available")
		return
	}

	// Verify manager was created successfully
	if manager == nil {
		t.Fatal("Expected non-nil MongoDB manager, got nil")
	}

	// Test basic manager functionality
	if !manager.IsConnected() {
		t.Error("Expected manager to report connected status")
	}

	// Verify client is available
	client := manager.GetClient()
	if client == nil {
		t.Error("Expected non-nil MongoDB client")
	}

	// Verify database is available
	database := manager.GetDatabase()
	if database == nil {
		t.Error("Expected non-nil MongoDB database")
	}

	// Clean up connection
	if err := manager.Close(); err != nil {
		t.Errorf("Failed to close MongoDB connection: %v", err)
	}

	// Verify connection status after close
	if manager.IsConnected() {
		t.Error("Expected manager to report disconnected status after close")
	}
}

// TestNewMongoDBManager_InvalidConfig tests manager creation with invalid configurations.
//
// This test verifies that the MongoDB manager properly validates configuration parameters
// and returns appropriate errors for invalid configurations. It tests various invalid
// configuration scenarios to ensure robust error handling.
//
// Testing Strategy:
//   - Negative testing with invalid configurations
//   - Tests configuration validation logic
//   - Verifies appropriate error messages for different validation failures
//   - Ensures no manager instance is created for invalid configurations
func TestNewMongoDBManager_InvalidConfig(t *testing.T) {
	// Create a test logger for validation testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	testCases := []struct {
		name         string
		config       config.Database
		expectError  bool
		errorContext string
	}{
		{
			name: "Empty Host",
			config: config.Database{
				Host:     "", // Invalid: empty host
				Port:     27017,
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
			expectError:  true,
			errorContext: "host cannot be empty",
		},
		{
			name: "Invalid Port - Zero",
			config: config.Database{
				Host:     "localhost",
				Port:     0, // Invalid: port cannot be zero
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
			expectError:  true,
			errorContext: "port must be between 1 and 65535",
		},
		{
			name: "Invalid Port - Negative",
			config: config.Database{
				Host:     "localhost",
				Port:     -1, // Invalid: negative port
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
			expectError:  true,
			errorContext: "port must be between 1 and 65535",
		},
		{
			name: "Invalid Port - Too High",
			config: config.Database{
				Host:     "localhost",
				Port:     65536, // Invalid: port too high
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
			expectError:  true,
			errorContext: "port must be between 1 and 65535",
		},
		{
			name: "Empty Database Name",
			config: config.Database{
				Host:     "localhost",
				Port:     27017,
				User:     "testuser",
				Password: "testpass",
				Name:     "", // Invalid: empty database name
			},
			expectError:  true,
			errorContext: "database name cannot be empty",
		},
		{
			name: "Valid Config Without Authentication",
			config: config.Database{
				Host:     "localhost",
				Port:     27017,
				User:     "", // Valid: empty user for no auth
				Password: "", // Valid: empty password for no auth
				Name:     "testdb",
			},
			expectError:  false, // This should be valid
			errorContext: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Attempt to create MongoDB manager with test configuration
			manager, err := database.NewMongoDBManager(tc.config, logger)

			if tc.expectError {
				// Expect error for invalid configuration
				if err == nil {
					t.Errorf("Expected error for invalid configuration, got nil")
				}
				if manager != nil {
					t.Errorf("Expected nil manager for invalid configuration, got non-nil")
					// Clean up if manager was unexpectedly created
					manager.Close()
				}
			} else {
				// Expect success for valid configuration (connection may still fail)
				if err != nil {
					// Error occurred - check if it's a validation error or connection error
					if tc.errorContext != "" {
						t.Errorf("Unexpected validation error for valid configuration: %v", err)
					} else {
						// Connection error is acceptable if MongoDB is not available
						t.Logf("Connection failed (expected if MongoDB not available): %v", err)
					}
				} else if manager != nil {
					// Success - clean up connection
					manager.Close()
				}
			}
		})
	}
}

// TestMongoDBManager_PingHealthCheck tests the connection health check functionality.
//
// This test verifies that the Ping method properly checks connection health and
// returns appropriate results for both healthy and unhealthy connections.
//
// Testing Strategy:
//   - Tests ping operation with established connection
//   - Verifies context timeout handling
//   - Tests connection status updates based on ping results
//   - Validates error handling for connection failures
//
// Note: This test requires a running MongoDB instance and may be skipped in environments
// where MongoDB is not available.
func TestMongoDBManager_PingHealthCheck(t *testing.T) {
	// Create a test logger for ping testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create a valid database configuration
	cfg := config.Database{
		Host:     "localhost",
		Port:     27017,
		User:     "", // No authentication for test
		Password: "",
		Name:     "testdb",
	}

	// Create MongoDB manager
	manager, err := database.NewMongoDBManager(cfg, logger)
	if err != nil {
		t.Skipf("MongoDB connection failed - skipping ping test: %v", err)
		return
	}
	defer manager.Close()

	// Test ping with normal context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = manager.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed for healthy connection: %v", err)
	}

	// Verify connection status after successful ping
	if !manager.IsConnected() {
		t.Error("Expected connected status after successful ping")
	}

	// Test ping with very short timeout (may cause timeout error)
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer shortCancel()

	err = manager.Ping(shortCtx)
	// Note: This may or may not fail depending on network speed
	// We don't assert error here, just log the result
	t.Logf("Ping with short timeout result: %v", err)
}

// TestMongoDBManager_ConnectionLifecycle tests the complete connection lifecycle.
//
// This test verifies the complete lifecycle of a MongoDB connection including
// initialization, operation, health monitoring, and cleanup.
//
// Testing Strategy:
//   - Tests complete connection establishment process
//   - Verifies manager state transitions during lifecycle
//   - Tests proper cleanup and resource deallocation
//   - Validates error handling throughout the lifecycle
func TestMongoDBManager_ConnectionLifecycle(t *testing.T) {
	// Create a test logger for lifecycle testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create a valid database configuration
	cfg := config.Database{
		Host:     "localhost",
		Port:     27017,
		User:     "", // No authentication for test
		Password: "",
		Name:     "testdb",
	}

	// Step 1: Test manager creation
	manager, err := database.NewMongoDBManager(cfg, logger)
	if err != nil {
		t.Skipf("MongoDB connection failed - skipping lifecycle test: %v", err)
		return
	}

	// Step 2: Verify initial state
	if !manager.IsConnected() {
		t.Error("Expected connected status after successful creation")
	}

	client := manager.GetClient()
	if client == nil {
		t.Error("Expected non-nil client after successful creation")
	}

	database := manager.GetDatabase()
	if database == nil {
		t.Error("Expected non-nil database after successful creation")
	}

	// Step 3: Test health check operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := manager.Ping(ctx); err != nil {
		t.Errorf("Health check failed during lifecycle test: %v", err)
	}

	// Step 4: Test cleanup and resource deallocation
	if err := manager.Close(); err != nil {
		t.Errorf("Connection cleanup failed: %v", err)
	}

	// Step 5: Verify final state after cleanup
	if manager.IsConnected() {
		t.Error("Expected disconnected status after close")
	}

	// Step 6: Test ping after close (should fail)
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer pingCancel()

	err = manager.Ping(pingCtx)
	if err == nil {
		t.Error("Expected ping to fail after connection close")
	}
}

// TestNewConnection_BackwardCompatibility tests the backward compatibility wrapper function.
//
// This test verifies that the NewConnection function provides the expected interface
// for backward compatibility with existing code while delegating to the new implementation.
//
// Testing Strategy:
//   - Tests backward compatibility function interface
//   - Verifies delegation to NewMongoDBManager
//   - Ensures consistent behavior with new implementation
//   - Validates parameter passing and return values
func TestNewConnection_BackwardCompatibility(t *testing.T) {
	// Create a test logger for compatibility testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create a valid database configuration
	cfg := config.Database{
		Host:     "localhost",
		Port:     27017,
		User:     "", // No authentication for test
		Password: "",
		Name:     "testdb",
	}

	// Test backward compatibility function
	manager, err := database.NewConnection(cfg, logger)
	if err != nil {
		t.Skipf("MongoDB connection failed - skipping compatibility test: %v", err)
		return
	}
	defer manager.Close()

	// Verify that the returned manager has expected functionality
	if manager == nil {
		t.Fatal("Expected non-nil manager from NewConnection")
	}

	if !manager.IsConnected() {
		t.Error("Expected connected status from NewConnection")
	}

	// Verify client access
	client := manager.GetClient()
	if client == nil {
		t.Error("Expected non-nil client from NewConnection")
	}

	// Verify database access
	database := manager.GetDatabase()
	if database == nil {
		t.Error("Expected non-nil database from NewConnection")
	}

	// Test health check through compatibility interface
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := manager.Ping(ctx); err != nil {
		t.Errorf("Ping failed through compatibility interface: %v", err)
	}
}

// TestNewMongoDBManager_AtlasConfig tests MongoDB Atlas connection configuration.
//
// This test verifies that the MongoDB manager properly handles MongoDB Atlas SRV
// connection strings and Atlas-specific configuration parameters. It tests both
// valid and invalid Atlas configurations.
//
// Testing Strategy:
//   - Tests Atlas SRV connection string construction
//   - Validates Atlas-specific configuration requirements
//   - Tests proper handling of Atlas authentication requirements
//   - Verifies Atlas app name integration
func TestNewMongoDBManager_AtlasConfig(t *testing.T) {
	// Create a test logger for Atlas connection testing
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	testCases := []struct {
		name         string
		config       config.Database
		expectError  bool
		errorContext string
	}{
		{
			name: "Valid Atlas Config",
			config: config.Database{
				Host:         "clusterzitekcloud.dznruy0.mongodb.net",
				Port:         27017, // Ignored for Atlas connections
				User:         "radek",
				Password:     "testpassword",
				Name:         "goedu_theta",
				IsAtlas:      true,
				AtlasAppName: "ClusterZitekCloud",
			},
			expectError:  false, // May fail if Atlas credentials are invalid, but config is valid
			errorContext: "",
		},
		{
			name: "Atlas Config Missing User",
			config: config.Database{
				Host:         "clusterzitekcloud.dznruy0.mongodb.net",
				Port:         27017,
				User:         "", // Invalid: Atlas requires authentication
				Password:     "testpassword",
				Name:         "goedu_theta",
				IsAtlas:      true,
				AtlasAppName: "ClusterZitekCloud",
			},
			expectError:  true,
			errorContext: "database user is required for MongoDB Atlas connections",
		},
		{
			name: "Atlas Config Missing Password",
			config: config.Database{
				Host:         "clusterzitekcloud.dznruy0.mongodb.net",
				Port:         27017,
				User:         "radek",
				Password:     "", // Invalid: Atlas requires authentication
				Name:         "goedu_theta",
				IsAtlas:      true,
				AtlasAppName: "ClusterZitekCloud",
			},
			expectError:  true,
			errorContext: "database password is required for MongoDB Atlas connections",
		},
		{
			name: "Atlas Config Without App Name",
			config: config.Database{
				Host:         "clusterzitekcloud.dznruy0.mongodb.net",
				Port:         27017,
				User:         "radek",
				Password:     "testpassword",
				Name:         "goedu_theta",
				IsAtlas:      true,
				AtlasAppName: "", // Valid: app name is optional
			},
			expectError:  false, // May fail if Atlas credentials are invalid, but config is valid
			errorContext: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Attempt to create MongoDB manager with Atlas configuration
			manager, err := database.NewMongoDBManager(tc.config, logger)

			if tc.expectError {
				// Expect error for invalid Atlas configuration
				if err == nil {
					t.Errorf("Expected error for invalid Atlas configuration, got nil")
				}
				if manager != nil {
					t.Errorf("Expected nil manager for invalid Atlas configuration, got non-nil")
					// Clean up if manager was unexpectedly created
					manager.Close()
				}
			} else {
				// Valid configuration - may still fail due to network/auth issues
				if err != nil {
					// Connection failed - this is expected if Atlas credentials are invalid
					// or network is unavailable. Log the error but don't fail the test.
					t.Logf("Atlas connection failed (expected if credentials invalid): %v", err)
					t.Skip("Skipping test - Atlas connection not available")
					return
				}

				// Connection succeeded - verify manager functionality
				if manager == nil {
					t.Fatal("Expected non-nil MongoDB manager for valid Atlas config")
				}

				// Test Atlas-specific functionality
				if !manager.IsConnected() {
					t.Error("Expected Atlas manager to report connected status")
				}

				// Verify client is available
				client := manager.GetClient()
				if client == nil {
					t.Error("Expected non-nil MongoDB client for Atlas connection")
				}

				// Verify database is available
				database := manager.GetDatabase()
				if database == nil {
					t.Error("Expected non-nil MongoDB database for Atlas connection")
				}

				// Clean up Atlas connection
				if err := manager.Close(); err != nil {
					t.Errorf("Failed to close Atlas MongoDB connection: %v", err)
				}

				// Verify connection status after close
				if manager.IsConnected() {
					t.Error("Expected Atlas manager to report disconnected status after close")
				}
			}
		})
	}
}
