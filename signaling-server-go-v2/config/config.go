package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds all configuration for the server
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	Tracing    TracingConfig    `mapstructure:"tracing"`
	WebSocket  WebSocketConfig  `mapstructure:"websocket"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

// ServerConfig holds HTTP server related configuration
type ServerConfig struct {
	Port            int    `mapstructure:"port"`
	Host            string `mapstructure:"host"`
	ShutdownTimeout int    `mapstructure:"shutdownTimeout"` // in seconds
	ReadTimeout     int    `mapstructure:"readTimeout"`     // in seconds
	WriteTimeout    int    `mapstructure:"writeTimeout"`    // in seconds
	IdleTimeout     int    `mapstructure:"idleTimeout"`     // in seconds
}

// LoggingConfig holds logging related configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	TimeFormat string `mapstructure:"timeFormat"`
}

// MetricsConfig holds Prometheus metrics related configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// TracingConfig holds OpenTelemetry tracing related configuration
type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Exporter    string `mapstructure:"exporter"`
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"serviceName"`
}

// WebSocketConfig holds WebSocket related configuration
type WebSocketConfig struct {
	Path           string `mapstructure:"path"`
	PingInterval   int    `mapstructure:"pingInterval"`    // in seconds
	PongWait       int    `mapstructure:"pongWait"`        // in seconds
	WriteWait      int    `mapstructure:"writeWait"`       // in seconds
	MaxMessageSize int64  `mapstructure:"maxMessageSize"`  // in bytes
}

// MonitoringConfig holds health checking related configuration
type MonitoringConfig struct {
	LivenessPath  string `mapstructure:"livenessPath"`
	ReadinessPath string `mapstructure:"readinessPath"`
}

// LoadConfig loads the configuration from environment variables and returns defaults for missing values
func LoadConfig(configPath string) (*Config, error) {
	// Create a default configuration
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnvInt("SERVER_PORT", 8080),
			Host:            getEnvString("SERVER_HOST", "0.0.0.0"),
			ShutdownTimeout: getEnvInt("SERVER_SHUTDOWN_TIMEOUT", 30),
			ReadTimeout:     getEnvInt("SERVER_READ_TIMEOUT", 15),
			WriteTimeout:    getEnvInt("SERVER_WRITE_TIMEOUT", 15),
			IdleTimeout:     getEnvInt("SERVER_IDLE_TIMEOUT", 60),
		},
		Logging: LoggingConfig{
			Level:      getEnvString("LOGGING_LEVEL", "info"),
			Format:     getEnvString("LOGGING_FORMAT", "json"),
			TimeFormat: getEnvString("LOGGING_TIME_FORMAT", "RFC3339"),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvBool("METRICS_ENABLED", true),
			Path:    getEnvString("METRICS_PATH", "/metrics"),
		},
		Tracing: TracingConfig{
			Enabled:     getEnvBool("TRACING_ENABLED", true),
			Exporter:    getEnvString("TRACING_EXPORTER", "otlp"),
			Endpoint:    getEnvString("TRACING_ENDPOINT", "localhost:4317"),
			ServiceName: getEnvString("TRACING_SERVICE_NAME", "signaling-server"),
		},
		WebSocket: WebSocketConfig{
			Path:           getEnvString("WEBSOCKET_PATH", "/ws"),
			PingInterval:   getEnvInt("WEBSOCKET_PING_INTERVAL", 30),
			PongWait:       getEnvInt("WEBSOCKET_PONG_WAIT", 60),
			WriteWait:      getEnvInt("WEBSOCKET_WRITE_WAIT", 10),
			MaxMessageSize: getEnvInt64("WEBSOCKET_MAX_MESSAGE_SIZE", 1024*1024), // 1MB
		},
		Monitoring: MonitoringConfig{
			LivenessPath:  getEnvString("MONITORING_LIVENESS_PATH", "/health/live"),
			ReadinessPath: getEnvString("MONITORING_READINESS_PATH", "/health/ready"),
		},
	}

	// In a real implementation, we would parse a config file here if one was provided
	fmt.Println("No config file found. Using environment variables and defaults.")

	return cfg, nil
}

// GetConfigPath returns the path to the config file specified by the environment variable
func GetConfigPath() string {
	configPath := os.Getenv("SERVER_CONFIG_PATH")
	if configPath == "" {
		// Try to find config file in the config directory
		defaultConfigPath := filepath.Join("config", "default.yaml")
		if _, err := os.Stat(defaultConfigPath); err == nil {
			return defaultConfigPath
		}
	}
	return configPath
}

// Environment variable helpers
func getEnvString(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	value = strings.ToLower(value)
	if value == "true" || value == "1" || value == "yes" || value == "y" {
		return true
	}

	if value == "false" || value == "0" || value == "no" || value == "n" {
		return false
	}

	return defaultValue
}