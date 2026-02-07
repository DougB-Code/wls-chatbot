// assertions.go provides matcher utilities and test diagnostics for captured entries.
// pkg/zerologtest/assertions.go
package zerologtest

import (
	"fmt"
	"strings"
	"testing"
)

// AssertionOption configures assertion diagnostics behavior.
type AssertionOption func(config *assertionConfig)

// assertionConfig defines assertion rendering defaults.
type assertionConfig struct {
	lastNEntries int
	focusFields  []string
}

var defaultFocusFields = []string{
	FieldEventCode,
	FieldLevel,
	FieldComponent,
	FieldOp,
	FieldStatus,
	FieldMessage,
}

// WithLastNEntries limits diagnostic output to the last N entries.
func WithLastNEntries(lastNEntries int) AssertionOption {

	return func(config *assertionConfig) {
		if lastNEntries > 0 {
			config.lastNEntries = lastNEntries
		}
	}
}

// WithFocusFields appends additional fields to diagnostic output.
func WithFocusFields(fields ...string) AssertionOption {

	return func(config *assertionConfig) {
		for _, field := range fields {
			if strings.TrimSpace(field) != "" {
				config.focusFields = append(config.focusFields, field)
			}
		}
	}
}

// Contains returns whether any entry matches the predicate.
func Contains(entries []Entry, predicate Predicate) bool {

	if predicate == nil {
		return false
	}

	for _, entry := range entries {
		if predicate.Match(entry) {
			return true
		}
	}
	return false
}

// InOrder returns whether predicates match in the same order as entries.
func InOrder(entries []Entry, predicates ...Predicate) bool {

	normalizedPredicates := normalizePredicates(predicates)
	if len(normalizedPredicates) == 0 {
		return true
	}

	predicateIndex := 0
	for _, entry := range entries {
		if normalizedPredicates[predicateIndex].Match(entry) {
			predicateIndex++
			if predicateIndex == len(normalizedPredicates) {
				return true
			}
		}
	}

	return false
}

// AssertContains fails the test when no entry matches.
func AssertContains(t testing.TB, entries []Entry, predicate Predicate, options ...AssertionOption) {

	t.Helper()

	if Contains(entries, predicate) {
		return
	}

	config := buildAssertionConfig(options...)
	t.Fatalf(
		"expected at least one entry matching %s, but none matched.\nRecent entries:\n%s",
		describePredicate(predicate),
		renderEntries(lastNEntries(entries, config.lastNEntries), config.focusFields),
	)
}

// AssertNotContains fails the test when any entry matches.
func AssertNotContains(t testing.TB, entries []Entry, predicate Predicate, options ...AssertionOption) {

	t.Helper()

	if !Contains(entries, predicate) {
		return
	}

	config := buildAssertionConfig(options...)
	matchingEntries := filterEntries(entries, predicate)
	t.Fatalf(
		"expected no entries matching %s, but found %d matching entries.\nMatching entries:\n%s",
		describePredicate(predicate),
		len(matchingEntries),
		renderEntries(lastNEntries(matchingEntries, config.lastNEntries), config.focusFields),
	)
}

// AssertInOrder fails the test when predicates do not match in order.
func AssertInOrder(t testing.TB, entries []Entry, predicates []Predicate, options ...AssertionOption) {

	t.Helper()

	if InOrder(entries, predicates...) {
		return
	}

	config := buildAssertionConfig(options...)
	t.Fatalf(
		"expected entries to match predicates in order (%s), but order check failed.\nRecent entries:\n%s",
		joinPredicateDescriptions(normalizePredicates(predicates)),
		renderEntries(lastNEntries(entries, config.lastNEntries), config.focusFields),
	)
}

// AssertRecorderContains fails when recorder entries do not contain a match.
func AssertRecorderContains(t testing.TB, recorder *Recorder, predicate Predicate, options ...AssertionOption) {

	t.Helper()

	entries := recorder.Entries()
	AssertContains(t, entries, predicate, options...)
}

// AssertRecorderNotContains fails when recorder entries contain a match.
func AssertRecorderNotContains(t testing.TB, recorder *Recorder, predicate Predicate, options ...AssertionOption) {

	t.Helper()

	entries := recorder.Entries()
	AssertNotContains(t, entries, predicate, options...)
}

// AssertRecorderInOrder fails when recorder entries do not match predicate order.
func AssertRecorderInOrder(t testing.TB, recorder *Recorder, predicates []Predicate, options ...AssertionOption) {

	t.Helper()

	entries := recorder.Entries()
	AssertInOrder(t, entries, predicates, options...)
}

// buildAssertionConfig constructs diagnostics defaults.
func buildAssertionConfig(options ...AssertionOption) assertionConfig {

	config := assertionConfig{
		lastNEntries: 10,
		focusFields:  append([]string{}, defaultFocusFields...),
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	config.focusFields = dedupe(config.focusFields)
	return config
}

// renderEntries renders entries with selected fields for failure diagnostics.
func renderEntries(entries []Entry, focusFields []string) string {

	if len(entries) == 0 {
		return "<no entries captured>"
	}

	var builder strings.Builder
	for index, entry := range entries {
		if index > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(fmt.Sprintf("[%d]", index))
		for _, field := range focusFields {
			value, ok := entry.Field(field)
			if !ok {
				continue
			}
			builder.WriteString(fmt.Sprintf(" %s=%q", field, fmt.Sprint(value)))
		}
	}
	return builder.String()
}

// filterEntries returns entries that match a predicate.
func filterEntries(entries []Entry, predicate Predicate) []Entry {

	if predicate == nil {
		return nil
	}

	filtered := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if predicate.Match(entry) {
			filtered = append(filtered, copyEntry(entry))
		}
	}
	return filtered
}

// lastNEntries returns the last N entries from a slice.
func lastNEntries(entries []Entry, lastNEntries int) []Entry {

	if lastNEntries <= 0 || len(entries) <= lastNEntries {
		copied := make([]Entry, len(entries))
		for index, entry := range entries {
			copied[index] = copyEntry(entry)
		}
		return copied
	}

	start := len(entries) - lastNEntries
	copied := make([]Entry, len(entries[start:]))
	for index, entry := range entries[start:] {
		copied[index] = copyEntry(entry)
	}
	return copied
}

// dedupe removes duplicated field names while preserving order.
func dedupe(items []string) []string {

	seen := map[string]struct{}{}
	deduplicated := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		deduplicated = append(deduplicated, item)
	}
	return deduplicated
}
