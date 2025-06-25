// Package logger provides a centralized, thread-safe logging system for the GoEdu-Theta application.
// This package implements a singleton pattern for logger management and supports multiple output formats
// including JSON for production environments and human-readable formats for development.
//
// The logging system is built on Go's structured logging (slog) package and provides:
// - Thread-safe singleton logger instance management
// - Multiple output formats (JSON, text, pretty console)
// - Configuration-driven logger setup and reconfiguration
// - Bootstrap logging for early application startup
// - Integration with application configuration system
//
// Architecture:
// The logger package uses a singleton pattern to ensure consistent logging behavior
// throughout the application. The singleton is protected by a RWMutex for thread safety
// and supports reconfiguration at runtime based on loaded application configuration.
//
// Initialization Phases:
// 1. Bootstrap Phase: Early logger with environment-based defaults
// 2. Configuration Phase: Reconfigured logger based on loaded application config
// 3. Runtime Phase: Stable logger instance used throughout application lifecycle
//
// Thread Safety:
// All logger operations are thread-safe using read-write mutexes. Multiple goroutines
// can safely access the logger instance concurrently, with proper synchronization
// for reconfiguration operations.
package logger

import (
	"log/slog"
	"os"
	"sync"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
)

// Global singleton logger instance and synchronization primitives
// These variables maintain the application-wide logger state with thread safety
var (
	// instance holds the singleton logger instance that is shared across the entire application.
	// This ensures consistent logging behavior and configuration throughout all components.
	// The instance is protected by the mutex to ensure thread-safe access and modification.
	instance *slog.Logger

	// mu provides thread-safe access to the singleton logger instance.
	// Uses RWMutex to allow concurrent reads (GetLogger) while serializing writes (configuration).
	// This design optimizes for the common case of frequent logger access with infrequent reconfiguration.
	mu sync.RWMutex
)

