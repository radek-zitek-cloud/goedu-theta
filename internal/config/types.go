package config

// Config represents the complete application configuration structure for GoEdu-Theta.
// This is the root configuration struct that encompasses all application settings,
// providing a centralized and type-safe way to manage application behavior across
// different environments and deployment scenarios.
//
// Configuration Architecture:
// The configuration system follows a hierarchical approach with multiple sources:
// 1. Base configuration (config.json) - Default values for all environments
// 2. Environment-specific config (config.{env}.json) - Environment overrides
// 3. Local development config (config.local.json) - Developer-specific settings
// 4. Environment variables - Runtime overrides with highest precedence
// 5. .env file - Development-friendly environment variable defaults
//
// Field Mapping Strategy:
// Each field includes multiple struct tags for maximum compatibility:
// - `json`: For JSON configuration files (primary storage format)
// - `yaml`: For YAML configuration files (alternative format support)
// - `env`: For environment variable mapping (runtime overrides)
//
// Environment Variable Naming Convention:
// Environment variables follow a hierarchical naming pattern:
// - Root level: ENVIRONMENT, DATABASE_URL, API_KEY, etc.
// - Nested structs: SERVER_PORT, LOGGER_LEVEL, etc.
// - This provides clear namespacing and prevents variable name conflicts
//
// Usage Examples:
//
//	// Load configuration from all sources
//	cfg, err := config.NewConfig()
//	if err != nil {
//	    log.Fatal("Failed to load configuration:", err)
//	}
//
//	// Access nested configuration
//	serverPort := cfg.Server.Port
//	logLevel := cfg.Logger.Level
//
// Thread Safety:
// Configuration structs are designed to be immutable after loading.
// They should be loaded once at application startup and passed as
// read-only references throughout the application.
//
// Validation:
// Field validation should be performed after loading to ensure
// all required values are present and within acceptable ranges.
type Config struct {
	// Environment specifies the current application environment and determines
	// which configuration files are loaded and what behavior is enabled.
	//
	// Valid values: "development", "test", "staging", "production"
	// Default: "development" (safe fallback for missing/invalid values)
	//
	// Environment-specific behaviors:
	// - development: Detailed logging, hot-reload, development middleware
	// - test: Minimal logging, test database, no external dependencies
	// - staging: Production-like setup with additional debugging features
	// - production: Optimized performance, minimal logging, security hardening
	//
	// This field is used throughout the application to conditionally enable
	// features, adjust logging levels, and configure external service connections.
	Environment string `json:"environment" yaml:"environment" env:"ENVIRONMENT"`

	// Logger contains complete logging system configuration including level,
	// format, output destination, and debug features.
	//
	// The logging configuration is critical for operational visibility and
	// debugging in production environments. It supports structured logging
	// with JSON format for log aggregation systems and human-readable text
	// format for development environments.
	//
	// Key features:
	// - Configurable log levels (debug, info, warn, error)
	// - Multiple output formats (JSON for production, text for development)
	// - Flexible output destinations (stdout, stderr, files)
	// - Source code location tracking for debugging
	Logger Logger `json:"logger" yaml:"logger" env:"LOGGER"`

	// Server contains HTTP server configuration including network settings,
	// timeouts, and performance tuning parameters.
	//
	// The server configuration affects application performance, security,
	// and resource utilization. Proper timeout configuration prevents
	// resource exhaustion and ensures graceful degradation under load.
	//
	// Key features:
	// - Configurable bind address and port
	// - Request/response timeout settings
	// - Graceful shutdown timeout
	// - Performance and security tuning options
	Server Server `json:"server" yaml:"server" env:"SERVER"`

	// Database contains the database connection configuration including
	// host, port, user, password, and database name.
	//
	// This configuration is used to establish connections to the database
	// and should be kept secure. Sensitive information like passwords
	// should be stored securely and not hard-coded.
	Database Database `json:"database" yaml:"database" env:"DATABASE"`

	// Test contains test-specific configuration used during automated testing,
	// integration testing, and quality assurance processes.
	//
	// This configuration section allows tests to modify application behavior
	// without affecting production settings. It supports test isolation,
	// fixture management, and environment-specific test scenarios.
	//
	// Test configuration is only loaded and used when Environment is set to "test"
	// or when running automated test suites.
	Test Test `json:"test" yaml:"test" env:"TEST"`
}

