package kitlog

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
)

// KitLogger is a simplified Logger implementation
type KitLogger struct {
	output io.Writer
	level  string
	ctx    map[string]interface{}
}

// NewKitLogger creates a new instance of KitLogger
func NewKitLogger(cfg config.LoggingConfig) (logging.Logger, error) {
	var output io.Writer = os.Stdout

	return &KitLogger{
		output: output,
		level:  strings.ToLower(cfg.Level),
		ctx:    make(map[string]interface{}),
	}, nil
}

// Debug logs a debug message
func (l *KitLogger) Debug(msg string, keyvals ...interface{}) {
	if l.level != "debug" {
		return
	}
	l.log("DEBUG", msg, keyvals...)
}

// Info logs an info message
func (l *KitLogger) Info(msg string, keyvals ...interface{}) {
	if l.level != "debug" && l.level != "info" {
		return
	}
	l.log("INFO", msg, keyvals...)
}

// Warn logs a warning message
func (l *KitLogger) Warn(msg string, keyvals ...interface{}) {
	if l.level != "debug" && l.level != "info" && l.level != "warn" {
		return
	}
	l.log("WARN", msg, keyvals...)
}

// Error logs an error message
func (l *KitLogger) Error(msg string, keyvals ...interface{}) {
	l.log("ERROR", msg, keyvals...)
}

// With returns a new Logger with the provided keyvals
func (l *KitLogger) With(keyvals ...interface{}) logging.Logger {
	// Ensure even number of keyvals
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "MISSING_VALUE")
	}

	// Create a new context map with existing and new values
	newCtx := make(map[string]interface{}, len(l.ctx)+len(keyvals)/2)
	for k, v := range l.ctx {
		newCtx[k] = v
	}

	// Add new key-value pairs
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}
		newCtx[key] = keyvals[i+1]
	}

	return &KitLogger{
		output: l.output,
		level:  l.level,
		ctx:    newCtx,
	}
}

// log formats and outputs a log message
func (l *KitLogger) log(level, msg string, keyvals ...interface{}) {
	// Create a map for all values
	logMap := make(map[string]interface{})

	// Add timestamp, level and message
	logMap["ts"] = time.Now().Format(time.RFC3339)
	logMap["level"] = level
	logMap["msg"] = msg

	// Add context values
	for k, v := range l.ctx {
		logMap[k] = v
	}

	// Add additional keyvals
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key, ok := keyvals[i].(string)
			if !ok {
				key = fmt.Sprintf("%v", keyvals[i])
			}
			logMap[key] = keyvals[i+1]
		}
	}

	// Simple implementation that outputs key-value pairs
	fmt.Fprintf(l.output, "%v %v: %v", logMap["ts"], logMap["level"], logMap["msg"])

	// Output additional fields
	for k, v := range logMap {
		if k != "ts" && k != "level" && k != "msg" {
			fmt.Fprintf(l.output, " %v=%v", k, v)
		}
	}

	fmt.Fprintln(l.output)
}
