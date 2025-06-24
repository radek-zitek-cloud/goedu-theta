package config

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"reflect"
	"strconv"
	"log/slog"
	"os"
)

type Config struct {
	// Environment variable to determine if the application is in production mode
	Environment string `json:"environment" yaml:"environment" env:"ENVIRONMENT"`
	Logger      Logger `json:"logger" yaml:"logger" env:"LOGGER"`
	Test        Test   `json:"test" yaml:"test" env:"TEST"`
}

type Logger struct {
	// Level indicates the log level (e.g., Debug, Info, Error)
	Level string `json:"level" yaml:"level" env:"SLOG_LEVEL"`
	// Format indicates the log format (e.g., JSON, Text)
	Format string `json:"format" yaml:"format" env:"SLOG_FORMAT"`
	// Output indicates where to output the logs (e.g., stdout, file)
	Output string `json:"output" yaml:"output" env:"SLOG_OUTPUT"`
	// AddSource indicates whether to include source file and line number in logs
	AddSource bool `json:"add_source" yaml:"add_source" env:"SLOG_ADD_SOURCE"`
}

type Test struct {
	Label_default  string `json:"label_def" yaml:"label_def" env:"TEST_LABEL_DEF"`
	Label_env      string `json:"label_env" yaml:"label_env" env:"TEST_LABEL_ENV"`
	Label_override string `json:"label_override" yaml:"label_override" env:"TEST_LABEL_OVERRIDE"`
}

func NewDefaultConfig(logger slog.Logger) *Config {
	logger.Debug("🔠 Initializing default configuration")
	return &Config{
		Environment: "development", // Default to development environment
		Logger: Logger{
			Level:      "debug",    // Default log level
			Format:     "text",     // Default log format
			Output:     "stdout",   // Default output to standard output
			AddSource:  true,       // Include source file and line number in logs
		},
	}
}

func NewConfig(logger slog.Logger) (*Config, error) {
	logger.Debug("🔠 Loading configuration")

	// Load environment variables and override default configuration
	var environment string = os.Getenv("ENVIRONMENT")
	logger.Debug("🔠 Environment variable loaded",
		slog.String("environment", environment),
	)
	// Validate the environment variable and set the environment
	// If the environment variable is not set or invalid, default to "development"
	switch environment {
	case "development", "test", "staging", "production":
		logger.Debug("🔠 Valid environment detected",
			slog.String("environment", environment),
		)
	default:
		logger.Warn("🔠 Invalid or unset environment variable, defaulting to development",
			slog.String("environment", environment),
		)
		environment = "development"
	}
	logger.Debug("🔠 Setting environment",
		slog.String("environment", environment),
	)

	const config_folder = "configs/"
	const config_file_name = "config"
	const config_file_extension = ".json"
	const config_local_string = "local"
	// Construct the path to the configuration file 
	base_config_file := config_folder + config_file_name + config_file_extension
	// Construct the path to the environment-specific configuration file
	environment_config_file := config_folder + config_file_name + "." + environment + config_file_extension
	// Construct the path to the local configuration file
	local_config_file := config_folder + config_file_name + "." + config_local_string + config_file_extension
	// Construct the path to the dotenv file
	const dotenv_file = ".env"

	logger.Debug("🔠 Configuration file paths",
		slog.String("base", base_config_file),
		slog.String("environment", environment_config_file),
		slog.String("local", local_config_file),
		slog.String("dotenv", dotenv_file),
	)

	var cfg Config

	if err := loadFromJSONFile(base_config_file, &cfg); err != nil {
		logger.Error("🔠 Error loading default configuration",
			slog.String("file", base_config_file),
			slog.Any("error", err),
		)
		return nil, err
	}
	if err := loadFromJSONFile(environment_config_file, &cfg); err != nil {
		logger.Error("🔠 Error loading environment configuration",
			slog.String("file", environment_config_file),
			slog.Any("error", err),
		)
		return nil, err
	}
	if err := loadFromJSONFile(local_config_file, &cfg); err != nil {
		logger.Error("🔠 Error loading local configuration",
			slog.String("file", local_config_file),
			slog.Any("error", err),
		)
		return nil, err
	}
	if err := overrideFromEnv(dotenv_file, &cfg); err != nil {
		logger.Error("🔠 Error overriding configuration from environment variables and .env file",
			slog.String("file", dotenv_file),
			slog.Any("error", err),
		)
		return nil, err
	}

	// Log the loaded configuration
	logger.Debug("🔠 Configuration loaded",
		slog.String("environment", cfg.Environment),
		slog.String("log_level", cfg.Logger.Level),
		slog.String("log_format", cfg.Logger.Format),
		slog.String("log_output", cfg.Logger.Output),
		slog.Bool("log_add_source", cfg.Logger.AddSource),
	)

	return &cfg, nil
}

// Load configuration structure from a JSON file
func loadFromJSONFile(filePath string, cfg *Config) (error) {
    // Check if file exists
	slog.Debug("💾 Loading configuration from JSON file",
		slog.String("file", filePath),
	)
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Error("‼️ Configuration file not found",
			slog.String("file", filePath),
		)
        return fmt.Errorf("config file not found: %s", filePath)
    }
	slog.Debug("💾 Configuration file found",
		slog.String("file", filePath),
	)

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

// Override configuration values from environment variables and .env file
func overrideFromEnv(dotenvFile string, cfg *Config) error {
    // Log the start of the override process
    slog.Debug("🔠 Overriding configuration from environment variables and .env file")

    // Load .env file into a map for lookup
    slog.Debug("🔠 Reading .env file to map",
        slog.String("file", dotenvFile),
    )
    dotenvMap, err := godotenv.Read(dotenvFile)
    if err != nil {
        slog.Warn("⚠️ Could not read .env file, will only use environment variables",
            slog.String("file", dotenvFile),
            slog.Any("error", err),
        )
    }

    // Helper function to set a struct field from a string value, handling type conversion
    setField := func(field reflect.Value, value string) {
        switch field.Kind() {
        case reflect.String:
            field.SetString(value)
            slog.Debug("🔠 Set string field",
                slog.String("value", value),
            )
        case reflect.Bool:
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