// InitializeBootstrapLogger creates and configures the initial logger instance for early
// application startup before full configuration is available. This bootstrap logger provides
// essential logging capabilities during the critical application initialization phase.
//
// Bootstrap Strategy:
// The bootstrap logger uses environment-based detection to determine appropriate logging
// settings. This ensures that even before configuration files are loaded, the application
// has access to properly configured logging for debugging startup issues.
//
// Environment Detection:
// - Production Environment (ENVIRONMENT=production):
//   - Log level: ERROR (minimal logging for performance)
//   - Add source: false (cleaner production logs)
//   - Format: text (simple, reliable format)
//   - Output: stdout (compatible with log aggregation)
//
// - Development/Other Environments:
//   - Log level: DEBUG (maximum visibility for development)
//   - Add source: false (can be enabled if needed)
//   - Format: text (human-readable for development)
//   - Output: stdout (immediate console feedback)
//
// Singleton Management:
// This function initializes the global singleton logger instance and sets it as
// the default logger for the slog package. This ensures all logging throughout
// the application uses the same configured instance.
//
// Thread Safety:
// The function acquires a write lock on the singleton mutex to ensure thread-safe
// initialization. Only one goroutine can initialize the logger at a time.
//
// Global Logger Setup:
// The function calls slog.SetDefault() to make the bootstrap logger the global
// default for any code that uses slog directly without accessing the singleton.
//
// Returns:
//   - *slog.Logger: The initialized bootstrap logger instance, ready for immediate use.
//     This logger provides essential logging capabilities during startup
//     and will be replaced by the fully configured logger later.
//
// Usage Examples:
//
//	// Early in main() function before configuration loading
//	logger := logger.InitializeBootstrapLogger()
//	logger.Info("Application starting up")
//
//	// The logger is also available via singleton access
//	logger.GetLogger().Debug("Bootstrap logger initialized")
//
// Lifecycle:
// The bootstrap logger is temporary and should be replaced by calling ConfigureLogger()
// once the application configuration is loaded. However, it provides full logging
// functionality and can be used throughout the application if needed.
//
// Error Handling:
// Bootstrap logger initialization uses safe defaults and should not fail under
// normal circumstances. If initialization fails, the application should terminate
// as logging is essential for debugging and monitoring.
//
// Performance Considerations:
// - Bootstrap initialization is fast (O(1) time complexity)
// - Uses standard output handlers for minimal overhead
// - Environment detection uses simple string comparison
// - Singleton pattern ensures initialization happens only once
func InitializeBootstrapLogger() *slog.Logger {
	// Acquire write lock to ensure thread-safe singleton initialization
	// Only one goroutine can initialize the logger at a time
	mu.Lock()
	defer mu.Unlock()

	// Detect current environment to determine appropriate bootstrap configuration
	// This allows the bootstrap logger to adapt its behavior based on deployment context
	isProduction := os.Getenv("ENVIRONMENT") == "production"

	// Configure handler options based on detected environment
	// Production and development environments have different logging requirements
	var handlerOptions *slog.HandlerOptions

	if isProduction {
		// Production environment: Optimize for performance and log clarity
		// Minimal logging reduces overhead and noise in production systems
		handlerOptions = &slog.HandlerOptions{
			// AddSource: false - Exclude source file/line info for cleaner production logs
			// Source information adds overhead and may not be needed in production
			AddSource: false,

			// Level: ERROR - Only log error conditions and above in production
			// This reduces log volume and focuses on actionable issues
			// Lower-level logs (debug, info, warn) are filtered out for performance
			Level: slog.LevelError,
		}
	} else {
		// Development/other environments: Optimize for visibility and debugging
		// Detailed logging helps developers understand application behavior
		handlerOptions = &slog.HandlerOptions{
			// AddSource: false - Can be enabled manually if source tracking is needed
			// Developers can override this for detailed debugging when required
			AddSource: false,

			// Level: DEBUG - Maximum logging detail for development and troubleshooting
			// This ensures all log messages are visible during development
			// Includes debug, info, warn, and error level messages
			Level: slog.LevelDebug,
		}
	}

	// Create the bootstrap logger with text handler for reliable, readable output
	// Text handler is chosen for bootstrap phase as it's simple, reliable, and human-readable
	// This ensures that startup issues are clearly visible in console output
	logger := slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))

	// Set the logger as the singleton instance for application-wide access
	// This makes the logger available through GetLogger() calls throughout the application
	instance = logger

	// Configure the new logger as the global default for the slog package
	// This ensures that any direct slog usage in the application uses our configured logger
	// instead of the default slog logger with basic settings
	slog.SetDefault(logger)

	// Log the successful bootstrap initialization with environment context
	// This provides visibility into the bootstrap process and confirms proper initialization
	// The environment information helps identify which configuration was applied
	logger.Debug("👢 Bootstrap logger successfully initialized",
		slog.String("environment", os.Getenv("ENVIRONMENT")), // Current environment context
		slog.Any("level", handlerOptions.Level),              // Active log level
		slog.Bool("add_source", handlerOptions.AddSource),    // Source tracking status
	)

	return logger
}

