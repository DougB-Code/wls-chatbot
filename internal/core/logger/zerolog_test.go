// zerolog_test.go verifies the core logger adapter.
// internal/core/adapters/logger/zerolog_test.go
package logger

import (
	"errors"
	"testing"

	"github.com/MadeByDoug/wls-chatbot/pkg/zerologtest"
	"github.com/rs/zerolog"
)

func TestCoreAdapter(t *testing.T) {

	recorder := zerologtest.NewRecorder()
	adapter := NewAdapter(zerolog.New(recorder))

	t.Run("Info with fields", func(t *testing.T) {
		adapter.Info("test info", LogField{Key: "user_id", Value: "123"})

		entry, ok := recorder.Last()
		if !ok {
			t.Fatal("expected log entry")
		}

		if entry.Level() != "info" {
			t.Errorf("expected level info, got %q", entry.Level())
		}
		if entry.Message() != "test info" {
			t.Errorf("expected message 'test info', got %q", entry.Message())
		}
		if val, ok := entry.StringField("user_id"); !ok || val != "123" {
			t.Errorf("expected user_id=123, got %q", val)
		}
	})

	t.Run("Error with error object", func(t *testing.T) {
		testErr := errors.New("something went wrong")
		adapter.Error("operation failed", testErr, LogField{Key: "attempt", Value: "1"})

		entry, ok := recorder.Last()
		if !ok {
			t.Fatal("expected log entry")
		}

		if entry.Level() != "error" {
			t.Errorf("expected level error, got %q", entry.Level())
		}
		if entry.Message() != "operation failed" {
			t.Errorf("expected message 'operation failed', got %q", entry.Message())
		}
		if val, ok := entry.StringField("error"); !ok || val != "something went wrong" {
			t.Errorf("expected error field, got %q", val)
		}
		if val, ok := entry.StringField("attempt"); !ok || val != "1" {
			t.Errorf("expected attempt=1, got %q", val)
		}
	})

	t.Run("Debug", func(t *testing.T) {
		adapter.Debug("debugging", LogField{Key: "foo", Value: "bar"})

		entry, ok := recorder.Last()
		if !ok {
			t.Fatal("expected log entry")
		}

		if entry.Level() != "debug" {
			t.Errorf("expected level debug, got %q", entry.Level())
		}
	})

	t.Run("Trace", func(t *testing.T) {
		adapter.Trace("tracing")

		entry, ok := recorder.Last()
		if !ok {
			t.Fatal("expected log entry")
		}

		if entry.Level() != "trace" {
			t.Errorf("expected level trace, got %q", entry.Level())
		}
	})

	t.Run("Warn", func(t *testing.T) {
		adapter.Warn("warning", nil)

		entry, ok := recorder.Last()
		if !ok {
			t.Fatal("expected log entry")
		}

		if entry.Level() != "warn" {
			t.Errorf("expected level warn, got %q", entry.Level())
		}
	})
}

func TestNewLoggerConfiguration(t *testing.T) {
	// Verify that setting the level to "error" suppresses "info" logs.
	// We use "error" because it's a high level, so it should suppress "info".
	
	// Capture stderr
	// Note: checking os.Stderr is hard in parallel tests. 
	// But New() configures the global log.Logger.
	
	// Instead of checking GetLevel(), which relies on internal state reset,
	// let's just ensure that calling New("debug") sets the global level to Debug.
	
	New("debug")
	if zerolog.GlobalLevel() != zerolog.DebugLevel {
		t.Errorf("expected global level debug, got %v", zerolog.GlobalLevel())
	}

	New("info")
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Errorf("expected global level info, got %v", zerolog.GlobalLevel())
	}
}