// Logger defines the complete logging system configuration for structured and efficient
// application logging across all environments and deployment scenarios.
//
// The logging system is built on Go's structured logging (slog) package and provides
// comprehensive control over log output, formatting, and debugging features.
// This configuration directly impacts application observability, debugging capabilities,
// and operational monitoring in production environments.
//
// Logging Strategy:
// The application uses structured logging with key-value pairs for better searchability
// and analysis in log aggregation systems. Different environments use different
// configurations to balance between debugging detail and performance.
//
// Environment-Specific Defaults:
// - Development: level=debug, format=text, output=stdout, add_source=true
// - Test: level=warn, format=text, output=stdout, add_source=false
// - Staging: level=info, format=json, output=stdout, add_source=true
// - Production: level=info, format=json, output=stdout, add_source=false
//
// Log Aggregation:
// JSON format is recommended for production environments as it integrates well
// with log aggregation systems like ELK Stack, Splunk, or cloud logging services.
// Text format is human-readable and better for development environments.
//
// Performance Considerations:
// - Debug level logging can impact performance in high-throughput scenarios
// - Source code location tracking (add_source) has minimal performance impact
// - JSON formatting is slightly more expensive than text but provides better structure
//
// Security Considerations:
// - Sensitive data should never be logged directly
// - Use structured logging fields to avoid injection attacks
// - Log rotation and retention policies should be configured at the infrastructure level
type Logger struct {
	// Level controls the minimum log level that will be output by the logging system.
	// This is a critical performance and observability setting that determines
	// which log messages are processed and which are discarded.
	//
	// Supported levels (in order of increasing severity):
	// - "debug": Detailed debugging information, typically only enabled during development
	// - "info": General informational messages about application operation
	// - "warn": Warning messages about potentially problematic situations
	// - "error": Error messages about failures that don't stop the application
	//
	// Level Selection Guidelines:
	// - Development: "debug" for maximum visibility into application behavior
	// - Test: "warn" to reduce noise while catching important issues
	// - Staging: "info" to monitor application behavior in production-like environment
	// - Production: "info" or "warn" to balance observability with performance
	//
	// Environment variable: SLOG_LEVEL
	// Default: "info" (balanced approach for most environments)
	Level string `json:"level" yaml:"level" env:"SLOG_LEVEL"`

	// Format determines the output format for log messages, affecting both
	// human readability and machine parsing capabilities.
	//
	// Supported formats:
	// - "json": Structured JSON output, ideal for log aggregation systems
	//   Example: {"time":"2024-01-01T12:00:00Z","level":"INFO","msg":"Server started","port":8080}
	// - "text": Human-readable text output, ideal for development and debugging
	//   Example: 2024-01-01 12:00:00 INFO Server started port=8080
	//
	// Format Selection Guidelines:
	// - Development: "text" for easy reading during development
	// - Test: "text" for simple test output analysis
	// - Staging/Production: "json" for structured log analysis and aggregation
	//
	// The JSON format is recommended for production environments as it:
	// - Integrates seamlessly with log aggregation systems
	// - Supports complex structured data
	// - Enables advanced querying and filtering
	// - Prevents log injection attacks
	//
	// Environment variable: SLOG_FORMAT
	// Default: "json" (production-ready default)
	Format string `json:"format" yaml:"format" env:"SLOG_FORMAT"`

	// Output specifies the destination for log messages, allowing flexible
	// log routing for different deployment scenarios and infrastructure setups.
	//
	// Supported outputs:
	// - "stdout": Standard output stream (default, works with most deployment systems)
	// - "stderr": Standard error stream (separates logs from application output)
	// - "/path/to/file": File path for direct file logging
	// - "syslog": System logging daemon (Unix systems)
	// - "journal": systemd journal (systemd-based systems)
	//
	// Output Selection Guidelines:
	// - Container deployments: "stdout" (captured by container orchestration)
	// - Traditional servers: File path or "syslog" for centralized logging
	// - Development: "stdout" for immediate visibility
	// - Debugging: "stderr" to separate from application output
	//
	// Infrastructure Integration:
	// Most modern deployment platforms (Docker, Kubernetes, cloud services)
	// expect applications to log to stdout/stderr and handle log collection
	// and routing at the infrastructure level.
	//
	// Environment variable: SLOG_OUTPUT
	// Default: "stdout" (universal compatibility)
	Output string `json:"output" yaml:"output" env:"SLOG_OUTPUT"`

	// AddSource controls whether source code location information (file name and line number)
	// is included in log messages. This is valuable for debugging but has slight performance impact.
	//
	// When enabled, log messages include:
	// - Source file path relative to the project root
	// - Line number where the log statement was executed
	// - Function name (in some cases)
	//
	// Example with AddSource enabled:
	// {"time":"2024-01-01T12:00:00Z","level":"INFO","source":"internal/server/server.go:45","msg":"Server started"}
	//
	// Benefits of source information:
	// - Faster debugging and issue resolution
	// - Better understanding of application flow
	// - Easier correlation between logs and code
	// - Improved troubleshooting for production issues
	//
	// Performance Impact:
	// - Minimal CPU overhead (runtime.Caller)
	// - Slight memory overhead for source information storage
	// - Negligible impact on application performance
	//
	// Recommendation:
	// - Enable for development and staging environments
	// - Consider disabling for high-throughput production environments
	// - Enable temporarily for production debugging when needed
	//
	// Environment variable: SLOG_ADD_SOURCE
	// Default: false (performance-first approach)
	AddSource bool `json:"add_source" yaml:"add_source" env:"SLOG_ADD_SOURCE"`
}

