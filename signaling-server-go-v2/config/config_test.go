package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test with default values
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check some default values
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default logging level 'info', got %s", cfg.Logging.Level)
	}
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("LOGGING_LEVEL", "debug")
	os.Setenv("METRICS_ENABLED", "false")

	// Load configuration
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check values are set from environment variables
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090 from env var, got %d", cfg.Server.Port)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected logging level 'debug' from env var, got %s", cfg.Logging.Level)
	}
	if cfg.Metrics.Enabled != false {
		t.Errorf("Expected metrics enabled 'false' from env var, got %t", cfg.Metrics.Enabled)
	}

	// Clean up
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("LOGGING_LEVEL")
	os.Unsetenv("METRICS_ENABLED")
}

func TestGetConfigPath(t *testing.T) {
	// Test without environment variable
	originalPath := os.Getenv("SERVER_CONFIG_PATH")
	os.Unsetenv("SERVER_CONFIG_PATH")
	path := GetConfigPath()
	
	// With no env var and no default file, it should return empty
	// or try to find the default config file
	if path != "" && path != "config/default.yaml" {
		t.Errorf("Expected empty path or default path, got: %s", path)
	}

	// Test with environment variable
	os.Setenv("SERVER_CONFIG_PATH", "test/config.yaml")
	path = GetConfigPath()
	if path != "test/config.yaml" {
		t.Errorf("Expected 'test/config.yaml', got: %s", path)
	}

	// Clean up
	if originalPath != "" {
		os.Setenv("SERVER_CONFIG_PATH", originalPath)
	} else {
		os.Unsetenv("SERVER_CONFIG_PATH")
	}
}