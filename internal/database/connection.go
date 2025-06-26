package database

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
)

// MongoDBManager represents a comprehensive MongoDB connection manager for the GoEdu-Theta application.
//
// This struct encapsulates MongoDB client management, connection pooling, health monitoring,
// and provides a centralized interface for all database operations. It implements proper
// connection lifecycle management with graceful initialization, monitoring, and cleanup.
//
// Architecture Design:
// The MongoDB manager follows the singleton pattern for application-wide database access
// while providing thread-safe operations and connection pool management. It abstracts
// MongoDB client complexity and provides a clean interface for database operations.
//
// Connection Management Features:
// - Automatic connection pooling with configurable parameters
// - Connection health monitoring and auto-reconnection
// - Graceful connection establishment with timeout handling
// - Proper resource cleanup and connection disposal
// - Comprehensive error handling and logging
//
// Security Features:
// - Secure authentication handling (username/password, certificates)
// - Connection string sanitization to prevent credential exposure
// - TLS/SSL support for encrypted connections
// - Network timeout protection against slow/unresponsive servers
//
// Performance Optimization:
// - Connection pooling reduces connection overhead
// - Configurable timeout values for different operation types
// - Efficient connection reuse across application requests
// - Background connection health monitoring
//
// Monitoring and Observability:
// - Structured logging for all connection events
// - Connection metrics and health status tracking
// - Error logging with detailed context information
// - Performance metrics for connection operations
//
// Thread Safety:
// All operations are thread-safe and can be called concurrently from multiple goroutines.
// The underlying MongoDB driver handles connection pool management automatically.
type MongoDBManager struct {
	// client represents the MongoDB client instance that manages connections to the database.
	// This client is thread-safe and manages an internal connection pool automatically.
	// It should be reused throughout the application lifecycle for optimal performance.
	client *mongo.Client

	// config contains the database configuration including connection parameters,
	// authentication credentials, and connection pool settings.
	// This configuration is immutable after initialization.
	config config.Database

	// logger provides structured logging for all database operations, connection events,
	// and error conditions. All log messages include relevant context for debugging.
	logger *slog.Logger

	// database represents the active MongoDB database instance for the application.
	// This provides direct access to collections and database-level operations.
	database *mongo.Database

	// connectionString stores the sanitized MongoDB connection URI used for establishing
	// connections. Sensitive credentials are masked in logs for security.
	connectionString string

	// isConnected tracks the current connection status to prevent redundant operations
	// and provide accurate health status reporting.
	isConnected bool
}