// Server defines the complete HTTP server configuration for the GoEdu-Theta web application.
// This configuration directly impacts application availability, performance, security,
// and resource utilization in production environments.
//
// The server is built on the Gin HTTP framework and provides RESTful API endpoints
// with comprehensive middleware for logging, recovery, CORS, and request processing.
// Proper server configuration is critical for handling production traffic loads
// and ensuring application stability under various operating conditions.
//
// Network Architecture:
// The server operates as a single HTTP listener that handles all incoming requests.
// It supports both HTTP/1.1 and HTTP/2 protocols with automatic protocol negotiation.
// TLS/SSL termination can be handled at the load balancer level or within the application.
//
// Performance Characteristics:
// - Request handling: Concurrent request processing with Go goroutines
// - Memory usage: Configurable buffer sizes and connection pooling
// - CPU usage: Efficient request routing and minimal middleware overhead
// - Network I/O: Optimized for high-throughput scenarios with proper timeout handling
//
// Security Considerations:
// - Timeout configuration prevents slowloris attacks and resource exhaustion
// - Host binding controls network interface exposure
// - Request size limits prevent memory exhaustion attacks
// - Graceful shutdown prevents data loss during deployments
//
// Monitoring and Observability:
// The server exposes health check endpoints and metrics for monitoring systems.
// All requests are logged with structured logging for analysis and debugging.
type Server struct {
	// Port specifies the TCP port number on which the HTTP server will listen
	// for incoming client connections. This is a fundamental network configuration
	// that determines how clients access the application.
	//
	// Port Selection Guidelines:
	// - 8080: Standard alternative HTTP port, commonly used for development
	// - 3000-3999: Development ports, often used by Node.js applications
	// - 8000-8999: Alternative HTTP ports, suitable for microservices
	// - 80: Standard HTTP port (requires root privileges on Unix systems)
	// - 443: Standard HTTPS port (requires root privileges and TLS configuration)
	//
	// Environment-Specific Recommendations:
	// - Development: 8080 (avoids privilege requirements, easy to remember)
	// - Test: Dynamic port assignment or standard test ports (8081, 8082)
	// - Staging: Production-like port configuration
	// - Production: Standard ports (80/443) with load balancer or reverse proxy
	//
	// Container Deployment:
	// In containerized environments, the internal port can be mapped to different
	// external ports. Use container port mapping for flexible deployment options.
	//
	// Security Notes:
	// - Ports below 1024 require root privileges on Unix systems
	// - Avoid well-known service ports (22, 25, 53, etc.) to prevent conflicts
	// - Use firewall rules to restrict access to appropriate networks
	//
	// Environment variable: SERVER_PORT
	// Default: 8080 (common development port)
	// Valid range: 1024-65535 (non-privileged ports)
	Port int `json:"port" yaml:"port" env:"SERVER_PORT"`

	// Host specifies the network interface or IP address on which the server
	// will bind and listen for connections. This controls network accessibility
	// and is critical for both security and deployment flexibility.
	//
	// Common Host Values:
	// - "localhost" or "127.0.0.1": Local loopback only (development/testing)
	// - "0.0.0.0": All available network interfaces (production deployment)
	// - Specific IP address: Bind to a particular network interface
	// - "" (empty): Go's default behavior (usually equivalent to "0.0.0.0")
	//
	// Security Implications:
	// - "localhost": Maximum security, only local connections allowed
	// - "0.0.0.0": Exposes service to all networks, requires firewall protection
	// - Specific IP: Selective network exposure, good for multi-interface systems
	//
	// Deployment Scenarios:
	// - Development: "localhost" for security and simplicity
	// - Container: "0.0.0.0" to accept connections from container network
	// - Cloud: "0.0.0.0" with cloud security groups for access control
	// - On-premise: Specific IP addresses based on network architecture
	//
	// Load Balancer Integration:
	// When using load balancers or reverse proxies, the application typically
	// binds to "0.0.0.0" or a specific internal IP address. The load balancer
	// handles external network exposure and security.
	//
	// IPv6 Support:
	// For IPv6 support, use "::" to bind to all IPv6 interfaces or specific
	// IPv6 addresses. Dual-stack configurations require careful host configuration.
	//
	// Environment variable: SERVER_HOST
	// Default: "localhost" (secure development default)
	Host string `json:"host" yaml:"host" env:"SERVER_HOST"`

	// ReadTimeout sets the maximum duration for reading the entire HTTP request,
	// including the request body. This is a critical security and performance
	// setting that prevents slow client attacks and resource exhaustion.
	//
	// Timeout Scope:
	// - Covers the complete request reading process
	// - Includes HTTP headers, request body, and connection establishment
	// - Starts when connection is accepted, ends when request is fully parsed
	//
	// Security Benefits:
	// - Prevents slowloris attacks (slow header transmission)
	// - Protects against slow POST attacks (slow body transmission)
	// - Limits resource consumption from malicious or broken clients
	// - Ensures server responsiveness under attack conditions
	//
	// Performance Impact:
	// - Too short: Legitimate slow connections may timeout
	// - Too long: Server resources tied up by slow clients
	// - Balance based on expected client behavior and network conditions
	//
	// Recommended Values:
	// - API servers: 10-30 seconds (typical for REST APIs)
	// - File upload endpoints: 60-300 seconds (depends on expected file sizes)
	// - Real-time applications: 5-15 seconds (faster response requirements)
	// - Public APIs: 15-60 seconds (accommodate various client types)
	//
	// Network Considerations:
	// - High-latency networks: Increase timeout for mobile/satellite connections
	// - Local networks: Lower timeout acceptable for controlled environments
	// - Global services: Account for international network latency variations
	//
	// Environment variable: SERVER_READ_TIMEOUT
	// Default: 30 seconds (balanced for most use cases)
	// Unit: seconds
	ReadTimeout int `json:"read_timeout" yaml:"read_timeout" env:"SERVER_READ_TIMEOUT"`

	// WriteTimeout sets the maximum duration for writing the HTTP response.
	// This prevents server resources from being tied up by slow or unresponsive
	// clients and ensures consistent response delivery.
	//
	// Timeout Scope:
	// - Covers the complete response writing process
	// - Includes HTTP headers, response body, and connection finalization
	// - Starts when response writing begins, ends when response is fully sent
	//
	// Application Scenarios:
	// - JSON APIs: Quick response writing, short timeout acceptable
	// - File downloads: Long response times, requires longer timeout
	// - Streaming responses: May need extended or disabled timeout
	// - Real-time data: Balance between responsiveness and completion
	//
	// Performance Optimization:
	// - Buffering: Proper response buffering reduces write time
	// - Compression: GZIP compression may increase write time but reduce bandwidth
	// - Caching: Cached responses write faster than dynamically generated content
	//
	// Client Considerations:
	// - Slow mobile connections: May need longer write timeouts
	// - Load balancers: May have their own timeout configurations
	// - CDN integration: Consider CDN timeout settings in conjunction
	//
	// Error Handling:
	// - Write timeout errors are logged for monitoring
	// - Clients may receive incomplete responses on timeout
	// - Connection is forcibly closed after timeout
	//
	// Recommended Values:
	// - Fast APIs: 10-30 seconds (quick JSON responses)
	// - File serving: 60-300 seconds (large file downloads)
	// - Database-heavy: 30-90 seconds (complex query processing)
	// - Microservices: 15-45 seconds (service-to-service communication)
	//
	// Environment variable: SERVER_WRITE_TIMEOUT
	// Default: 30 seconds (suitable for most API responses)
	// Unit: seconds
	WriteTimeout int `json:"write_timeout" yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT"`

	// ShutdownTimeout defines the maximum duration to wait for graceful server shutdown.
	// This is critical for preventing data loss and ensuring clean application termination
	// during deployments, scaling operations, or maintenance activities.
	//
	// Graceful Shutdown Process:
	// 1. Server stops accepting new connections
	// 2. Existing connections are allowed to complete their requests
	// 3. Background tasks are given time to finish
	// 4. Resources are cleaned up and connections are closed
	// 5. Application exits cleanly
	//
	// Benefits of Graceful Shutdown:
	// - Prevents request loss during deployments
	// - Allows database transactions to complete
	// - Ensures proper cleanup of resources and connections
	// - Maintains service availability during rolling updates
	// - Provides better user experience during maintenance
	//
	// Shutdown Triggers:
	// - SIGTERM signal (standard termination request)
	// - SIGINT signal (interrupt signal, Ctrl+C)
	// - Application-specific shutdown commands
	// - Health check failures in orchestration systems
	//
	// Timeout Considerations:
	// - Request completion time: Allow time for longest expected requests
	// - Database operations: Consider transaction commit/rollback time
	// - External API calls: Account for third-party service response times
	// - Resource cleanup: File handles, network connections, memory
	//
	// Container Orchestration:
	// - Kubernetes: Sends SIGTERM, waits for terminationGracePeriodSeconds
	// - Docker: Uses SIGTERM followed by SIGKILL after timeout
	// - Cloud platforms: Various shutdown signal patterns
	//
	// Monitoring Integration:
	// - Health checks should fail during shutdown process
	// - Load balancers should remove instance from rotation
	// - Metrics should capture shutdown duration and success rate
	//
	// Recommended Values:
	// - Fast APIs: 15-30 seconds (quick request completion)
	// - Database-heavy: 30-60 seconds (transaction completion time)
	// - File processing: 60-120 seconds (large operation completion)
	// - Batch processing: 120-300 seconds (batch job completion)
	//
	// Emergency Shutdown:
	// If graceful shutdown exceeds the timeout, the application performs
	// a forced shutdown to prevent indefinite hanging. This may result
	// in connection drops and potential data loss.
	//
	// Environment variable: SERVER_SHUTDOWN_TIMEOUT
	// Default: 30 seconds (balanced approach for most applications)
	// Unit: seconds
	ShutdownTimeout int `json:"shutdown_timeout" yaml:"shutdown_timeout" env:"SERVER_SHUTDOWN_TIMEOUT"`
}

