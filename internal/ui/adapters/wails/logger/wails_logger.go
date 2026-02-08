// wails_logger.go adapts zerolog to the Wails logger interface.
// internal/core/adapters/logger/wails_logger.go
package logger

import (
	"strings"

	"github.com/rs/zerolog"
)

// WailsLogger routes Wails internal logs through zerolog.
type WailsLogger struct {
	logger zerolog.Logger
}

// NewWailsLogger creates a logger adapter for Wails internals.
func NewWailsLogger(l zerolog.Logger) *WailsLogger {

	return &WailsLogger{logger: l}
}

// Print logs a plain Wails message.
func (l *WailsLogger) Print(message string) {

	l.log("print", message)
}

// Trace logs a Wails trace message.
func (l *WailsLogger) Trace(message string) {

	l.log("trace", message)
}

// Debug logs a Wails debug message.
func (l *WailsLogger) Debug(message string) {

	l.log("debug", message)
}

// Info logs a Wails info message.
func (l *WailsLogger) Info(message string) {

	l.log("info", message)
}

// Warning logs a Wails warning message.
func (l *WailsLogger) Warning(message string) {

	l.log("warning", message)
}

// Error logs a Wails error message.
func (l *WailsLogger) Error(message string) {

	l.log("error", message)
}

// Fatal logs a Wails fatal message.
func (l *WailsLogger) Fatal(message string) {

	l.log("fatal", message)
}

// log emits the message at the selected level unless filtered.
func (l *WailsLogger) log(level, message string) {

	if l == nil {
		return
	}

	msg := strings.TrimSpace(message)
	if msg == "" {
		return
	}

	scopedLogger := l.logger.With().
		Str("source", "wails").
		Str("wails_level", level).
		Logger()

	event := scopedLogger.Info()

	switch level {
	case "trace":
		event = scopedLogger.Trace()
	case "debug":
		event = scopedLogger.Debug()
	case "warning":
		event = scopedLogger.Warn()
	case "error", "fatal":
		event = scopedLogger.Error()
	}

	event.Msg(msg)
}

