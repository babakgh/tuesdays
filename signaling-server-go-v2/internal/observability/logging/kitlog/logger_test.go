package kitlog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
)

func TestNewKitLogger(t *testing.T) {
	// Create a config
	cfg := config.LoggingConfig{
		Level:      "debug",
		Format:     "json",
		TimeFormat: "RFC3339",
	}

	// Create a logger
	logger, err := NewKitLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Verify logger is not nil
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// Type assertion
	_, ok := logger.(*KitLogger)
	if !ok {
		t.Error("Logger should be a *KitLogger")
	}
}

func TestLogLevels(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a logger with the buffer
	logger := &KitLogger{
		output: &buf,
		level:  "info",
		ctx:    make(map[string]interface{}),
	}

	// Test debug level (should be filtered)
	buf.Reset()
	logger.Debug("Debug message")
	if buf.Len() > 0 {
		t.Errorf("Debug message was logged when level is info: %s", buf.String())
	}

	// Test info level
	buf.Reset()
	logger.Info("Info message")
	output := buf.String()
	if !strings.Contains(output, "INFO") || !strings.Contains(output, "Info message") {
		t.Errorf("Expected info message to be logged, got: %s", output)
	}

	// Test warn level
	buf.Reset()
	logger.Warn("Warn message")
	output = buf.String()
	if !strings.Contains(output, "WARN") || !strings.Contains(output, "Warn message") {
		t.Errorf("Expected warn message to be logged, got: %s", output)
	}

	// Test error level
	buf.Reset()
	logger.Error("Error message")
	output = buf.String()
	if !strings.Contains(output, "ERROR") || !strings.Contains(output, "Error message") {
		t.Errorf("Expected error message to be logged, got: %s", output)
	}
}

func TestLoggerContextWith(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a logger with the buffer
	logger := &KitLogger{
		output: &buf,
		level:  "debug",
		ctx:    make(map[string]interface{}),
	}

	// Add context with With
	contextLogger := logger.With("key1", "value1", "key2", 42)

	// Verify it's a KitLogger
	ctxLogger, ok := contextLogger.(*KitLogger)
	if !ok {
		t.Fatal("Contextual logger should be a *KitLogger")
	}

	// Verify context was added
	if ctxLogger.ctx["key1"] != "value1" {
		t.Errorf("Expected key1=value1 in context, got %v", ctxLogger.ctx["key1"])
	}

	if ctxLogger.ctx["key2"] != 42 {
		t.Errorf("Expected key2=42 in context, got %v", ctxLogger.ctx["key2"])
	}

	// Log with context logger
	buf.Reset()
	ctxLogger.Info("Test with context")
	output := buf.String()

	// Verify context values are in the output
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("Expected key1=value1 in log output, got: %s", output)
	}

	if !strings.Contains(output, "key2=42") {
		t.Errorf("Expected key2=42 in log output, got: %s", output)
	}
}

func TestWithOddKeyvals(t *testing.T) {
	// Create a logger
	logger := &KitLogger{
		output: &bytes.Buffer{},
		level:  "debug",
		ctx:    make(map[string]interface{}),
	}

	// Test With() with odd number of keyvals
	ctxLogger := logger.With("key1", "value1", "orphan")

	// Verify context was properly handled
	l, ok := ctxLogger.(*KitLogger)
	if !ok {
		t.Fatal("Expected *KitLogger")
	}

	// Check that orphan key got a MISSING_VALUE
	if l.ctx["orphan"] != "MISSING_VALUE" {
		t.Errorf("Expected orphan key to have MISSING_VALUE, got %v", l.ctx["orphan"])
	}
}
