// recorder.go captures zerolog JSON output into concurrency-safe test entries.
// pkg/zerologtest/recorder.go
package zerologtest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Transform scrubs or normalizes decoded entry fields before storage.
type Transform func(fields map[string]any) map[string]any

// Option configures Recorder behavior.
type Option func(recorder *Recorder)

// Recorder captures newline-delimited JSON log entries from zerolog.
type Recorder struct {
	mu                 sync.RWMutex
	pending            []byte
	entries            []Entry
	ignoreFields       map[string]struct{}
	presenceOnlyFields map[string]struct{}
	transform          Transform
	notifyCh           chan struct{}
}

// NewRecorder creates a new structured log recorder.
func NewRecorder(options ...Option) *Recorder {

	recorder := &Recorder{
		ignoreFields:       map[string]struct{}{},
		presenceOnlyFields: map[string]struct{}{},
		notifyCh:           make(chan struct{}, 1),
	}

	for _, option := range options {
		if option != nil {
			option(recorder)
		}
	}

	return recorder
}

// WithIgnoredFields removes configured fields before entries are stored.
func WithIgnoredFields(fields ...string) Option {

	return func(recorder *Recorder) {
		for _, field := range fields {
			if field != "" {
				recorder.ignoreFields[field] = struct{}{}
			}
		}
	}
}

// WithPresenceOnlyFields replaces configured field values with PresenceOnlyValue.
func WithPresenceOnlyFields(fields ...string) Option {

	return func(recorder *Recorder) {
		for _, field := range fields {
			if field != "" {
				recorder.presenceOnlyFields[field] = struct{}{}
			}
		}
	}
}

// WithTransform applies a normalization transform before entry storage.
func WithTransform(transform Transform) Option {

	return func(recorder *Recorder) {
		recorder.transform = transform
	}
}

// Write buffers log bytes and records complete newline-delimited JSON entries.
func (recorder *Recorder) Write(payload []byte) (int, error) {

	if recorder == nil {
		return 0, errors.New("recorder is nil")
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	recorder.pending = append(recorder.pending, payload...)

	for {
		newlineIndex := bytes.IndexByte(recorder.pending, '\n')
		if newlineIndex < 0 {
			break
		}

		line := make([]byte, newlineIndex)
		copy(line, recorder.pending[:newlineIndex])
		recorder.pending = recorder.pending[newlineIndex+1:]
		recorder.captureLine(line)
	}

	return len(payload), nil
}

// Entries returns a snapshot copy of captured entries.
func (recorder *Recorder) Entries() []Entry {

	if recorder == nil {
		return nil
	}

	recorder.mu.RLock()
	defer recorder.mu.RUnlock()

	copied := make([]Entry, len(recorder.entries))
	for index, entry := range recorder.entries {
		copied[index] = copyEntry(entry)
	}
	return copied
}

// Last returns the most recent captured entry.
func (recorder *Recorder) Last() (Entry, bool) {

	if recorder == nil {
		return Entry{}, false
	}

	recorder.mu.RLock()
	defer recorder.mu.RUnlock()

	if len(recorder.entries) == 0 {
		return Entry{}, false
	}
	return copyEntry(recorder.entries[len(recorder.entries)-1]), true
}

// Filter returns entries that match the predicate.
func (recorder *Recorder) Filter(predicate Predicate) []Entry {

	if recorder == nil || predicate == nil {
		return nil
	}

	recorder.mu.RLock()
	defer recorder.mu.RUnlock()

	filtered := make([]Entry, 0, len(recorder.entries))
	for _, entry := range recorder.entries {
		if predicate.Match(entry) {
			filtered = append(filtered, copyEntry(entry))
		}
	}
	return filtered
}

// Contains returns whether at least one entry matches.
func (recorder *Recorder) Contains(predicate Predicate) bool {

	if recorder == nil || predicate == nil {
		return false
	}

	recorder.mu.RLock()
	defer recorder.mu.RUnlock()

	for _, entry := range recorder.entries {
		if predicate.Match(entry) {
			return true
		}
	}
	return false
}

// WaitFor blocks until an entry matches or context cancellation occurs.
func (recorder *Recorder) WaitFor(ctx context.Context, predicate Predicate) (Entry, error) {

	if recorder == nil {
		return Entry{}, errors.New("recorder is nil")
	}
	if predicate == nil {
		return Entry{}, errors.New("predicate is nil")
	}

	for {
		if entry, ok := recorder.firstMatch(predicate); ok {
			return entry, nil
		}

		select {
		case <-ctx.Done():
			return Entry{}, ctx.Err()
		case <-recorder.notifyCh:
		}
	}
}

// firstMatch returns the first matching entry snapshot.
func (recorder *Recorder) firstMatch(predicate Predicate) (Entry, bool) {

	recorder.mu.RLock()
	defer recorder.mu.RUnlock()

	for _, entry := range recorder.entries {
		if predicate.Match(entry) {
			return copyEntry(entry), true
		}
	}
	return Entry{}, false
}

// captureLine decodes and stores one complete log line.
func (recorder *Recorder) captureLine(line []byte) {

	trimmed := strings.TrimSpace(strings.TrimSuffix(string(line), "\r"))
	if trimmed == "" {
		return
	}

	entryFields := map[string]any{}
	if err := json.Unmarshal([]byte(trimmed), &entryFields); err != nil {
		entryFields = map[string]any{
			FieldLevel:     "error",
			FieldEventCode: EventCodeDecodeFailed,
			FieldComponent: "zerologtest.recorder",
			FieldOp:        "decode",
			FieldStatus:    "failed",
			FieldErrKind:   "json_decode",
			FieldMessage:   "failed to decode log line",
			"decode_error": err.Error(),
			"raw_line":     trimmed,
		}
	}

	normalized := recorder.normalizeFields(entryFields)
	recorder.entries = append(recorder.entries, Entry{Fields: normalized})
	recorder.signalWaiters()
}

// normalizeFields applies transform, ignore, and presence-only behavior.
func (recorder *Recorder) normalizeFields(fields map[string]any) map[string]any {

	normalized := copyFields(fields)
	if recorder.transform != nil {
		transformed := recorder.transform(copyFields(normalized))
		if transformed != nil {
			normalized = transformed
		}
	}

	if normalized == nil {
		normalized = map[string]any{}
	}

	for ignoredField := range recorder.ignoreFields {
		delete(normalized, ignoredField)
	}
	for presenceOnlyField := range recorder.presenceOnlyFields {
		if _, ok := normalized[presenceOnlyField]; ok {
			normalized[presenceOnlyField] = PresenceOnlyValue
		}
	}

	return normalized
}

// signalWaiters notifies async waiters that new entries were captured.
func (recorder *Recorder) signalWaiters() {

	select {
	case recorder.notifyCh <- struct{}{}:
	default:
	}
}

// String returns a concise snapshot count for quick diagnostics.
func (recorder *Recorder) String() string {

	if recorder == nil {
		return "Recorder(<nil>)"
	}

	recorder.mu.RLock()
	defer recorder.mu.RUnlock()

	return fmt.Sprintf("Recorder(entries=%d)", len(recorder.entries))
}
