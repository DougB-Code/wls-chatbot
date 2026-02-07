// predicate.go defines reusable log predicates and combinators.
// pkg/zerologtest/predicate.go
package zerologtest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Predicate matches a captured log entry.
type Predicate interface {
	// Match returns whether an entry satisfies this predicate.
	Match(entry Entry) bool
	// String describes the predicate for diagnostics.
	String() string
}

// predicateFunc adapts a function into a named Predicate.
type predicateFunc struct {
	description string
	match       func(entry Entry) bool
}

// Match runs the underlying predicate function.
func (predicate predicateFunc) Match(entry Entry) bool {

	if predicate.match == nil {
		return false
	}
	return predicate.match(entry)
}

// String returns a stable predicate description.
func (predicate predicateFunc) String() string {

	return predicate.description
}

// Match defines a custom predicate with a stable description.
func Match(description string, match func(entry Entry) bool) Predicate {

	desc := strings.TrimSpace(description)
	if desc == "" {
		desc = "custom predicate"
	}
	return predicateFunc{
		description: desc,
		match:       match,
	}
}

// HasEventCode matches entries with the provided event code.
func HasEventCode(eventCode string) Predicate {

	return Match(fmt.Sprintf("event_code == %q", eventCode), func(entry Entry) bool {
		return FieldEq(FieldEventCode, eventCode).Match(entry)
	})
}

// HasLevel matches entries with the provided level.
func HasLevel(level string) Predicate {

	return Match(fmt.Sprintf("level == %q", level), func(entry Entry) bool {
		return FieldEq(FieldLevel, level).Match(entry)
	})
}

// HasField matches entries that contain the field key.
func HasField(key string) Predicate {

	return Match(fmt.Sprintf("has field %q", key), func(entry Entry) bool {
		_, ok := entry.Field(key)
		return ok
	})
}

// FieldEq matches entries where the field value equals expected.
func FieldEq(key string, expected any) Predicate {

	return Match(fmt.Sprintf("%s == %v", key, expected), func(entry Entry) bool {
		value, ok := entry.Field(key)
		if !ok {
			return false
		}
		return valuesEqual(value, expected)
	})
}

// FieldContains matches entries where a string field includes a substring.
func FieldContains(key string, substring string) Predicate {

	return Match(fmt.Sprintf("%s contains %q", key, substring), func(entry Entry) bool {
		value, ok := entry.Field(key)
		if !ok {
			return false
		}

		switch typed := value.(type) {
		case string:
			return strings.Contains(typed, substring)
		case []string:
			for _, item := range typed {
				if strings.Contains(item, substring) {
					return true
				}
			}
		case []any:
			for _, item := range typed {
				text, ok := item.(string)
				if ok && strings.Contains(text, substring) {
					return true
				}
			}
		}

		return false
	})
}

// And matches when all predicates are true.
func And(predicates ...Predicate) Predicate {

	normalized := normalizePredicates(predicates)
	return Match("and("+joinPredicateDescriptions(normalized)+")", func(entry Entry) bool {
		for _, predicate := range normalized {
			if !predicate.Match(entry) {
				return false
			}
		}
		return true
	})
}

// Or matches when at least one predicate is true.
func Or(predicates ...Predicate) Predicate {

	normalized := normalizePredicates(predicates)
	return Match("or("+joinPredicateDescriptions(normalized)+")", func(entry Entry) bool {
		for _, predicate := range normalized {
			if predicate.Match(entry) {
				return true
			}
		}
		return false
	})
}

// Not negates a predicate.
func Not(predicate Predicate) Predicate {

	description := "not(<nil>)"
	if predicate != nil {
		description = "not(" + predicate.String() + ")"
	}
	return Match(description, func(entry Entry) bool {
		if predicate == nil {
			return true
		}
		return !predicate.Match(entry)
	})
}

// describePredicate returns a stable fallback-safe predicate description.
func describePredicate(predicate Predicate) string {

	if predicate == nil {
		return "<nil predicate>"
	}
	description := strings.TrimSpace(predicate.String())
	if description == "" {
		return "<unnamed predicate>"
	}
	return description
}

// normalizePredicates filters nil predicates for combinator safety.
func normalizePredicates(predicates []Predicate) []Predicate {

	normalized := make([]Predicate, 0, len(predicates))
	for _, predicate := range predicates {
		if predicate != nil {
			normalized = append(normalized, predicate)
		}
	}
	return normalized
}

// joinPredicateDescriptions concatenates predicate descriptions.
func joinPredicateDescriptions(predicates []Predicate) string {

	if len(predicates) == 0 {
		return ""
	}

	descriptions := make([]string, 0, len(predicates))
	for _, predicate := range predicates {
		descriptions = append(descriptions, describePredicate(predicate))
	}
	return strings.Join(descriptions, ", ")
}

// valuesEqual compares JSON-decoded values and expected values with numeric tolerance.
func valuesEqual(actual any, expected any) bool {

	if reflect.DeepEqual(actual, expected) {
		return true
	}

	actualNumber, actualIsNumber := asFloat64(actual)
	expectedNumber, expectedIsNumber := asFloat64(expected)
	if actualIsNumber && expectedIsNumber {
		return actualNumber == expectedNumber
	}

	return false
}

// asFloat64 converts common numeric JSON value types for robust equality checks.
func asFloat64(value any) (float64, bool) {

	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	case string:
		parsed, err := strconv.ParseFloat(typed, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