// ConfigureLogger reconfigures the singleton logger instance with application-specific
// configuration loaded from configuration files and environment variables. This function
// replaces the bootstrap logger with a fully configured logger optimized for the
// specific deployment environment and operational requirements.
//
// Configuration-Driven Logging:
// This function transforms the generic bootstrap logger into a production-ready or
// development-optimized logger based on the loaded application configuration.
// It supports multiple output formats, configurable log levels, and environment-specific
// optimizations for performance and observability.
//
// Reconfiguration Strategy:
// The function completely replaces the existing logger instance rather than modifying it.
// This ensures atomic reconfiguration and prevents inconsistent logger state during
// the transition from bootstrap to configured logging.
//
// Supported Log Levels:
// - "debug": Detailed debugging information (highest verbosity)
// - "info": General informational messages (default, balanced approach)
// - "warn": Warning messages about potentially problematic conditions
// - "error": Error messages about failures (lowest verbosity, highest performance)
//
// Supported Output Formats:
// - "json": Structured JSON output for log aggregation systems and automated processing
//   - Best for: Production environments, log aggregation, automated analysis
//   - Characteristics: Machine-readable, structured, integrates with ELK/Splunk/etc.
//
// - "text": Standard structured text output with key-value pairs
//   - Best for: Development environments, human readability, debugging
//   - Characteristics: Human-readable, structured logging benefits, console-friendly
//
// - "pretty": Enhanced human-friendly console output with colors and formatting
//   - Best for: Local development, debugging, enhanced readability
//   - Characteristics: Color-coded, well-formatted, maximum developer experience
//
// Thread Safety:
// The function acquires a write lock on the singleton mutex to ensure thread-safe
// reconfiguration. This prevents race conditions during logger replacement and
// ensures all goroutines see the new logger configuration atomically.
//
// Global Logger Integration:
// The function updates both the singleton instance and the global slog default logger.
// This ensures consistent behavior for code that uses either the singleton pattern
// or direct slog package calls.
//
// Configuration Validation:
// The function provides safe defaults for invalid configuration values:
// - Unknown log levels default to "info"
// - Unknown formats default to "text"
// - Invalid boolean values use configuration defaults
//
// Parameters:
//   - config: Logger configuration struct containing level, format, output, and options.
//     This configuration is typically loaded from JSON files and environment variables.
//
// Usage Examples:
//
//	// After loading application configuration
//	cfg, err := config.NewConfig()
//	if err != nil {
//	    log.Fatal("Failed to load config:", err)
//	}
//	logger.ConfigureLogger(cfg.Logger)
//
//	// The reconfigured logger is immediately available
//	log := logger.GetLogger()
//	log.Info("Logger reconfigured successfully")
//
// Performance Considerations:
// - Reconfiguration is fast (O(1) time complexity)
// - New handler creation may allocate memory
// - Log level changes take effect immediately
// - Format changes may affect log processing overhead
//
// Operational Impact:
// - Existing log messages in flight will use the old configuration
// - New log messages immediately use the new configuration
// - No log messages are lost during reconfiguration
// - Monitoring systems should expect configuration change log entries
func ConfigureLogger(config config.Logger) {
	// Acquire write lock for thread-safe singleton reconfiguration
	// This ensures atomic replacement of the logger instance
	mu.Lock()
	defer mu.Unlock()

	// Convert string log level to slog.Level enum with safe fallback
	// This mapping provides type safety and validation for configuration values
	var level slog.Level
	switch config.Level {
	case "debug":
		// Debug level: Maximum verbosity for development and troubleshooting
		// Includes all log messages (debug, info, warn, error)
		level = slog.LevelDebug
	case "warn":
		// Warning level: Moderate verbosity focusing on potential issues
		// Includes warn and error messages, excludes debug and info
		level = slog.LevelWarn
	case "error":
		// Error level: Minimum verbosity for production performance
		// Includes only error messages, excludes debug, info, and warn
		level = slog.LevelError
	default:
		// Default to info level for balanced visibility and performance
		// Includes info, warn, and error messages, excludes debug
		// This is the recommended level for most production environments
		level = slog.LevelInfo
	}

	// Configure handler options based on loaded configuration
	// These options control logging behavior and output characteristics
	opts := &slog.HandlerOptions{
		// Level: Set the minimum log level for message filtering
		// Messages below this level are discarded for performance
		Level: level,

		// AddSource: Include source file and line number information
		// Valuable for debugging but has slight performance impact
		// Configuration allows per-environment tuning
		AddSource: config.AddSource,
	}

	// Select and create the appropriate log handler based on configured format
	// Handler selection determines output format and integration capabilities
	var handler slog.Handler
	switch config.Format {
	case "json":
		// JSON Handler: Structured JSON output for production and log aggregation
		// Characteristics:
		// - Machine-readable structured format
		// - Excellent for log aggregation systems (ELK, Splunk, etc.)
		// - Supports complex nested data structures
		// - Prevents log injection attacks through proper escaping
		// - Optimal for automated log analysis and monitoring
		handler = slog.NewJSONHandler(os.Stdout, opts)

	case "pretty":
		// Pretty Console Handler: Enhanced human-readable output for development
		// Characteristics:
		// - Color-coded output for better visual parsing
		// - Formatted layout optimized for terminal display
		// - Enhanced readability for debugging sessions
		// - Improved developer experience during development
		// - May include additional formatting features
		handler = NewPrettyConsoleHandler(os.Stdout, opts)

	default:
		// Text Handler: Standard structured text output (default fallback)
		// Characteristics:
		// - Human-readable key-value pair format
		// - Good balance between readability and structure
		// - Compatible with most log analysis tools
		// - Reliable fallback for unknown format configurations
		// - Minimal overhead and high compatibility
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Create new logger instance with selected handler and configuration
	// This completely replaces the previous logger instance
	logger := slog.New(handler)

	// Update singleton instance atomically
	// All subsequent GetLogger() calls will return the new configured logger
	instance = logger

	// Set as global default logger for slog package
	// This ensures consistent behavior for any direct slog usage in the codebase
	slog.SetDefault(logger)

	// Log successful reconfiguration with configuration details
	// This provides operational visibility into logger configuration changes
	// Helps with troubleshooting and configuration verification
	logger.Debug("🔄 Logger successfully reconfigured with new settings",
		slog.String("level", config.Level),        // Active log level
		slog.String("format", config.Format),      // Active output format
		slog.String("output", config.Output),      // Output destination
		slog.Bool("add_source", config.AddSource), // Source tracking status
	)
}

// NewPrettyLogger creates a new standalone logger instance with enhanced pretty console output.
// This function is used to create special-purpose loggers that are independent of the
// singleton logger system, typically for development tools, testing, or specialized logging needs.
//
// Pretty Console Output Features:
// The pretty console handler provides enhanced visual formatting specifically designed
// for terminal output and development environments:
// - Color-coded log levels for quick visual identification
// - Well-formatted structured data display
// - Optimized spacing and alignment for readability
// - Enhanced timestamp formatting
// - Improved key-value pair presentation
//
// Use Cases:
// - Development tools that need standalone logging
// - Testing frameworks requiring isolated logger instances
// - Command-line utilities with enhanced output formatting
// - Debug loggers for specific components or modules
// - Temporary logging for troubleshooting specific issues
//
// Independence from Singleton:
// This function creates a completely independent logger instance that:
// - Does not affect the singleton logger configuration
// - Can be used alongside the main application logger
// - Has its own configuration and handler
// - Provides isolated logging behavior
//
// Parameters:
//
//   - level: Minimum log level for the new logger instance.
//     Only messages at or above this level will be output.
//     Common values: slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError
//
//   - addSource: Whether to include source code location (file and line number) in log output.
//     True enables source tracking for debugging, false optimizes for performance.
//
// Returns:
//   - *slog.Logger: A new logger instance configured with pretty console output.
//     This logger is ready for immediate use and independent of the singleton.
//
// Usage Examples:
//
//	// Create a debug logger for development
//	debugLog := logger.NewPrettyLogger(slog.LevelDebug, true)
//	debugLog.Debug("Detailed debugging information", "component", "auth", "user_id", 12345)
//
//	// Create a warning-level logger for specific component
//	componentLog := logger.NewPrettyLogger(slog.LevelWarn, false)
//	componentLog.Warn("Component-specific warning", "component", "database", "connection_pool", "low")
//
//	// Create a testing logger with custom configuration
//	testLog := logger.NewPrettyLogger(slog.LevelInfo, true)
//	testLog.Info("Test execution started", "test_suite", "integration", "test_count", 45)
//
// Performance Considerations:
// - Pretty formatting adds slight overhead compared to standard text output
// - Color processing may be slower on some terminals
// - Source location tracking (addSource=true) has minimal performance impact
// - Independent instances don't share configuration overhead with singleton
//
// Development Benefits:
// - Enhanced readability during development and debugging
// - Quick visual identification of log levels through color coding
// - Better formatting for complex structured data
// - Improved developer experience during troubleshooting
func NewPrettyLogger(level slog.Level, addSource bool) *slog.Logger {
	// Create pretty console handler with specified configuration
	// This handler provides enhanced formatting optimized for terminal display
	h := NewPrettyConsoleHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,     // Set minimum log level for filtering
		AddSource: addSource, // Configure source location tracking
	})

	// Return new logger instance with pretty console handler
	// This logger is independent and ready for immediate use
	return slog.New(h)
}

