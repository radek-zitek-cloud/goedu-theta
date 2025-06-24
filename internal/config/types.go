package config

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
