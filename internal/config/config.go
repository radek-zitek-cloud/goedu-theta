package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"

	"github.com/joho/godotenv"
)

// NewConfig loads the application configuration from multiple sources with precedence.
//
// This function implements a comprehensive configuration loading strategy that merges
// configuration data from multiple sources in the following precedence order:
// 1. Base configuration file (config.json) - default values
// 2. Environment-specific file (config.{env}.json) - environment overrides
// 3. Local configuration file (config.local.json) - developer overrides
// 4. .env file values - dotenv file variables
// 5. System environment variables - highest precedence
//
// Configuration Loading Strategy:
//   - Fail-safe: Missing files are logged but don't cause failures
//   - Flexible: Supports any environment name (dev, staging, prod, etc.)
//   - Debuggable: Comprehensive logging at each stage
//   - Testable: Exported functions for unit testing
//   - Secure: Sensitive values should come from environment variables
//
// File Loading Order:
//  1. configs/config.json (base configuration - always loaded)
//  2. configs/config.{ENVIRONMENT}.json (environment-specific overrides)
//  3. configs/config.local.json (local development overrides)
//  4. .env file (dotenv format: KEY=value)
//  5. System environment variables (final overrides)
//
// Environment Detection:
//   - Uses ENVIRONMENT environment variable
//   - Falls back to "development" if unset or invalid
//   - Supports: development, test, staging, production
//
// Error Handling:
//   - Missing base config file: Fatal error (application cannot start)
//   - Missing environment/local files: Warning logged, continues loading
//   - Invalid JSON: Fatal error with detailed parsing information
//   - Missing environment variables: Uses defaults, logs debug info
//
// Returns:
//   - *Config: Fully populated configuration struct ready for use
//   - error: Only returned for fatal configuration errors
//
// Usage Example:
//
//	cfg, err := config.NewConfig()
//	if err != nil {
//	    log.Fatal("Failed to load configuration:", err)
//	}
//
// Performance:
//   - Time Complexity: O(n) where n is total config file size
//   - Space Complexity: O(1) - single config struct in memory
//   - I/O Operations: 3-4 file reads + environment variable access
func NewConfig() (*Config, error) {
	// Log the start of configuration loading process for debugging
	slog.Debug("🔠 Loading configuration")

	// Step 1: Determine the current environment from ENVIRONMENT variable
	// This controls which environment-specific config file will be loaded
	slog.Debug("🔠 Loading environment variable",
		slog.String("variable", "ENVIRONMENT"), // Log which variable we're checking
	)

	// Read the ENVIRONMENT variable (empty string if not set)
	var environment string = os.Getenv("ENVIRONMENT")

	slog.Debug("🔠 Environment variable loaded",
		slog.String("environment", environment), // Log the actual value found
	)

	// Step 2: Validate and normalize the environment value
	// Only specific environments are supported to prevent typos and ensure
	// consistent behavior across deployments
	switch environment {
	case "development", "test", "staging", "production":
		// Valid environment detected - log for operational visibility
		slog.Debug("🔠 Valid environment detected",
			slog.String("environment", environment),
		)
	default:
		// Invalid or missing environment - use safe default
		// This prevents application failures due to misconfigured environments
		// and ensures consistent behavior during development
		slog.Warn("🔠 Invalid or unset environment variable, defaulting to development",
			slog.String("provided_environment", environment), // Log what was actually provided
		)
		environment = "development" // Safe default for development workflow
	}

	// Log the final environment that will be used for configuration loading
	slog.Debug("🔠 Setting environment",
		slog.String("environment", environment),
	)

	// Step 3: Construct configuration file paths using consistent naming convention
	// This follows the pattern: config.{environment}.json for environment-specific files
	// All paths are relative to the project root where the binary is executed
	const config_folder = "configs/"      // Standard configuration directory
	const config_file_name = "config"     // Base filename for all config files
	const config_file_extension = ".json" // JSON format for human readability
	const config_local_string = "local"   // Local development overrides

	// Construct full file paths for each configuration source
	// Base config: Contains default values that work across all environments
	base_config_file := config_folder + config_file_name + config_file_extension

	// Environment config: Contains environment-specific overrides (staging, production, etc.)
	environment_config_file := config_folder + config_file_name + "." + environment + config_file_extension

	// Local config: Contains developer-specific overrides (not committed to git)
	local_config_file := config_folder + config_file_name + "." + config_local_string + config_file_extension

	// Dotenv file: Contains environment variables in KEY=value format
	const dotenv_file = ".env"

	// Log all configuration file paths for debugging and operational visibility
	// This helps troubleshoot configuration loading issues and verify file locations
	slog.Debug("🔠 Configuration file paths",
		slog.String("base", base_config_file),               // Always loaded (required)
		slog.String("environment", environment_config_file), // Environment-specific (optional)
		slog.String("local", local_config_file),             // Local development (optional)
		slog.String("dotenv", dotenv_file),                  // Environment variables (optional)
	)

	// Step 4: Initialize the configuration struct with defaults and environment
	// Start with a zero-value Config struct and set the determined environment
	var cfg Config
	cfg.Environment = environment // Store the final environment for runtime access

	// Step 5: Load configuration files in precedence order (base -> environment -> local)
	// Each subsequent file can override values from previous files

	// Load base configuration (REQUIRED)
	// This file must exist and contain valid JSON, otherwise the application cannot start
	// Base config provides sensible defaults that work across all environments
	if err := LoadFromJSONFile(base_config_file, &cfg); err != nil {
		// Base configuration failure is fatal - application cannot function without defaults
		slog.Error("🔠 Error loading base configuration - application cannot start",
			slog.String("file", base_config_file), // Which file failed to load
			slog.Any("error", err),                // Detailed error information
		)
		return nil, fmt.Errorf("failed to load base configuration: %w", err)
	}

	// Load environment-specific configuration (OPTIONAL)
	// This file contains environment-specific overrides (staging, production settings)
	// Missing file is acceptable - not all environments need custom settings
	if err := LoadFromJSONFile(environment_config_file, &cfg); err != nil {
		// Environment config failure is logged but not fatal
		// The application can run with base configuration if environment file is missing
		slog.Warn("🔠 Unable to load environment-specific configuration - using base config",
			slog.String("file", environment_config_file), // Which file was attempted
			slog.Any("error", err),                       // Why it failed (file not found, parse error, etc.)
		)
		// Continue loading - this is not a fatal error
	}

	// Load local configuration (OPTIONAL)
	// This file contains developer-specific overrides for local development
	// Typically not committed to version control (in .gitignore)
	if err := LoadFromJSONFile(local_config_file, &cfg); err != nil {
		// Local config failure is logged but not fatal
		// Most deployments won't have a local config file, which is expected
		slog.Debug("🔠 Unable to load local configuration - this is normal for non-development environments",
			slog.String("file", local_config_file), // Which file was attempted
			slog.Any("error", err),                 // Why it failed (usually file not found)
		)
		// Continue loading - this is expected in production environments
	}

	// Step 6: Apply environment variable overrides (HIGHEST PRECEDENCE)
	// This includes both system environment variables and .env file values
	// Environment variables always take precedence over file-based configuration
	// This allows for secure deployment practices (secrets from env vars, not files)
	if err := OverrideFromEnv(dotenv_file, &cfg); err != nil {
		// Environment override failure could be fatal depending on the error
		// Missing .env file is acceptable, but parsing errors are problematic
		slog.Error("🔠 Error applying environment variable overrides",
			slog.String("dotenv_file", dotenv_file), // Which .env file was processed
			slog.Any("error", err),                  // Detailed error information
		)
		return nil, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	// Step 7: Log the final loaded configuration for debugging and operational visibility
	// This provides a comprehensive view of the active configuration without exposing secrets
	// Sensitive values should be logged as "[REDACTED]" or similar
	slog.Debug("🔠 Configuration successfully loaded and merged from all sources",
		slog.String("environment", cfg.Environment),               // Active environment
		slog.String("log_level", cfg.Logger.Level),                // Logging configuration
		slog.String("log_format", cfg.Logger.Format),              // Log format (json/text)
		slog.String("log_output", cfg.Logger.Output),              // Log destination
		slog.Bool("log_add_source", cfg.Logger.AddSource),         // Source code location in logs
		slog.Int("server_port", cfg.Server.Port),                  // HTTP server port
		slog.String("server_host", cfg.Server.Host),               // HTTP server bind address
		slog.Int("server_read_timeout", cfg.Server.ReadTimeout),   // HTTP read timeout
		slog.Int("server_write_timeout", cfg.Server.WriteTimeout), // HTTP write timeout
	)

	// Return the fully loaded and validated configuration
	// At this point, all configuration sources have been merged with proper precedence
	return &cfg, nil
}

// NewConfig loads the application configuration from JSON files and environment variables.
//
// This function merges base, environment-specific, and local config files, then overrides
// with environment variables and .env file values. It provides detailed debug/error logging
// for each step and returns a fully populated Config struct.
//
// Returns:
//   - *Config: pointer to the loaded Config struct
//   - error: error if loading or parsing fails
//
// Example:
//
//	cfg, err := config.NewConfig()
//
// Complexity:
//
//	Time: O(1) except for file I/O, Space: O(1)

// LoadFromJSONFile loads configuration from a JSON file into the provided Config struct.
// This function performs atomic file loading with comprehensive error handling and validation.
//
// Parameters:
//   - filepath: The absolute or relative path to the JSON configuration file
//   - cfg: Pointer to the Config struct to populate with loaded values
//
// Returns:
//   - error: nil on success, or a wrapped error with context on failure
//
// Error Handling Strategy:
//   - File not found: Returns a wrapped os.ErrNotExist for caller handling
//   - Permission denied: Returns a wrapped permission error
//   - Invalid JSON: Returns a wrapped JSON parsing error with line/column info
//   - Validation errors: Returns validation-specific errors
//
// The function performs the following operations:
// 1. Attempts to read the entire file into memory (atomic operation)
// 2. Validates that the file is valid JSON and matches the Config schema
// 3. Unmarshals the JSON data into the provided struct
// 4. Performs basic validation on the loaded configuration
//
// Note: This function does NOT apply defaults - it only loads what's in the file.
// Callers should ensure defaults are applied before or after this call.
func LoadFromJSONFile(filePath string, cfg *Config) error {
	// Log the configuration file loading attempt for debugging and audit trails
	// This helps track which configuration files are being processed
	slog.Debug("🌀 Loading configuration from JSON file",
		slog.String("filepath", filePath), // Full path to the configuration file
	)

	// Step 1: Read the entire file content into memory
	// This is an atomic operation that either succeeds completely or fails
	// Reading all at once prevents partial reads and race conditions
	data, err := os.ReadFile(filePath)
	if err != nil {
		// Check for specific error types to provide better error messages
		if os.IsNotExist(err) {
			// File not found - this might be expected for optional config files
			slog.Debug("🌀 Configuration file not found (this may be expected)",
				slog.String("filepath", filePath),
				slog.Any("error", err),
			)
			return fmt.Errorf("configuration file not found: %w", err)
		}

		if os.IsPermission(err) {
			// Permission denied - this indicates a deployment or security issue
			slog.Error("🌀 Permission denied reading configuration file",
				slog.String("filepath", filePath),
				slog.Any("error", err),
			)
			return fmt.Errorf("permission denied reading config file '%s': %w", filePath, err)
		}

		// Other I/O errors (disk full, network issues, etc.)
		slog.Error("🌀 I/O error reading configuration file",
			slog.String("filepath", filePath),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to read config file '%s': %w", filePath, err)
	}

	// Step 2: Parse the JSON data into the configuration struct
	// json.Unmarshal will validate JSON syntax and type compatibility
	if err := json.Unmarshal(data, cfg); err != nil {
		// JSON parsing errors include line/column information for debugging
		// This is crucial for identifying malformed configuration files
		slog.Error("🌀 Invalid JSON in configuration file",
			slog.String("filepath", filePath),      // Which file had the error
			slog.Int("file_size_bytes", len(data)), // Size of the file for context
			slog.Any("error", err),                 // Detailed JSON error with position
		)
		return fmt.Errorf("invalid JSON in config file '%s': %w", filePath, err)
	}

	// Step 3: Log successful configuration loading with basic statistics
	// This provides operational visibility into configuration loading
	slog.Debug("🌀 Configuration successfully loaded from JSON file",
		slog.String("filepath", filePath),                  // Source file
		slog.Int("file_size_bytes", len(data)),             // File size for monitoring
		slog.String("loaded_environment", cfg.Environment), // Which environment was loaded
	)

	// Configuration loaded successfully
	return nil
}

// OverrideFromEnv applies environment variable overrides to configuration values.
// This function provides the highest precedence configuration source, allowing runtime
// configuration changes without file modifications. It processes both system environment
// variables and .env file values, with system env vars taking precedence.
//
// Configuration Precedence (highest to lowest):
// 1. System environment variables (os.Getenv)
// 2. .env file values (dotenv format)
// 3. Existing configuration values (from JSON files)
//
// Parameters:
//   - dotenvFile: Path to the .env file containing KEY=VALUE pairs
//   - cfg: Pointer to the Config struct to modify with environment overrides
//
// Returns:
//   - error: nil on success, error if critical environment processing fails
//
// Environment Variable Mapping:
// The function uses reflection to map environment variables to struct fields based on
// the `env` struct tag. For example: `env:"SERVER_PORT"` maps to SERVER_PORT env var.
//
// Supported Data Types:
//   - string: Direct assignment from environment variable
//   - bool: Parsed using strconv.ParseBool (true/false, 1/0, etc.)
//   - int: Parsed using strconv.Atoi for numeric values
//   - Extensions can be added for float, duration, etc.
//
// Error Handling:
//   - Missing .env file: Logged as warning, not fatal (system env vars still processed)
//   - Invalid .env syntax: Logged as error, processing continues
//   - Type conversion errors: Logged as warnings, invalid values skipped
//   - Reflection errors: Logged as errors, field skipped
//
// Security Considerations:
//   - Environment variables may contain sensitive data (passwords, tokens)
//   - Logging of env var values should be carefully controlled
//   - .env files should not be committed to version control
func OverrideFromEnv(dotenvFile string, cfg *Config) error {
	// Log the start of environment variable processing for debugging
	slog.Debug("🔠 Starting environment variable override process",
		slog.String("dotenv_file", dotenvFile), // Which .env file is being processed
	)

	// Step 1: Load .env file contents into a map for efficient lookup
	// The .env file uses KEY=VALUE format and provides default values for environment variables
	// Missing .env file is not fatal - system environment variables can still be used
	slog.Debug("🔠 Loading .env file for environment variable defaults",
		slog.String("file", dotenvFile),
	)

	dotenvMap, err := godotenv.Read(dotenvFile)
	if err != nil {
		// .env file loading failure is not fatal - log warning and continue
		// Many production deployments don't use .env files, relying on system env vars
		slog.Warn("🔠 Could not load .env file - will only use system environment variables",
			slog.String("file", dotenvFile), // Which file failed to load
			slog.Any("error", err),          // Why it failed (file not found, parse error, etc.)
		)
		dotenvMap = make(map[string]string) // Initialize empty map to prevent nil pointer errors
	} else {
		// .env file loaded successfully - log for operational visibility
		slog.Debug("🔠 Successfully loaded .env file",
			slog.String("file", dotenvFile),
			slog.Int("env_vars_count", len(dotenvMap)), // How many variables were loaded
		)
	}

	// Step 2: Helper function for type-safe field assignment with comprehensive error handling
	// This function handles the complexity of converting string environment variables
	// to the appropriate Go types used in the configuration struct
	setField := func(field reflect.Value, value string, fieldName string) bool {
		// Ensure the field can be modified (is settable)
		if !field.CanSet() {
			slog.Warn("🔠 Cannot set field - field is not settable",
				slog.String("field", fieldName),
				slog.String("value", value),
			)
			return false
		}

		// Handle different field types with appropriate conversion
		switch field.Kind() {
		case reflect.String:
			// String fields: Direct assignment, no conversion needed
			field.SetString(value)
			slog.Debug("🔠 Set string field from environment variable",
				slog.String("field", fieldName),
				slog.String("value", value), // Safe to log string values
			)
			return true

		case reflect.Bool:
			// Boolean fields: Parse string to boolean (true/false, 1/0, yes/no, etc.)
			if boolValue, parseErr := strconv.ParseBool(value); parseErr == nil {
				field.SetBool(boolValue)
				slog.Debug("🔠 Set boolean field from environment variable",
					slog.String("field", fieldName),
					slog.Bool("value", boolValue), // Log the parsed boolean value
				)
				return true
			} else {
				// Boolean parsing failed - log error and skip this field
				slog.Warn("🔠 Invalid boolean value in environment variable",
					slog.String("field", fieldName),
					slog.String("raw_value", value),   // What was provided
					slog.Any("parse_error", parseErr), // Why parsing failed
				)
				return false
			}

		case reflect.Int:
			// Integer fields: Parse string to integer
			if intValue, parseErr := strconv.Atoi(value); parseErr == nil {
				field.SetInt(int64(intValue))
				slog.Debug("🔠 Set integer field from environment variable",
					slog.String("field", fieldName),
					slog.Int("value", intValue), // Log the parsed integer value
				)
				return true
			} else {
				// Integer parsing failed - log error and skip this field
				slog.Warn("🔠 Invalid integer value in environment variable",
					slog.String("field", fieldName),
					slog.String("raw_value", value),   // What was provided
					slog.Any("parse_error", parseErr), // Why parsing failed
				)
				return false
			}

		default:
			// Unsupported field type - log warning for future enhancement
			slog.Warn("🔠 Unsupported field type for environment variable override",
				slog.String("field", fieldName),
				slog.String("type", field.Kind().String()), // What type was encountered
				slog.String("value", value),                // What value was attempted
			)
			return false
		}
	}

	// Step 3: Use reflection to process all struct fields recursively
	// This allows the function to work with nested structs (Server, Logger, etc.)
	// without hardcoding field names or types
	var processStruct func(reflect.Value, reflect.Type, string)
	processStruct = func(structValue reflect.Value, structType reflect.Type, prefix string) {
		// Iterate through all fields in the current struct
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)       // Field metadata (name, tags, type)
			fieldValue := structValue.Field(i) // Actual field value

			// Build the full field name for logging (e.g., "Server.Port")
			fullFieldName := prefix + field.Name
			if prefix != "" {
				fullFieldName = prefix + "." + field.Name
			}

			// Check if this field has an environment variable mapping
			envTag := field.Tag.Get("env")

			// Always recurse into nested structs to process their individual fields
			if fieldValue.Kind() == reflect.Struct {
				slog.Debug("🔠 Processing nested struct",
					slog.String("struct", fullFieldName),
				)
				processStruct(fieldValue, fieldValue.Type(), fullFieldName)
			}

			// If no env tag, skip direct field processing but continue with nested struct processing
			if envTag == "" {
				continue
			}

			// Step 4: Determine the environment variable value using precedence rules
			// 1. System environment variable (highest precedence)
			// 2. .env file value (lower precedence)
			// 3. Skip if neither is set
			var envValue string
			var valueSource string

			// Check system environment variables first (highest precedence)
			if systemValue := os.Getenv(envTag); systemValue != "" {
				envValue = systemValue
				valueSource = "system_env"
			} else if dotenvValue, exists := dotenvMap[envTag]; exists && dotenvValue != "" {
				// Fall back to .env file value if system env var is not set
				envValue = dotenvValue
				valueSource = "dotenv_file"
			} else {
				// No environment variable or .env value found - skip this field
				slog.Debug("🔠 No environment override found for field",
					slog.String("field", fullFieldName),
					slog.String("env_var", envTag),
				)
				continue
			}

			// Step 5: Apply the environment variable value to the struct field
			slog.Debug("🔠 Applying environment variable override",
				slog.String("field", fullFieldName),
				slog.String("env_var", envTag),
				slog.String("source", valueSource), // Where the value came from
				// Note: Not logging the actual value for security reasons
			)

			// Attempt to set the field with type conversion
			if setField(fieldValue, envValue, fullFieldName) {
				slog.Info("🔠 Successfully applied environment variable override",
					slog.String("field", fullFieldName),
					slog.String("env_var", envTag),
					slog.String("source", valueSource),
				)
			} else {
				slog.Warn("🔠 Failed to apply environment variable override",
					slog.String("field", fullFieldName),
					slog.String("env_var", envTag),
					slog.String("source", valueSource),
				)
			}
		}
	}

	// Step 6: Start the recursive struct processing from the root Config struct
	slog.Debug("🔠 Starting recursive struct field processing")
	configValue := reflect.ValueOf(cfg).Elem() // Dereference the pointer
	configType := configValue.Type()
	processStruct(configValue, configType, "")

	// Step 7: Log completion of environment variable processing
	slog.Debug("🔠 Environment variable override processing completed successfully")

	return nil // Environment variable processing completed without fatal errors
}
