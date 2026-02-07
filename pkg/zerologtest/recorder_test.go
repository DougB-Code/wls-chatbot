// recorder_test.go verifies recorder capture, normalization, and async waiting.
// pkg/zerologtest/recorder_test.go
package zerologtest

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestRecorderWriteCapturesPartialJSONLines verifies newline-delimited buffering behavior.
func TestRecorderWriteCapturesPartialJSONLines(t *testing.T) {

	recorder := NewRecorder()

	firstLine := `{"level":"info","event_code":"CHAT_STARTED","component":"chat","op":"stream","status":"ok","message":"started"}`
	secondLine := `{"level":"error","event_code":"CHAT_STREAM_FAILED","component":"chat","op":"stream","status":"failed","err_kind":"upstream","message":"stream failed"}`
	firstPayload := firstLine + "\n"

	cutIndex := len(firstPayload) / 2
	if _, err := recorder.Write([]byte(firstPayload[:cutIndex])); err != nil {
		t.Fatalf("write first fragment: %v", err)
	}
	if len(recorder.Entries()) != 0 {
		t.Fatalf("expected no complete entries after first fragment")
	}
	if _, err := recorder.Write([]byte(firstPayload[cutIndex:] + secondLine + "\n")); err != nil {
		t.Fatalf("write second fragment: %v", err)
	}

	entries := recorder.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].EventCode() != "CHAT_STARTED" {
		t.Fatalf("expected first event_code CHAT_STARTED, got %q", entries[0].EventCode())
	}
	if entries[1].EventCode() != "CHAT_STREAM_FAILED" {
		t.Fatalf("expected second event_code CHAT_STREAM_FAILED, got %q", entries[1].EventCode())
	}

	last, ok := recorder.Last()
	if !ok {
		t.Fatalf("expected last entry to exist")
	}
	if last.EventCode() != "CHAT_STREAM_FAILED" {
		t.Fatalf("expected last event CHAT_STREAM_FAILED, got %q", last.EventCode())
	}
}

// TestRecorderWriteSynthesizesDecodeFailure verifies malformed JSON handling.
func TestRecorderWriteSynthesizesDecodeFailure(t *testing.T) {

	recorder := NewRecorder()
	if _, err := recorder.Write([]byte(`{"level":"info"` + "\n")); err != nil {
		t.Fatalf("write malformed line: %v", err)
	}

	entries := recorder.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	entry := entries[0]
	if entry.EventCode() != EventCodeDecodeFailed {
		t.Fatalf("expected %s event, got %q", EventCodeDecodeFailed, entry.EventCode())
	}
	if entry.Level() != "error" {
		t.Fatalf("expected level error, got %q", entry.Level())
	}
	if _, ok := entry.Field("decode_error"); !ok {
		t.Fatalf("expected decode_error field in synthetic entry")
	}
}

// TestRecorderNormalizationOptions verifies ignore, presence-only, and transform hooks.
func TestRecorderNormalizationOptions(t *testing.T) {

	recorder := NewRecorder(
		WithIgnoredFields(FieldTime),
		WithPresenceOnlyFields(FieldCaller),
		WithTransform(func(fields map[string]any) map[string]any {
			fields["api_key"] = "[redacted]"
			return fields
		}),
	)

	line := `{"time":"2026-01-01T00:00:00Z","caller":"app/main.go:10","level":"info","event_code":"READY","api_key":"super-secret"}`
	if _, err := recorder.Write([]byte(line + "\n")); err != nil {
		t.Fatalf("write normalized line: %v", err)
	}

	entry := recorder.Entries()[0]
	if _, ok := entry.Field(FieldTime); ok {
		t.Fatalf("expected %q to be ignored", FieldTime)
	}
	callerValue, ok := entry.StringField(FieldCaller)
	if !ok {
		t.Fatalf("expected caller field to exist")
	}
	if callerValue != PresenceOnlyValue {
		t.Fatalf("expected caller to be %q, got %q", PresenceOnlyValue, callerValue)
	}
	apiKeyValue, ok := entry.StringField("api_key")
	if !ok {
		t.Fatalf("expected api_key to exist")
	}
	if apiKeyValue != "[redacted]" {
		t.Fatalf("expected api_key to be redacted, got %q", apiKeyValue)
	}
}

// TestRecorderWaitForReturnsMatchingEntry verifies asynchronous waiting behavior.
func TestRecorderWaitForReturnsMatchingEntry(t *testing.T) {

	recorder := NewRecorder()
	logger := NewLogger(recorder)

	go func() {
		time.Sleep(25 * time.Millisecond)
		logger.Info().
			Str(FieldEventCode, "SERVER_READY").
			Str(FieldComponent, "http").
			Str(FieldOp, "startup").
			Str(FieldStatus, "ok").
			Msg("ready")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	entry, err := recorder.WaitFor(ctx, HasEventCode("SERVER_READY"))
	if err != nil {
		t.Fatalf("wait for ready event: %v", err)
	}
	if entry.EventCode() != "SERVER_READY" {
		t.Fatalf("expected SERVER_READY, got %q", entry.EventCode())
	}
	if _, ok := entry.Field(FieldTime); !ok {
		t.Fatalf("expected timestamp field from zerolog With().Timestamp()")
	}
	if _, ok := entry.Field(FieldCaller); !ok {
		t.Fatalf("expected caller field from zerolog With().Caller()")
	}
}

// TestRecorderWaitForHonorsContextCancel verifies wait timeout behavior.
func TestRecorderWaitForHonorsContextCancel(t *testing.T) {

	recorder := NewRecorder()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_, err := recorder.WaitFor(ctx, HasEventCode("MISSING"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}