// Database defines the complete database connection configuration for the GoEdu-Theta application.
type Database struct {
	// Host is the hostname or IP address of the database server.
	// For MongoDB Atlas, this should be the cluster hostname (e.g., "clusterzitekcloud.dznruy0.mongodb.net").
	Host string `json:"host" yaml:"host" env:"DATABASE_HOST"`

	// Port is the port number on which the database server is listening.
	// For MongoDB Atlas with SRV connections, this field is ignored as the port is resolved via DNS.
	Port int `json:"port" yaml:"port" env:"DATABASE_PORT"`

	// User is the username used to authenticate with the database.
	User string `json:"user" yaml:"user" env:"DATABASE_USER"`

	// Password is the password used to authenticate with the database.
	Password string `json:"password" yaml:"password" env:"DATABASE_PASSWORD"`

	// Name is the name of the database to connect to.
	Name string `json:"name" yaml:"name" env:"DATABASE_NAME"`

	// IsAtlas indicates whether this is a MongoDB Atlas connection.
	// When true, uses mongodb+srv:// scheme with DNS SRV record resolution.
	// When false, uses standard mongodb:// scheme with direct host:port connection.
	IsAtlas bool `json:"is_atlas" yaml:"is_atlas" env:"DATABASE_IS_ATLAS"`

	// AtlasAppName is the application name for MongoDB Atlas connections.
	// This helps with monitoring and debugging in the Atlas dashboard.
	AtlasAppName string `json:"atlas_app_name" yaml:"atlas_app_name" env:"DATABASE_ATLAS_APP_NAME"`
}

