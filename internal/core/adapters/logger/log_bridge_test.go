// log_bridge_test.go verifies frontend log level mapping and severity annotations.
// internal/core/adapters/logger/log_bridge_test.go
package logger

import (
	"testing"

	"github.com/MadeByDoug/wls-chatbot/pkg/zerologtest"
	"github.com/rs/zerolog"
)

// TestLogPreservesFatalSeverityWithoutTerminating verifies fatal logs are mapped to error level with explicit severity metadata.
func TestLogPreservesFatalSeverityWithoutTerminating(t *testing.T) {

	recorder := zerologtest.NewRecorder()
	bridge := NewLogBridge(zerolog.New(recorder))

	bridge.Log(LogEntry{
		Level:   "fatal",
		Message: "frontend fatal",
		Fields: map[string]string{
			"component": "frontend",
		},
	})

	entry, ok := recorder.Last()
	if !ok {
		t.Fatalf("expected one log entry")
	}

	if level := entry.Level(); level != "error" {
		t.Fatalf("expected mapped level error, got %q", level)
	}
	if value, ok := entry.StringField("frontend_level"); !ok || value != "fatal" {
		t.Fatalf("expected frontend_level=fatal, got %q", value)
	}
	if value, ok := entry.StringField("severity"); !ok || value != "fatal" {
		t.Fatalf("expected severity=fatal, got %q", value)
	}
	if value, ok := entry.StringField("frontend_severity"); !ok || value != "fatal" {
		t.Fatalf("expected frontend_severity=fatal, got %q", value)
	}
}
