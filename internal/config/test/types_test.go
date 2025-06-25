package config_test

import (
	"reflect"
	"testing"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
)

func TestConfigStructTags(t *testing.T) {
	cfgType := reflect.TypeOf(config.Config{})
	loggerType := reflect.TypeOf(config.Logger{})
	serverType := reflect.TypeOf(config.Server{})
	testType := reflect.TypeOf(config.Test{})

	// Check Config struct tags
	if _, ok := cfgType.FieldByName("Environment"); !ok {
		t.Error("Config struct missing 'Environment' field")
	}
	if _, ok := cfgType.FieldByName("Logger"); !ok {
		t.Error("Config struct missing 'Logger' field")
	}
	if _, ok := cfgType.FieldByName("Server"); !ok {
		t.Error("Config struct missing 'Server' field")
	}
	if _, ok := cfgType.FieldByName("Test"); !ok {
		t.Error("Config struct missing 'Test' field")
	}

	// Check Logger struct tags
	if tag := loggerType.Field(0).Tag.Get("json"); tag != "level" {
		t.Errorf("Logger.Level json tag = %s, want 'level'", tag)
	}
	if tag := loggerType.Field(1).Tag.Get("json"); tag != "format" {
		t.Errorf("Logger.Format json tag = %s, want 'format'", tag)
	}
	if tag := loggerType.Field(2).Tag.Get("json"); tag != "output" {
		t.Errorf("Logger.Output json tag = %s, want 'output'", tag)
	}
	if tag := loggerType.Field(3).Tag.Get("json"); tag != "add_source" {
		t.Errorf("Logger.AddSource json tag = %s, want 'add_source'", tag)
	}

	// Check Server struct tags
	if tag := serverType.Field(0).Tag.Get("json"); tag != "port" {
		t.Errorf("Server.Port json tag = %s, want 'port'", tag)
	}
	if tag := serverType.Field(1).Tag.Get("json"); tag != "host" {
		t.Errorf("Server.Host json tag = %s, want 'host'", tag)
	}
	if tag := serverType.Field(2).Tag.Get("json"); tag != "read_timeout" {
		t.Errorf("Server.ReadTimeout json tag = %s, want 'read_timeout'", tag)
	}
	if tag := serverType.Field(3).Tag.Get("json"); tag != "write_timeout" {
		t.Errorf("Server.WriteTimeout json tag = %s, want 'write_timeout'", tag)
	}
	if tag := serverType.Field(4).Tag.Get("json"); tag != "shutdown_timeout" {
		t.Errorf("Server.ShutdownTimeout json tag = %s, want 'shutdown_timeout'", tag)
	}

	// Check Test struct tags
	if tag := testType.Field(0).Tag.Get("json"); tag != "label_def" {
		t.Errorf("Test.Label_default json tag = %s, want 'label_def'", tag)
	}
	if tag := testType.Field(1).Tag.Get("json"); tag != "label_env" {
		t.Errorf("Test.Label_env json tag = %s, want 'label_env'", tag)
	}
	if tag := testType.Field(2).Tag.Get("json"); tag != "label_override" {
		t.Errorf("Test.Label_override json tag = %s, want 'label_override'", tag)
	}
}