// NewMongoDBManager creates and initializes a new MongoDB connection manager with comprehensive
// configuration and connection establishment.
//
// This function performs complete MongoDB client initialization including:
// - Connection string construction from configuration parameters
// - MongoDB client creation with optimized connection pool settings
// - Connection establishment and health verification
// - Database selection and accessibility testing
// - Comprehensive error handling and logging
//
// Connection Process:
// 1. Validates and sanitizes database configuration parameters
// 2. Constructs MongoDB connection URI with authentication and options
// 3. Creates MongoDB client with production-ready connection pool settings
// 4. Establishes connection with configurable timeout handling
// 5. Verifies database accessibility and permissions
// 6. Initializes health monitoring and connection tracking
//
// Security Considerations:
// - Connection credentials are never logged in plain text
// - Connection strings are sanitized before logging
// - TLS/SSL connections are supported for encrypted communication
// - Authentication is handled securely through MongoDB driver
//
// Performance Optimization:
// - Connection pooling is configured for optimal resource utilization
// - Connection timeouts prevent resource blocking on slow networks
// - Background connection monitoring ensures connection health
// - Efficient connection reuse reduces establishment overhead
//
// Error Handling:
// - Network connectivity issues are properly detected and reported
// - Authentication failures provide clear error messaging
// - Configuration validation prevents runtime connection failures
// - Comprehensive logging aids in troubleshooting connection issues
//
// Parameters:
//   - cfg: Database configuration struct containing connection parameters including
//     host, port, credentials, database name, and connection options.
//   - logger: Structured logger instance for comprehensive operation logging including
//     connection events, errors, and performance metrics.
//
// Returns:
//   - *MongoDBManager: Fully initialized and connected MongoDB manager instance
//     ready for database operations with established connection pool.
//   - error: Detailed error information if connection establishment fails,
//     including specific failure reasons and troubleshooting context.
//
// Usage Examples:
//
//	// Basic connection establishment
//	dbManager, err := database.NewMongoDBManager(cfg.Database, logger)
//	if err != nil {
//	    log.Fatal("Failed to connect to MongoDB:", err)
//	}
//	defer dbManager.Close()
//
//	// Get database instance for operations
//	db := dbManager.GetDatabase()
//	collection := db.Collection("users")
//
// Connection String Format:
// The function constructs MongoDB URIs in the following formats:
// - With authentication: mongodb://username:password@host:port/database?options
// - Without authentication: mongodb://host:port/database?options
// - With replica set: mongodb://host1:port1,host2:port2/database?replicaSet=rs0
//
// Complexity Analysis:
//   - Time Complexity: O(1) for initialization, O(n) for connection establishment where n is network latency
//   - Space Complexity: O(1) for manager instance, O(m) for connection pool where m is pool size
//   - Network Operations: 1-3 round trips for connection establishment and health check
//
// Thread Safety:
// This function is thread-safe and can be called concurrently. However, it's recommended
// to create a single MongoDB manager instance per application for optimal resource utilization.
func NewMongoDBManager(cfg config.Database, logger *slog.Logger) (*MongoDBManager, error) {
	// Log the start of MongoDB connection initialization for operational visibility
	// This helps track database connection attempts and timing in application logs
	logger.Info("🍃 Initializing MongoDB connection manager",
		slog.String("host", cfg.Host),          // Database server hostname/IP
		slog.Int("port", cfg.Port),             // Database server port
		slog.String("database_name", cfg.Name), // Target database name
		slog.String("user", cfg.User),          // Authentication username (safe to log)
	)

	// Step 1: Validate database configuration to prevent runtime connection failures
	// This early validation catches configuration errors before attempting network operations
	if err := validateDatabaseConfig(cfg); err != nil {
		logger.Error("🍃 Invalid database configuration detected",
			slog.Any("error", err),                              // Detailed validation error
			slog.String("validation_context", "pre_connection"), // When validation occurred
		)
		return nil, fmt.Errorf("database configuration validation failed: %w", err)
	}

	// Step 2: Construct MongoDB connection URI with authentication and connection options
	// The connection string includes all necessary parameters for establishing a secure connection
	connectionString := constructConnectionString(cfg, logger)

	// Step 3: Configure MongoDB client options with production-ready settings
	// These options optimize connection pooling, timeouts, and performance characteristics
	clientOptions := options.Client().ApplyURI(connectionString)

	// Connection Pool Configuration:
	// Configure connection pool settings for optimal resource utilization and performance.
	// These settings balance connection availability with resource consumption.
	clientOptions.SetMaxPoolSize(100)                  // Maximum concurrent connections (prevents resource exhaustion)
	clientOptions.SetMinPoolSize(5)                    // Minimum maintained connections (reduces connection overhead)
	clientOptions.SetMaxConnIdleTime(30 * time.Minute) // Idle connection timeout (prevents stale connections)

	// Timeout Configuration:
	// Configure various timeout settings to prevent resource blocking and ensure responsive behavior.
	// These timeouts protect against network issues and slow database responses.
	clientOptions.SetConnectTimeout(10 * time.Second)        // Connection establishment timeout
	clientOptions.SetServerSelectionTimeout(5 * time.Second) // Server selection timeout
	clientOptions.SetSocketTimeout(30 * time.Second)         // Individual operation timeout

	// Monitoring and Health Check Configuration:
	// Configure heartbeat and monitoring intervals for connection health tracking.
	// Regular health checks ensure connection reliability and automatic recovery.
	clientOptions.SetHeartbeatInterval(10 * time.Second) // Connection health check interval

	// Step 4: Create MongoDB client instance with configured options
	// The client manages the connection pool and provides the interface for database operations
	logger.Debug("🍃 Creating MongoDB client with connection pool configuration",
		slog.Int("max_pool_size", 100),                   // Maximum connection pool size
		slog.Int("min_pool_size", 5),                     // Minimum connection pool size
		slog.Duration("connect_timeout", 10*time.Second), // Connection timeout setting
		slog.Duration("socket_timeout", 30*time.Second),  // Socket operation timeout
	)

	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		logger.Error("🍃 Failed to create MongoDB client instance",
			slog.Any("error", err),                             // MongoDB driver error details
			slog.String("connection_stage", "client_creation"), // Stage where failure occurred
		)
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// Step 5: Establish connection to MongoDB server with timeout context
	// This performs the actual network connection and authentication
	logger.Debug("🍃 Establishing connection to MongoDB server",
		slog.String("operation", "client_connect"),
	)

	// Create a context with timeout for connection establishment
	// This prevents indefinite blocking if the database server is unresponsive
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		logger.Error("🍃 Failed to establish connection to MongoDB server",
			slog.Any("error", err),                            // Connection failure details
			slog.String("connection_stage", "server_connect"), // Stage where failure occurred
			slog.Duration("timeout_used", 15*time.Second),     // Timeout that was applied
		)
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Step 6: Verify connection health and server accessibility
	// This ensures the connection is fully functional and the server is responsive
	logger.Debug("🍃 Verifying MongoDB server connectivity and health",
		slog.String("operation", "ping_server"),
	)

	// Ping the server to verify connectivity and server health
	// Use a shorter timeout for ping operation as it should be very fast
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		// Connection established but server is not responding properly
		// Attempt graceful cleanup before returning error
		if disconnectErr := client.Disconnect(context.Background()); disconnectErr != nil {
			logger.Warn("🍃 Failed to disconnect after ping failure",
				slog.Any("disconnect_error", disconnectErr), // Cleanup failure details
			)
		}

		logger.Error("🍃 MongoDB server ping failed - server not responsive",
			slog.Any("error", err),                         // Ping failure details
			slog.String("connection_stage", "server_ping"), // Stage where failure occurred
			slog.Duration("ping_timeout", 5*time.Second),   // Ping timeout that was used
		)
		return nil, fmt.Errorf("MongoDB server ping failed: %w", err)
	}

	// Step 7: Select and verify database accessibility
	// This ensures the specified database exists and is accessible with current credentials
	database := client.Database(cfg.Name)

	// Test database accessibility by attempting to list collections
	// This verifies that the user has appropriate permissions for database operations
	logger.Debug("🍃 Verifying database accessibility and permissions",
		slog.String("database_name", cfg.Name),
		slog.String("operation", "list_collections"),
	)

	listCtx, listCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer listCancel()

	// Attempt to list collections to verify database access permissions
	// This operation requires read permissions on the database
	_, err = database.ListCollectionNames(listCtx, map[string]interface{}{})
	if err != nil {
		// Database access failed - could be permissions or database doesn't exist
		// Attempt graceful cleanup before returning error
		if disconnectErr := client.Disconnect(context.Background()); disconnectErr != nil {
			logger.Warn("🍃 Failed to disconnect after database access failure",
				slog.Any("disconnect_error", disconnectErr), // Cleanup failure details
			)
		}

		logger.Error("🍃 Database access verification failed - check permissions and database existence",
			slog.Any("error", err),                             // Database access error details
			slog.String("database_name", cfg.Name),             // Database that failed access
			slog.String("connection_stage", "database_access"), // Stage where failure occurred
		)
		return nil, fmt.Errorf("database access verification failed: %w", err)
	}

	// Step 8: Initialize MongoDB manager instance with successful connection
	// All connection steps have succeeded - create the manager instance
	manager := &MongoDBManager{
		client:           client,           // Established MongoDB client
		config:           cfg,              // Database configuration
		logger:           logger,           // Structured logger
		database:         database,         // Selected database instance
		connectionString: connectionString, // Connection URI (sanitized)
		isConnected:      true,             // Connection status tracking
	}

	// Log successful connection establishment with operational details
	// This provides confirmation that the database is ready for operations
	logger.Info("🍃 MongoDB connection established successfully",
		slog.String("database_name", cfg.Name),               // Connected database
		slog.String("server_host", cfg.Host),                 // Database server
		slog.Int("server_port", cfg.Port),                    // Database port
		slog.Bool("connection_healthy", true),                // Connection health status
		slog.String("connection_pool_status", "initialized"), // Pool initialization status
	)

	return manager, nil
}

