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

// Config holds the top-level application configuration.
//
// Fields:
//   - Environment: The current environment (development, test, staging, production)
//   - Logger: Logger configuration struct
//   - Test: Test configuration struct
//
// Each field is tagged for JSON, YAML, and environment variable mapping.
type Config struct {
	Environment string `json:"environment" yaml:"environment" env:"ENVIRONMENT"` // Application environment
	Logger      Logger `json:"logger" yaml:"logger" env:"LOGGER"`                // Logger configuration
	Test        Test   `json:"test" yaml:"test" env:"TEST"`                      // Test configuration
}

// Logger holds configuration for the application logger.
//
// Fields:
//   - Level: Log level (debug, info, warn, error)
//   - Format: Log format (json, text)
//   - Output: Output destination (stdout, file, etc.)
//   - AddSource: Whether to include source file and line number in logs
type Logger struct {
	Level     string `json:"level" yaml:"level" env:"SLOG_LEVEL"`                // Log level
	Format    string `json:"format" yaml:"format" env:"SLOG_FORMAT"`             // Log format
	Output    string `json:"output" yaml:"output" env:"SLOG_OUTPUT"`             // Output destination
	AddSource bool   `json:"add_source" yaml:"add_source" env:"SLOG_ADD_SOURCE"` // Include source info
}

// Test holds configuration for test-related settings.
//
// Fields:
//   - Label_default: Default test label
//   - Label_env: Test label from environment
//   - Label_override: Test label from override
type Test struct {
	Label_default  string `json:"label_def" yaml:"label_def" env:"TEST_LABEL_DEF"`                // Default test label
	Label_env      string `json:"label_env" yaml:"label_env" env:"TEST_LABEL_ENV"`                // Test label from env
	Label_override string `json:"label_override" yaml:"label_override" env:"TEST_LABEL_OVERRIDE"` // Test label from override
}

// NewDefaultConfig returns a Config struct with default values.
//
// Parameters:
//   - logger: slog.Logger for debug logging during initialization
//
// Returns:
//   - *Config: pointer to a Config struct with default values
//
// Example:
//
//	cfg := config.NewDefaultConfig(logger)
//
// Complexity:
//
//	Time: O(1), Space: O(1)
func NewDefaultConfig(logger slog.Logger) *Config {
	logger.Debug("🔠 Initializing default configuration")
	return &Config{
		Environment: "development", // Default to development environment
		Logger: Logger{
			Level:     "debug",  // Default log level
			Format:    "text",   // Default log format
			Output:    "stdout", // Default output to standard output
			AddSource: true,     // Include source file and line number in logs
		},
	}
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
func NewConfig() (*Config, error) {
	slog.Debug("🔠 Loading configuration")

	// Load environment variable to determine the current environment
	var environment string = os.Getenv("ENVIRONMENT")
	slog.Debug("🔠 Environment variable loaded",
		slog.String("environment", environment),
	)

	// Validate the environment variable and set the environment
	// If the environment variable is not set or invalid, default to "development"
	switch environment {
	case "development", "test", "staging", "production":
		slog.Debug("🔠 Valid environment detected",
			slog.String("environment", environment),
		)
	default:
		slog.Warn("🔠 Invalid or unset environment variable, defaulting to development",
			slog.String("environment", environment),
		)
		environment = "development"
	}
	slog.Debug("🔠 Setting environment",
		slog.String("environment", environment),
	)

	// Define config file paths and .env file
	const config_folder = "configs/"
	const config_file_name = "config"
	const config_file_extension = ".json"
	const config_local_string = "local"
	base_config_file := config_folder + config_file_name + config_file_extension
	environment_config_file := config_folder + config_file_name + "." + environment + config_file_extension
	local_config_file := config_folder + config_file_name + "." + config_local_string + config_file_extension
	const dotenv_file = ".env"

	slog.Debug("🔠 Configuration file paths",
		slog.String("base", base_config_file),
		slog.String("environment", environment_config_file),
		slog.String("local", local_config_file),
		slog.String("dotenv", dotenv_file),
	)

	var cfg Config

	// Load base configuration
	if err := loadFromJSONFile(base_config_file, &cfg); err != nil {
		slog.Error("🔠 Error loading default configuration",
			slog.String("file", base_config_file),
			slog.Any("error", err),
		)
		return nil, err
	}
	// Load environment-specific configuration
	if err := loadFromJSONFile(environment_config_file, &cfg); err != nil {
		slog.Error("🔠 Error loading environment configuration",
			slog.String("file", environment_config_file),
			slog.Any("error", err),
		)
		return nil, err
	}
	// Load local configuration (optional, for developer overrides)
	if err := loadFromJSONFile(local_config_file, &cfg); err != nil {
		slog.Error("🔠 Error loading local configuration",
			slog.String("file", local_config_file),
			slog.Any("error", err),
		)
		return nil, err
	}
	// Override config values from environment variables and .env file
	if err := overrideFromEnv(dotenv_file, &cfg); err != nil {
		slog.Error("🔠 Error overriding configuration from environment variables and .env file",
			slog.String("file", dotenv_file),
			slog.Any("error", err),
		)
		return nil, err
	}

	// Log the loaded configuration for debugging and traceability
	slog.Debug("🔠 Configuration loaded",
		slog.String("environment", cfg.Environment),
		slog.String("log_level", cfg.Logger.Level),
		slog.String("log_format", cfg.Logger.Format),
		slog.String("log_output", cfg.Logger.Output),
		slog.Bool("log_add_source", cfg.Logger.AddSource),
	)

	return &cfg, nil
}

// loadFromJSONFile loads configuration values from a JSON file into the provided Config struct.
//
// Parameters:
//   - filePath: Path to the JSON configuration file
//   - cfg: Pointer to the Config struct to populate
//
// Returns:
//   - error: error if the file does not exist or cannot be parsed
//
// Example:
//
//	err := loadFromJSONFile("configs/config.json", &cfg)
//
// Complexity:
//
//	Time: O(n) where n is the file size, Space: O(1)
func loadFromJSONFile(filePath string, cfg *Config) error {
	// Log the attempt to load the configuration file
	slog.Debug("💾 Loading configuration from JSON file",
		slog.String("file", filePath),
	)
	// Check if the file exists before attempting to read
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Error("‼️ Configuration file not found",
			slog.String("file", filePath),
		)
		return fmt.Errorf("config file not found: %s", filePath)
	}
	slog.Debug("💾 Configuration file found",
		slog.String("file", filePath),
	)

	// Read the file contents
	slog.Debug("💾 Reading configuration file",
		slog.String("file", filePath),
	)
	data, err := os.ReadFile(filePath)
	if err != nil {
		slog.Error("‼️ Error reading configuration file",
			slog.String("file", filePath),
			slog.Any("error", err),
		)
		return fmt.Errorf("error reading config file: %s", filePath)
	}
	slog.Debug("💾 Configuration file read successfully",
		slog.String("file", filePath),
	)

	// Parse the JSON data into the Config struct
	slog.Debug("💾 Parsing configuration file",
		slog.String("file", filePath),
	)
	if err := json.Unmarshal(data, cfg); err != nil {
		slog.Error("‼️ Error parsing configuration file",
			slog.String("file", filePath),
			slog.Any("error", err),
		)
		return fmt.Errorf("error parsing config file: %s", filePath)
	}
	slog.Debug("💾 Configuration file parsed successfully",
		slog.String("file", filePath),
		slog.Any("config", *cfg),
	)

	return nil
}