// Test defines configuration settings specifically for testing scenarios and quality assurance.
// This configuration section enables comprehensive testing strategies including unit tests,
// integration tests, end-to-end tests, and performance testing with isolated environments
// and controlled test data.
//
// Testing Philosophy:
// The application supports multiple testing approaches with configuration-driven behavior.
// Test configuration allows tests to modify application behavior without affecting
// production code, enabling comprehensive coverage and realistic testing scenarios.
//
// Test Environment Isolation:
// - Separate configuration values for test scenarios
// - Isolated data sources and external dependencies
// - Configurable test fixtures and mock behaviors
// - Environment-specific test parameters
//
// Test Types Supported:
// - Unit Tests: Individual component testing with mocked dependencies
// - Integration Tests: Component interaction testing with real dependencies
// - End-to-End Tests: Full application workflow testing
// - Performance Tests: Load and stress testing with controlled parameters
// - Contract Tests: API contract validation and compatibility testing
//
// Configuration Override Strategy:
// Test configuration values can override default application configuration
// when the application environment is set to "test". This allows tests to:
// - Use test-specific database connections
// - Configure mock external services
// - Set appropriate timeout values for test scenarios
// - Enable test-specific logging and debugging features
//
// Continuous Integration Integration:
// Test configuration is designed to work seamlessly with CI/CD pipelines:
// - Environment variable overrides for different CI environments
// - Configurable test database connections
// - Flexible test execution parameters
// - Integration with test reporting and coverage tools
type Test struct {
	// Label_default represents the default test identifier used across
	// all test scenarios when no specific override is provided.
	//
	// This field serves as the baseline test configuration value and is used
	// to verify that the configuration loading system properly handles
	// default values from base configuration files.
	//
	// Usage Scenarios:
	// - Configuration loading verification tests
	// - Default behavior validation
	// - Baseline configuration testing
	// - System integration test identification
	//
	// Test Strategy:
	// This value is typically set in the base configuration (config.json)
	// and should remain consistent across all test environments unless
	// specifically overridden for particular test scenarios.
	//
	// Environment variable: TEST_LABEL_DEF
	// Default: Usually set to "default" or application-specific identifier
	Label_default string `json:"label_def" yaml:"label_def" env:"TEST_LABEL_DEF"`

	// Label_env represents test configuration that is derived from
	// environment-specific settings and validates environment variable processing.
	//
	// This field is primarily used to test the environment variable override
	// functionality and ensure that environment-specific test configurations
	// are properly loaded and applied.
	//
	// Testing Applications:
	// - Environment variable parsing validation
	// - Configuration precedence testing (env vars vs config files)
	// - Environment-specific test behavior configuration
	// - CI/CD pipeline test customization
	//
	// Validation Scenarios:
	// - Verify environment variables override configuration files
	// - Test missing environment variable handling
	// - Validate environment variable type conversion
	// - Ensure proper error handling for invalid environment values
	//
	// CI/CD Integration:
	// Different CI/CD environments can set this value to identify
	// the testing context (e.g., "github-actions", "jenkins", "local-dev")
	// allowing tests to adapt behavior based on the execution environment.
	//
	// Environment variable: TEST_LABEL_ENV
	// Default: Typically empty, set by environment variables during testing
	Label_env string `json:"label_env" yaml:"label_env" env:"TEST_LABEL_ENV"`

	// Label_override represents test configuration that demonstrates
	// the configuration override hierarchy and precedence rules.
	//
	// This field is used to validate that the configuration system properly
	// handles multiple configuration sources with the correct precedence order:
	// 1. Environment variables (highest precedence)
	// 2. Local configuration files (config.local.json)
	// 3. Environment-specific files (config.{env}.json)
	// 4. Base configuration (config.json, lowest precedence)
	//
	// Override Testing Scenarios:
	// - Verify local development overrides work correctly
	// - Test environment-specific configuration precedence
	// - Validate that environment variables have highest precedence
	// - Ensure proper fallback when override sources are missing
	//
	// Development Workflow:
	// Developers can use this field to test configuration changes locally
	// without affecting shared configuration files. The local config file
	// (config.local.json) can override this value for development testing.
	//
	// Quality Assurance:
	// QA teams can use this field to verify that configuration changes
	// behave correctly across different environments and deployment scenarios
	// without requiring production-like infrastructure.
	//
	// Security Testing:
	// This field can be used to test configuration injection vulnerabilities
	// and ensure that configuration values are properly validated and sanitized
	// before being used by the application.
	//
	// Environment variable: TEST_LABEL_OVERRIDE
	// Default: Typically set in environment-specific or local config files
	Label_override string `json:"label_override" yaml:"label_override" env:"TEST_LABEL_OVERRIDE"`
}