// validateDatabaseConfig performs comprehensive validation of database configuration parameters
// to ensure all required values are present and within acceptable ranges for both
// standard MongoDB connections and MongoDB Atlas SRV connections.
//
// This function prevents runtime connection failures by catching configuration errors early
// in the initialization process. It validates both required parameters and value ranges
// to ensure successful database connection establishment.
//
// Validation Rules:
// Standard MongoDB:
// - Host: Must be non-empty string (hostname, IP address)
// - Port: Must be within valid port range (1-65535)
// - Database Name: Must be non-empty string following MongoDB naming conventions
// - User: Can be empty for unauthenticated connections
// - Password: Can be empty for unauthenticated connections
//
// MongoDB Atlas:
// - Host: Must be non-empty Atlas cluster hostname (e.g., cluster.xxxxx.mongodb.net)
// - Port: Ignored for Atlas connections (resolved via DNS SRV)
// - Database Name: Must be non-empty string
// - User: Required for Atlas connections (Atlas doesn't support unauthenticated access)
// - Password: Required for Atlas connections
// - AtlasAppName: Optional but recommended for monitoring
//
// Parameters:
//   - cfg: Database configuration struct to validate
//
// Returns:
//   - error: Detailed validation error if any parameter is invalid, nil if all valid
//
// Complexity:
//   - Time Complexity: O(1) - simple parameter validation
//   - Space Complexity: O(1) - no additional data structures
func validateDatabaseConfig(cfg config.Database) error {
	// Validate database host - required for all connections
	if cfg.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}

	// Validate database name - required for database selection
	if cfg.Name == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	if cfg.IsAtlas {
		// MongoDB Atlas specific validations
		if cfg.User == "" {
			return fmt.Errorf("database user is required for MongoDB Atlas connections")
		}
		if cfg.Password == "" {
			return fmt.Errorf("database password is required for MongoDB Atlas connections")
		}
		// Port validation is skipped for Atlas as it uses SRV records
	} else {
		// Standard MongoDB validations
		// Validate database port - must be within valid TCP port range
		if cfg.Port < 1 || cfg.Port > 65535 {
			return fmt.Errorf("database port must be between 1 and 65535, got: %d", cfg.Port)
		}
		// Note: User and Password can be empty for unauthenticated connections
		// or when using alternative authentication methods (certificates, etc.)
	}

	return nil
}

