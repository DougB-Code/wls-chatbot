// bridge frontend log entries into zerolog.
// internal/adapters/logger/log_bridge.go
package logger

import (
	"github.com/rs/zerolog"
)

// LogEntry is a strictly typed log structure for the frontend bridge.
type LogEntry struct {
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

// Logger acts as a bridge for frontend logging.
type Logger struct {
	logger zerolog.Logger
}

// NewLogBridge creates a new Logger bridge.
func NewLogBridge(l zerolog.Logger) *Logger {

	return &Logger{
		logger: l,
	}
}

// Log logs an entry from the frontend.
func (b *Logger) Log(entry LogEntry) {

	// Create a sub-logger event based on the level
	var event *zerolog.Event
	switch entry.Level {
	case "trace":
		event = b.logger.Trace()
	case "debug":
		event = b.logger.Debug()
	case "info":
		event = b.logger.Info()
	case "warn":
		event = b.logger.Warn()
	case "error":
		event = b.logger.Error()
	case "fatal":
		event = b.logger.Error()
	case "panic":
		event = b.logger.Error()
	default:
		event = b.logger.Info()
	}

	// Attach fields
	for k, v := range entry.Fields {
		event.Str(k, v)
	}
	if entry.Level == "fatal" || entry.Level == "panic" {
		event.Str("frontend_level", entry.Level)
	}

	// Log with message
	event.Msg(entry.Message)
}
