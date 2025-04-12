package logging

import (
	"fmt"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
)

var defaultLogger Logger = &NoopLogger{}

// SetDefaultLogger sets the default logger instance
func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

// GetDefaultLogger returns the default logger instance
func GetDefaultLogger() Logger {
	return defaultLogger
}

// NewLogger creates a new logger based on the configuration
func NewLogger(cfg config.LoggingConfig, impl string) (Logger, error) {
	// The actual implementation will be in a subpackage like kitlog or zerolog
	// This provides a layer of indirection so we can swap implementations
	switch impl {
	case "kit", "kitlog", "":
		// We're returning NoopLogger here to avoid circular dependencies
		// The actual implementation should use kitlog.NewKitLogger directly
		return &NoopLogger{}, nil
	default:
		return nil, fmt.Errorf("unknown logger implementation: %s", impl)
	}
}