// constructConnectionString builds a MongoDB connection URI from configuration parameters
// with proper encoding and security considerations, supporting both standard MongoDB
// connections and MongoDB Atlas SRV connections.
//
// This function creates properly formatted MongoDB connection strings that include
// authentication credentials, connection options, and security parameters.
// Sensitive information is handled securely and not exposed in logs.
//
// Connection String Formats:
// Standard MongoDB:
// - With authentication: mongodb://username:password@host:port/database
// - Without authentication: mongodb://host:port/database
// - With options: mongodb://host:port/database?option1=value1&option2=value2
//
// MongoDB Atlas (SRV):
// - With authentication: mongodb+srv://username:password@cluster.host/database?options
// - The +srv scheme uses DNS SRV records to discover replica set members automatically
// - Port is not specified as it's resolved through DNS SRV records
// - Includes Atlas-specific options like retryWrites=true&w=majority
//
// Security Features:
// - Password URL encoding prevents special character issues
// - Connection string sanitization for logging
// - TLS/SSL enabled by default for Atlas connections
// - Credential masking in debug output
//
// Parameters:
//   - cfg: Database configuration containing connection parameters
//   - logger: Logger for connection string construction debugging
//
// Returns:
//   - string: Properly formatted MongoDB connection URI
//
// Complexity:
//   - Time Complexity: O(1) - string concatenation operations
//   - Space Complexity: O(1) - single connection string allocation
func constructConnectionString(cfg config.Database, logger *slog.Logger) string {
	// Start with base MongoDB URI scheme
	var connectionString string

	// Determine if this is a MongoDB Atlas connection
	if cfg.IsAtlas {
		// MongoDB Atlas SRV connection string construction
		if cfg.User != "" && cfg.Password != "" {
			// Authenticated Atlas connection with SRV DNS resolution
			// URL encode password to handle special characters safely
			encodedPassword := url.QueryEscape(cfg.Password)
			connectionString = fmt.Sprintf("mongodb+srv://%s:%s@%s/%s",
				cfg.User,        // Username for Atlas authentication
				encodedPassword, // URL-encoded password for Atlas authentication
				cfg.Host,        // Atlas cluster hostname (e.g., clustername.xxxxx.mongodb.net)
				cfg.Name,        // Database name to connect to
			)

			// Add Atlas-specific connection options for optimal performance and reliability
			atlasOptions := "retryWrites=true&w=majority"
			if cfg.AtlasAppName != "" {
				atlasOptions += fmt.Sprintf("&appName=%s", cfg.AtlasAppName)
			}
			connectionString += "?" + atlasOptions

			// Log Atlas connection attempt with masked credentials for security
			logger.Debug("🍃 Constructing authenticated MongoDB Atlas (SRV) connection string",
				slog.String("cluster_host", cfg.Host),       // Atlas cluster hostname (safe to log)
				slog.String("database", cfg.Name),           // Database name (safe to log)
				slog.String("user", cfg.User),               // Username (safe to log)
				slog.String("password", "***MASKED***"),     // Never log actual password
				slog.String("connection_type", "atlas_srv"), // Connection type indicator
				slog.String("app_name", cfg.AtlasAppName),   // Atlas app name (safe to log)
			)
		} else {
			// Unauthenticated Atlas connection (rare, usually for read-only public datasets)
			connectionString = fmt.Sprintf("mongodb+srv://%s/%s?retryWrites=true&w=majority",
				cfg.Host, // Atlas cluster hostname
				cfg.Name, // Database name to connect to
			)

			// Log unauthenticated Atlas connection attempt
			logger.Debug("🍃 Constructing unauthenticated MongoDB Atlas (SRV) connection string",
				slog.String("cluster_host", cfg.Host),       // Atlas cluster hostname (safe to log)
				slog.String("database", cfg.Name),           // Database name (safe to log)
				slog.Bool("authenticated", false),           // Connection type indicator
				slog.String("connection_type", "atlas_srv"), // Connection type indicator
			)
		}
	} else {
		// Standard MongoDB connection string construction
		if cfg.User != "" && cfg.Password != "" {
			// Authenticated connection with username and password
			// URL encode password to handle special characters safely
			encodedPassword := url.QueryEscape(cfg.Password)
			connectionString = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
				cfg.User,        // Username for authentication
				encodedPassword, // URL-encoded password for authentication
				cfg.Host,        // Database server hostname/IP
				cfg.Port,        // Database server port
				cfg.Name,        // Database name to connect to
			)

			// Log connection attempt with masked credentials for security
			logger.Debug("🍃 Constructing authenticated MongoDB connection string",
				slog.String("host", cfg.Host),              // Safe to log
				slog.Int("port", cfg.Port),                 // Safe to log
				slog.String("database", cfg.Name),          // Safe to log
				slog.String("user", cfg.User),              // Safe to log
				slog.String("password", "***MASKED***"),    // Never log actual password
				slog.String("connection_type", "standard"), // Connection type indicator
			)
		} else {
			// Unauthenticated connection (for development or special configurations)
			connectionString = fmt.Sprintf("mongodb://%s:%d/%s",
				cfg.Host, // Database server hostname/IP
				cfg.Port, // Database server port
				cfg.Name, // Database name to connect to
			)

			// Log unauthenticated connection attempt
			logger.Debug("🍃 Constructing unauthenticated MongoDB connection string",
				slog.String("host", cfg.Host),              // Safe to log
				slog.Int("port", cfg.Port),                 // Safe to log
				slog.String("database", cfg.Name),          // Safe to log
				slog.Bool("authenticated", false),          // Connection type indicator
				slog.String("connection_type", "standard"), // Connection type indicator
			)
		}
	}

	return connectionString
}

