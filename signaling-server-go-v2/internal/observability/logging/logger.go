package logging

// Logger interface for abstracting logging implementations
type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	With(keyvals ...interface{}) Logger
}

// NoopLogger is a logger implementation that does nothing
type NoopLogger struct{}

// Debug implements Logger.Debug
func (l *NoopLogger) Debug(msg string, keyvals ...interface{}) {}

// Info implements Logger.Info
func (l *NoopLogger) Info(msg string, keyvals ...interface{}) {}

// Warn implements Logger.Warn
func (l *NoopLogger) Warn(msg string, keyvals ...interface{}) {}

// Error implements Logger.Error
func (l *NoopLogger) Error(msg string, keyvals ...interface{}) {}

// With implements Logger.With
func (l *NoopLogger) With(keyvals ...interface{}) Logger {
	return l
}