// overrideFromEnv overrides configuration values in cfg from environment variables and .env file.
//
// This function uses the `env` struct tag to determine which environment variable to look for.
// If the environment variable is not set, it looks for the value in the .env file.
// Supports string and bool types by default; can be extended for int, float, etc.
//
// Parameters:
//   - dotenvFile: Path to the .env file
//   - cfg: Pointer to the Config struct to override
//
// Returns:
//   - error: error if .env file cannot be read (non-fatal, logs warning)
//
// Example:
//
//	err := overrideFromEnv(".env", &cfg)
//
// Complexity:
//
//	Time: O(n) where n is the number of struct fields, Space: O(1)
func overrideFromEnv(dotenvFile string, cfg *Config) error {
	// Log the start of the override process
	slog.Debug("🔠 Overriding configuration from environment variables and .env file")

	// Load .env file into a map for lookup
	slog.Debug("🔠 Reading .env file to map",
		slog.String("file", dotenvFile),
	)
	dotenvMap, err := godotenv.Read(dotenvFile)
	if err != nil {
		// Log a warning if .env file cannot be read, but continue using environment variables
		slog.Warn("⚠️ Could not read .env file, will only use environment variables",
			slog.String("file", dotenvFile),
			slog.Any("error", err),
		)
	}

	// Helper function to set a struct field from a string value, handling type conversion
	setField := func(field reflect.Value, value string) {
		switch field.Kind() {
		case reflect.String:
			// Set string field directly
			field.SetString(value)
			slog.Debug("🔠 Set string field",
				slog.String("value", value),
			)
		case reflect.Bool:
			// Parse and set boolean field
			b, err := strconv.ParseBool(value)
			if err == nil {
				field.SetBool(b)
				slog.Debug("🔠 Set bool field",
					slog.Bool("value", b),
				)
			} else {
				slog.Warn("⚠️ Could not parse bool value",
					slog.String("value", value),
				)
			}
		// Add more types as needed (int, float, etc.)
		default:
			// Log a warning for unsupported field types
			slog.Warn("⚠️ Unsupported field type for override",
				slog.String("kind", field.Kind().String()),
			)
		}
	}

	// Recursive function to process struct fields, including nested structs
	var processStruct func(reflect.Value, reflect.Type)
	processStruct = func(v reflect.Value, t reflect.Type) {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			valField := v.Field(i)
			envTag := field.Tag.Get("env")
			// Only process fields with an 'env' tag and that can be set
			if envTag != "" && valField.CanSet() {
				// Try to get the value from the environment variable
				val := os.Getenv(envTag)
				source := "env"
				// If not found in environment, try the .env file
				if val == "" {
					val = dotenvMap[envTag]
					source = ".env"
				}
				// If a value was found, set the field and log the action
				if val != "" {
					slog.Debug("🔠 Overriding field from "+source,
						slog.String("field", field.Name),
						slog.String("env_tag", envTag),
						slog.String("value", val),
					)
					setField(valField, val)
				} else {
					// Log if no override was found for this field
					slog.Debug("🔠 No override found for field",
						slog.String("field", field.Name),
						slog.String("env_tag", envTag),
					)
				}
			}
			// Recursively process nested structs
			if valField.Kind() == reflect.Struct {
				processStruct(valField, field.Type)
			}
		}
	}

	// Start processing from the root config struct
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	slog.Debug("🔠 Processing struct for overrides",
		slog.String("type", t.Name()),
	)
	processStruct(v, t)

	slog.Debug("🔠 Override from environment and .env file complete")
	return nil
}