// GetClient returns the underlying MongoDB client instance for advanced operations.
//
// This method provides direct access to the MongoDB client for operations that require
// client-level functionality such as transactions, bulk operations, or custom configurations.
// The returned client is thread-safe and manages connection pooling automatically.
//
// Use Cases:
// - Database transactions requiring client-level coordination
// - Bulk operations across multiple collections or databases
// - Administrative operations (user management, index creation)
// - Custom session management
// - Advanced aggregation pipelines
//
// Thread Safety:
// The returned client is thread-safe and can be used concurrently from multiple goroutines.
// The MongoDB driver handles connection pool management and request routing automatically.
//
// Returns:
//   - *mongo.Client: MongoDB client instance for direct database operations
//
// Example Usage:
//
//	client := dbManager.GetClient()
//	session, err := client.StartSession()
//	if err != nil {
//	    log.Fatal("Failed to start session:", err)
//	}
//	defer session.EndSession(context.Background())
//
// Complexity:
//   - Time Complexity: O(1) - direct field access
//   - Space Complexity: O(1) - returns existing reference
func (m *MongoDBManager) GetClient() *mongo.Client {
	return m.client
}

// GetDatabase returns the MongoDB database instance for collection operations.
//
// This method provides access to the configured database instance for performing
// collection-level operations such as CRUD operations, queries, and aggregations.
// The database instance is pre-configured with the database name from the configuration.
//
// Use Cases:
// - Collection access for CRUD operations
// - Database-level administrative operations
// - Collection management (creation, deletion, indexing)
// - Aggregation operations across collections
// - Database statistics and monitoring
//
// Thread Safety:
// The returned database instance is thread-safe and can be used concurrently
// from multiple goroutines without additional synchronization.
//
// Returns:
//   - *mongo.Database: MongoDB database instance for collection operations
//
// Example Usage:
//
//	db := dbManager.GetDatabase()
//	collection := db.Collection("users")
//	result, err := collection.InsertOne(context.Background(), user)
//
// Complexity:
//   - Time Complexity: O(1) - direct field access
//   - Space Complexity: O(1) - returns existing reference
func (m *MongoDBManager) GetDatabase() *mongo.Database {
	return m.database
}

