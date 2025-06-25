package config

import (
	"log/slog"
)

// NewDefaultConfig creates and returns a Config struct populated with carefully chosen
// default values that provide safe, development-friendly behavior while maintaining
// production viability when properly overridden.
//
// Default Value Philosophy:
// The default configuration is designed with the following principles:
// 1. Safety First: Secure defaults that don't expose unnecessary attack surfaces
// 2. Development Friendly: Values that work well for local development without setup
// 3. Production Ready: Defaults that can be safely overridden for production use
// 4. Fail Safe: Conservative timeouts and limits that prevent resource exhaustion
// 5. Observable: Logging configuration that provides good visibility during development
//
// Configuration Strategy:
// Default values serve as the foundation layer in the configuration hierarchy.
// They are overridden by:
// 1. Base configuration files (config.json)
// 2. Environment-specific files (config.{env}.json)
// 3. Local development files (config.local.json)
// 4. Environment variables (highest precedence)
//
// These defaults ensure the application can start and function correctly even
// if no configuration files are present, making it robust for various deployment
// scenarios and development environments.
//
// Environment-Specific Behavior:
// While these are "development" defaults, they are designed to be appropriate
// starting points for all environments when properly overridden:
// - Development: Used as-is for local development and testing
// - Test: Overridden with test-specific values (faster timeouts, test databases)
// - Staging: Overridden with production-like values for testing
// - Production: Fully overridden with production-optimized values
//
// Parameters:
//   - logger: Structured logger instance for debugging configuration initialization.
//     Used to log the default configuration creation process for troubleshooting.
//
// Returns:
//   - *Config: Pointer to a fully initialized Config struct with default values.
//     The returned config is ready to use for development scenarios and
//     serves as the base for environment-specific overrides.
//
// Usage Examples:
//
//	// Basic usage for development
//	cfg := config.NewDefaultConfig(slog.Default())
//
//	// In testing scenarios
//	testLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
//	cfg := config.NewDefaultConfig(testLogger)
//
//	// In production (typically not used directly, but as base for overrides)
//	cfg := config.NewDefaultConfig(logger)
//	// cfg is then overridden by production configuration files and environment variables
//
// Thread Safety:
// This function creates a new Config instance on each call and is thread-safe.
// The returned Config should be treated as immutable after creation.
//
// Performance Characteristics:
// - Time Complexity: O(1) - constant time allocation and initialization
// - Space Complexity: O(1) - fixed size struct allocation
// - Memory allocation: Single heap allocation for the Config struct
// - No I/O operations: Pure in-memory initialization
func NewDefaultConfig(logger slog.Logger) *Config {
	// Log the start of default configuration initialization for debugging
	// This helps track configuration creation in complex initialization sequences
	logger.Debug("🔠 Creating default configuration with development-friendly settings")

	return &Config{
		// Environment: Set to "development" as the safest default for local development.
		// This ensures development-friendly behavior (detailed logging, permissive settings)
		// while preventing accidental production deployment with development settings.
		//
		// Override in production environments via environment variables or config files.
		Environment: "development",

		// Logger: Configure structured logging optimized for development visibility.
		// These settings provide maximum debugging information during development
		// while maintaining structured logging compatibility for production overrides.
		Logger: Logger{
			// Level: "debug" provides maximum visibility for development and troubleshooting.
			// This allows developers to see detailed application flow and identify issues quickly.
			// Production environments should override this to "info" or "warn" for performance.
			Level: "debug",

			// Format: "text" provides human-readable output for development environments.
			// Text format is easier to read in terminal output during development.
			// Production environments should override to "json" for log aggregation systems.
			Format: "text",

			// Output: "stdout" works universally across development and deployment environments.
			// Standard output integrates well with development tools and container orchestration.
			// Can be overridden to files or syslog for specific deployment requirements.
			Output: "stdout",

			// AddSource: true includes file and line number information in log messages.
			// This is invaluable for debugging and development but has minimal performance impact.
			// Can be disabled in high-throughput production environments if needed.
			AddSource: true,
		},

		// Server: Configure HTTP server with balanced settings for development and production.
		// These values work well for local development while being reasonable defaults
		// for production when properly overridden with environment-specific values.
		Server: Server{
			// Port: 8080 is a common alternative HTTP port that doesn't require root privileges.
			// This port is widely recognized and commonly used for web applications.
			// Production deployments often use port 80/443 with reverse proxy or load balancer.
			Port: 8080,

			// Host: "localhost" provides secure default that only accepts local connections.
			// This prevents accidental exposure during development while being overrideable
			// for production deployment where "0.0.0.0" is typically required.
			Host: "localhost",

			// ReadTimeout: 30 seconds provides reasonable timeout for most API requests.
			// This prevents slow client attacks while accommodating legitimate slow connections.
			// Can be adjusted based on expected request complexity and network conditions.
			ReadTimeout: 30,

			// WriteTimeout: 30 seconds allows sufficient time for response generation and transmission.
			// This accommodates database queries and API processing while preventing resource exhaustion.
			// Should be tuned based on application response time characteristics.
			WriteTimeout: 30,

			// ShutdownTimeout: 15 seconds provides time for graceful shutdown without being excessive.
			// This allows in-flight requests to complete while not delaying deployments too long.
			// Can be increased for applications with longer-running request processing.
			ShutdownTimeout: 15,
		},

		// Database: Configure database connection settings with secure defaults.
		Database: Database{
			// Host: The hostname or IP address of the database server.
			Host: "localhost",

			// Port: The port number on which the database server is listening.
			Port: 27017,

			// User: The username used to authenticate with the database.
			User: "user",
			
			// Password: The password used to authenticate with the database.
			Password: "pass",

			// Name: The name of the database to connect to.
			Name: "database",
		},

		// Test: Test configuration is typically empty for defaults since test values
		// are usually provided through test-specific configuration files or environment variables.
		// This ensures tests have explicit configuration rather than relying on defaults.
		Test: Test{
			// Test fields are left with zero values (empty strings) to ensure
			// test scenarios provide explicit configuration values rather than
			// relying on potentially inappropriate defaults.
		},
	}
}