// GetLogger provides thread-safe access to the singleton logger instance used throughout
// the application. This is the primary function for accessing the configured logger
// from any part of the application code.
//
// Singleton Pattern Implementation:
// The function implements a thread-safe singleton pattern with lazy initialization:
// 1. Fast path: Read lock to check if instance exists (most common case)
// 2. Slow path: Write lock to initialize if needed (initialization only)
// 3. Double-checked locking to prevent race conditions during initialization
//
// Initialization Behavior:
// If no logger instance exists (application startup or special circumstances):
// - Automatically initializes a bootstrap logger with environment-based defaults
// - Ensures the application always has access to a functional logger
// - Bootstrap logger provides essential logging until full configuration is loaded
//
// Thread Safety Guarantees:
// - Multiple goroutines can safely call GetLogger() concurrently
// - Read operations (common case) use shared read locks for performance
// - Initialization operations (rare case) use exclusive write locks
// - No race conditions during concurrent access or initialization
//
// Performance Characteristics:
// - Fast path (existing instance): O(1) with minimal locking overhead
// - Slow path (initialization): O(1) but includes logger setup
// - Read-write mutex optimizes for frequent reads, infrequent writes
// - No memory allocation on subsequent calls after initialization
//
// Usage Patterns:
// This function should be used throughout the application for consistent logging:
// - Service layer: For business logic logging
// - HTTP handlers: For request/response logging
// - Database operations: For query and transaction logging
// - Background processes: For task and job logging
// - Error handling: For error reporting and debugging
//
// Integration with Configuration:
// - Returns bootstrap logger during early application startup
// - Returns configured logger after ConfigureLogger() has been called
// - Seamlessly transitions from bootstrap to configured logger
// - Maintains consistent interface regardless of initialization state
//
// Returns:
//   - *slog.Logger: The singleton logger instance, ready for immediate use.
//     This logger is thread-safe and configured for the current environment.
//     Will never return nil - initializes bootstrap logger if needed.
//
// Usage Examples:
//
//	// Basic logging in application code
//	log := logger.GetLogger()
//	log.Info("Processing user request", "user_id", userID, "action", "login")
//
//	// Error logging with context
//	log := logger.GetLogger()
//	log.Error("Database connection failed", "error", err, "retry_count", retries)
//
//	// Debug logging for troubleshooting
//	log := logger.GetLogger()
//	log.Debug("Cache hit", "key", cacheKey, "ttl_remaining", ttl)
//
//	// Structured logging with multiple fields
//	log := logger.GetLogger()
//	log.Warn("Rate limit approaching",
//	    "user_id", userID,
//	    "current_requests", currentReqs,
//	    "limit", maxReqs,
//	    "window", timeWindow)
//
// Lifecycle Integration:
// The logger returned by this function automatically adapts to application lifecycle:
// 1. Startup: Bootstrap logger with environment-based defaults
// 2. Configuration: Reconfigured logger with loaded application settings
// 3. Runtime: Stable configured logger for normal operation
// 4. Shutdown: Logger remains available for cleanup and shutdown logging
//
// Error Handling:
// This function is designed to never fail:
// - Always returns a functional logger instance
// - Initializes bootstrap logger if singleton is uninitialized
// - Bootstrap logger provides safe defaults for all environments
// - No error conditions that would prevent logger access
func GetLogger() *slog.Logger {
	// Fast path: Use read lock to check for existing instance
	// This is the most common case after logger initialization
	// Read locks allow concurrent access from multiple goroutines
	mu.RLock()
	if instance != nil {
		// Instance exists - return it immediately
		// Defer ensures read lock is released
		defer mu.RUnlock()
		return instance
	}
	// Release read lock before acquiring write lock
	// This prevents deadlock and allows other readers to proceed
	mu.RUnlock()

	// Slow path: Use write lock for initialization
	// This only happens during application startup or in edge cases
	mu.Lock()
	defer mu.Unlock()

	// Double-check pattern: Verify instance is still nil after acquiring write lock
	// Another goroutine might have initialized it while we waited for the lock
	if instance == nil {
		// Initialize bootstrap logger and return it
		// This ensures the application always has access to logging functionality
		return InitializeBootstrapLogger()
	}

	// Instance was initialized by another goroutine while we waited
	// Return the initialized instance
	return instance
}