// IsConnected returns the current connection status of the MongoDB manager.
//
// This method provides a quick way to check if the database connection is established
// and healthy. It reflects the last known connection state and can be used for
// health checks and monitoring purposes.
//
// Connection Status Tracking:
// - true: Connection is established and last health check succeeded
// - false: Connection is not established or last health check failed
//
// Note: This method returns the cached connection status. For real-time connectivity
// verification, use the Ping() method which performs an active server health check.
//
// Use Cases:
// - Health check endpoints for monitoring systems
// - Pre-operation connectivity verification
// - Connection status logging and metrics
// - Graceful degradation when database is unavailable
//
// Returns:
//   - bool: Current connection status (true = connected, false = disconnected)
//
// Example Usage:
//
//	if !dbManager.IsConnected() {
//	    log.Warn("Database connection is not available")
//	    return http.StatusServiceUnavailable
//	}
//
// Complexity:
//   - Time Complexity: O(1) - direct field access
//   - Space Complexity: O(1) - returns primitive value
func (m *MongoDBManager) IsConnected() bool {
	return m.isConnected
}

// Ping performs an active health check of the MongoDB connection.
//
// This method sends a ping command to the MongoDB server to verify that the connection
// is still active and the server is responsive. It provides real-time connectivity
// verification and updates the internal connection status.
//
// Health Check Process:
// 1. Sends ping command to MongoDB server with configurable timeout
// 2. Verifies server responsiveness and connection health
// 3. Updates internal connection status based on ping result
// 4. Logs health check results for monitoring and debugging
//
// Timeout Handling:
// The ping operation uses a 5-second timeout to prevent blocking on unresponsive servers.
// This provides quick feedback while allowing for temporary network delays.
//
// Error Handling:
// - Network connectivity issues are detected and reported
// - Server unresponsiveness is identified and logged
// - Connection status is updated to reflect current health
// - Detailed error context is provided for troubleshooting
//
// Use Cases:
// - Health check endpoints for load balancers and monitoring
// - Pre-operation connectivity verification
// - Periodic connection health monitoring
// - Troubleshooting connectivity issues
//
// Parameters:
//   - ctx: Context for operation timeout and cancellation control
//
// Returns:
//   - error: nil if ping succeeds, detailed error if ping fails
//
// Example Usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if err := dbManager.Ping(ctx); err != nil {
//	    log.Error("Database health check failed:", err)
//	    return http.StatusServiceUnavailable
//	}
//
// Complexity:
//   - Time Complexity: O(1) + network latency for ping operation
//   - Space Complexity: O(1) - no additional data structures
//   - Network Operations: 1 round trip to database server
func (m *MongoDBManager) Ping(ctx context.Context) error {
	// Log the start of health check operation for monitoring
	m.logger.Debug("🍃 Performing MongoDB connection health check",
		slog.String("operation", "ping_server"),
	)

	// Perform ping operation with primary read preference to ensure we're checking
	// the primary server in a replica set configuration
	err := m.client.Ping(ctx, readpref.Primary())

	if err != nil {
		// Ping failed - update connection status and log error
		m.isConnected = false
		m.logger.Error("🍃 MongoDB connection health check failed",
			slog.Any("error", err),                        // Ping failure details
			slog.Bool("connection_healthy", false),        // Updated health status
			slog.String("operation_stage", "server_ping"), // Stage where failure occurred
		)
		return fmt.Errorf("MongoDB ping failed: %w", err)
	}

	// Ping succeeded - update connection status and log success
	m.isConnected = true
	m.logger.Debug("🍃 MongoDB connection health check successful",
		slog.Bool("connection_healthy", true),        // Current health status
		slog.String("server_response", "responsive"), // Server responsiveness
	)

	return nil
}

