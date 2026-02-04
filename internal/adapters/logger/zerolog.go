// configure and adapt zerolog for core logging.
// internal/adapters/logger/zerolog.go
package logger

import (
	"os"
	"time"

	"github.com/MadeByDoug/wls-chatbot/internal/core/ports"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// New creates a new configured zerolog.Logger.
// It defaults to writing to a ConsoleWriter for human-readable output during development.
// This is compatible with Windows, Linux, and macOS.
func New(level string) zerolog.Logger {

	// Set global log level
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Use ConsoleWriter for nicer development logs.
	// Out defaults to os.Stderr which is standard.
	// TimeFormat conforms to standard convenient reading.
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.TimeOnly,
	}

	// Create a new logger with timestamp and caller information
	l := log.Output(output).With().Timestamp().Caller().Logger()

	return l
}

// Adapter bridges zerolog into the core Logger port.
type Adapter struct {
	logger zerolog.Logger
}

// NewAdapter creates a logger adapter for core use cases.
func NewAdapter(l zerolog.Logger) *Adapter {
	return &Adapter{logger: l}
}

// Trace logs a trace message.
func (a *Adapter) Trace(message string, fields ...ports.LogField) {
	if a == nil {
		return
	}
	event := a.logger.Trace()
	writeFields(event, fields)
	event.Msg(message)
}

// Debug logs a debug message.
func (a *Adapter) Debug(message string, fields ...ports.LogField) {
	if a == nil {
		return
	}
	event := a.logger.Debug()
	writeFields(event, fields)
	event.Msg(message)
}

// Info logs an info message.
func (a *Adapter) Info(message string, fields ...ports.LogField) {
	if a == nil {
		return
	}
	event := a.logger.Info()
	writeFields(event, fields)
	event.Msg(message)
}

// Warn logs a warning message with an optional error.
func (a *Adapter) Warn(message string, err error, fields ...ports.LogField) {
	if a == nil {
		return
	}
	event := a.logger.Warn()
	if err != nil {
		event = event.Err(err)
	}
	writeFields(event, fields)
	event.Msg(message)
}

// Error logs an error message with an optional error.
func (a *Adapter) Error(message string, err error, fields ...ports.LogField) {

	if a == nil {
		return
	}
	event := a.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	writeFields(event, fields)
	event.Msg(message)
}

func writeFields(event *zerolog.Event, fields []ports.LogField) {

	for _, field := range fields {
		event.Str(field.Key, field.Value)
	}
}
