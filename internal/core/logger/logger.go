// define structured logging contracts for core use cases.
// internal/core/interfaces/logger.go
package logger

// LogField represents a structured logging field.
type LogField struct {
	Key   string
	Value string
}

// Logger provides structured logging for use cases.
type Logger interface {
	// Trace logs a trace message.
	Trace(message string, fields ...LogField)
	// Debug logs a debug message.
	Debug(message string, fields ...LogField)
	// Info logs an info message.
	Info(message string, fields ...LogField)
	// Warn logs a warning message with an optional error.
	Warn(message string, err error, fields ...LogField)
	// Error logs an error message with an optional error.
	Error(message string, err error, fields ...LogField)
}