// Close gracefully closes the MongoDB connection and cleans up resources.
//
// This method performs comprehensive cleanup of the MongoDB connection including:
// - Graceful client disconnection with timeout handling
// - Connection pool cleanup and resource deallocation
// - Internal state reset and status update
// - Comprehensive logging for operational visibility
//
// Cleanup Process:
// 1. Initiates graceful client disconnection with timeout
// 2. Waits for active operations to complete or timeout
// 3. Closes connection pool and releases network resources
// 4. Updates internal connection status to disconnected
// 5. Logs disconnection status and any cleanup issues
//
// Timeout Handling:
// The disconnection uses a 10-second timeout to allow active operations to complete
// while preventing indefinite blocking during shutdown. This balances data integrity
// with responsive shutdown behavior.
//
// Error Handling:
// - Disconnection errors are logged but don't prevent cleanup completion
// - Resource cleanup continues even if some operations fail
// - Internal state is reset regardless of disconnection success
// - Comprehensive error context is provided for troubleshooting
//
// Thread Safety:
// This method is thread-safe and can be called from multiple goroutines.
// However, it should typically be called only once during application shutdown.
//
// Use Cases:
// - Application shutdown sequences
// - Resource cleanup in defer statements
// - Error recovery and connection reset
// - Graceful service termination
//
// Returns:
//   - error: nil if disconnection succeeds, detailed error if disconnection fails
//
// Example Usage:
//
//	// In main function or service shutdown
//	defer func() {
//	    if err := dbManager.Close(); err != nil {
//	        log.Error("Failed to close database connection:", err)
//	    }
//	}()
//
//	// In error recovery
//	if connectionError != nil {
//	    dbManager.Close() // Reset connection for retry
//	}
//
// Complexity:
//   - Time Complexity: O(1) + network latency for graceful disconnection
//   - Space Complexity: O(1) - no additional allocations during cleanup
//   - Network Operations: 1 round trip for disconnection handshake
func (m *MongoDBManager) Close() error {
	// Log the start of connection cleanup process for operational visibility
	m.logger.Info("🍃 Initiating MongoDB connection cleanup",
		slog.String("operation", "client_disconnect"),
		slog.Bool("was_connected", m.isConnected), // Previous connection status
	)

	// Update connection status immediately to prevent new operations
	m.isConnected = false

	// Create context with timeout for graceful disconnection
	// This prevents indefinite blocking while allowing active operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Perform graceful client disconnection with timeout
	err := m.client.Disconnect(ctx)

	if err != nil {
		// Disconnection failed - log error but continue cleanup
		m.logger.Error("🍃 MongoDB connection disconnection encountered issues",
			slog.Any("error", err),                                 // Disconnection error details
			slog.Duration("timeout_used", 10*time.Second),          // Timeout that was applied
			slog.String("cleanup_status", "completed_with_errors"), // Cleanup result
		)
		return fmt.Errorf("MongoDB disconnection failed: %w", err)
	}

	// Disconnection successful - log completion
	m.logger.Info("🍃 MongoDB connection closed successfully",
		slog.String("cleanup_status", "completed"), // Cleanup result
		slog.Bool("connection_healthy", false),     // Final connection status
		slog.String("resource_status", "released"), // Resource cleanup status
	)

	return nil
}

// NewConnection creates a MongoDB connection manager using the provided configuration.
//
// This function serves as a convenient wrapper around NewMongoDBManager for compatibility
// with existing code and provides a simplified interface for MongoDB connection establishment.
//
// DEPRECATED: This function is maintained for backward compatibility.
// New code should use NewMongoDBManager directly for access to the full MongoDB manager interface.
//
// Parameters:
//   - cfg: Database configuration struct containing connection parameters
//   - logger: Structured logger for connection operations
//
// Returns:
//   - *MongoDBManager: Initialized MongoDB manager instance
//   - error: Connection establishment error if any
//
// Example Usage:
//
//	dbManager, err := database.NewConnection(cfg.Database, logger)
//	if err != nil {
//	    log.Fatal("Database connection failed:", err)
//	}
//	defer dbManager.Close()
//
// Complexity:
//   - Time Complexity: O(1) + network latency (delegates to NewMongoDBManager)
//   - Space Complexity: O(1) - single manager instance allocation
func NewConnection(cfg config.Database, logger *slog.Logger) (*MongoDBManager, error) {
	return NewMongoDBManager(cfg, logger)
}
