// wails_logger_test.go verifies Wails logger adapter behavior.
// internal/core/adapters/logger/wails_logger_test.go
package logger

import (
	"testing"

	"github.com/MadeByDoug/wls-chatbot/pkg/zerologtest"
	"github.com/rs/zerolog"
)

func TestWailsLoggerLevels(t *testing.T) {

	tests := []struct {
		name          string
		logFunc       func(l *WailsLogger, msg string)
		expectedLevel string
		message       string
	}{
		{
			name: "Print",
			logFunc: func(l *WailsLogger, msg string) {
				l.Print(msg)
			},
			expectedLevel: "info",
			message:       "test print",
		},
		{
			name: "Trace",
			logFunc: func(l *WailsLogger, msg string) {
				l.Trace(msg)
			},
			expectedLevel: "trace",
			message:       "test trace",
		},
		{
			name: "Debug",
			logFunc: func(l *WailsLogger, msg string) {
				l.Debug(msg)
			},
			expectedLevel: "debug",
			message:       "test debug",
		},
		{
			name: "Info",
			logFunc: func(l *WailsLogger, msg string) {
				l.Info(msg)
			},
			expectedLevel: "info",
			message:       "test info",
		},
		{
			name: "Warning",
			logFunc: func(l *WailsLogger, msg string) {
				l.Warning(msg)
			},
			expectedLevel: "warn",
			message:       "test warning",
		},
		{
			name: "Error",
			logFunc: func(l *WailsLogger, msg string) {
				l.Error(msg)
			},
			expectedLevel: "error",
			message:       "test error",
		},
		{
			name: "Fatal",
			logFunc: func(l *WailsLogger, msg string) {
				l.Fatal(msg)
			},
			expectedLevel: "error", // Wails Fatal maps to zerolog Error to prevent app exit
			message:       "test fatal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := zerologtest.NewRecorder()
			logger := NewWailsLogger(zerolog.New(recorder))

			tt.logFunc(logger, tt.message)

			entry, ok := recorder.Last()
			if !ok {
				t.Fatalf("expected log entry")
			}

			if entry.Level() != tt.expectedLevel {
				t.Errorf("expected level %q, got %q", tt.expectedLevel, entry.Level())
			}

			if entry.Message() != tt.message {
				t.Errorf("expected message %q, got %q", tt.message, entry.Message())
			}

			if val, ok := entry.StringField("source"); !ok || val != "wails" {
				t.Errorf("expected source=wails, got %q", val)
			}
		})
	}
}

func TestWailsLoggerIgnoresEmptyMessages(t *testing.T) {

	recorder := zerologtest.NewRecorder()
	logger := NewWailsLogger(zerolog.New(recorder))

	logger.Info("")
	logger.Info("   ")

	if len(recorder.Entries()) > 0 {
		t.Errorf("expected 0 entries, got %d", len(recorder.Entries()))
	}
}
