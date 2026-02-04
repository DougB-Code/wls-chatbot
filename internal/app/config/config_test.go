// config_test.go verifies configuration helpers.
// internal/app/config/config_test.go
package config

import (
	"testing"
	"time"
)

// TestResolveUpdateFrequenciesReturnsNilWhenEmpty verifies no frequencies are returned for empty values.
func TestResolveUpdateFrequenciesReturnsNilWhenEmpty(t *testing.T) {

	cfg := AppConfig{
		Providers: []ProviderConfig{
			{Name: "openai"},
		},
	}

	frequencies, err := ResolveUpdateFrequencies(cfg)
	if err != nil {
		t.Fatalf("resolve frequencies: %v", err)
	}
	if frequencies != nil {
		t.Fatalf("expected nil frequencies, got %+v", frequencies)
	}
}

// TestResolveUpdateFrequenciesSkipsManual verifies manual update frequency is ignored.
func TestResolveUpdateFrequenciesSkipsManual(t *testing.T) {

	cfg := AppConfig{
		Providers: []ProviderConfig{
			{Name: "openai", UpdateFrequency: UpdateFrequencyManual},
		},
	}

	frequencies, err := ResolveUpdateFrequencies(cfg)
	if err != nil {
		t.Fatalf("resolve frequencies: %v", err)
	}
	if frequencies != nil {
		t.Fatalf("expected nil frequencies, got %+v", frequencies)
	}
}

// TestResolveUpdateFrequenciesParsesValues verifies known frequencies are parsed.
func TestResolveUpdateFrequenciesParsesValues(t *testing.T) {

	cfg := AppConfig{
		Providers: []ProviderConfig{
			{Name: "openai", UpdateFrequency: UpdateFrequencyHourly},
			{Name: "gemini", UpdateFrequency: UpdateFrequencyWeekly},
		},
	}

	frequencies, err := ResolveUpdateFrequencies(cfg)
	if err != nil {
		t.Fatalf("resolve frequencies: %v", err)
	}
	if len(frequencies) != 2 {
		t.Fatalf("expected 2 frequencies, got %d", len(frequencies))
	}
	if frequencies["openai"] != time.Hour {
		t.Fatalf("expected hourly frequency, got %v", frequencies["openai"])
	}
	if frequencies["gemini"] != 7*24*time.Hour {
		t.Fatalf("expected weekly frequency, got %v", frequencies["gemini"])
	}
}

// TestResolveUpdateFrequenciesRejectsUnknown verifies invalid frequencies error.
func TestResolveUpdateFrequenciesRejectsUnknown(t *testing.T) {

	cfg := AppConfig{
		Providers: []ProviderConfig{
			{Name: "openai", UpdateFrequency: "fortnightly"},
		},
	}

	if _, err := ResolveUpdateFrequencies(cfg); err == nil {
		t.Fatalf("expected error for unknown update frequency")
	}
}

// TestParseUpdateFrequencyHandlesKnownValues verifies the known update frequencies.
func TestParseUpdateFrequencyHandlesKnownValues(t *testing.T) {

	cases := []struct {
		name     string
		input    UpdateFrequency
		expected time.Duration
	}{
		{name: "manual", input: UpdateFrequencyManual, expected: 0},
		{name: "hourly", input: UpdateFrequencyHourly, expected: time.Hour},
		{name: "daily", input: UpdateFrequencyDaily, expected: 24 * time.Hour},
		{name: "weekly", input: UpdateFrequencyWeekly, expected: 7 * 24 * time.Hour},
	}

	for _, testCase := range cases {
		result, err := parseUpdateFrequency(testCase.input)
		if err != nil {
			t.Fatalf("parse frequency %s: %v", testCase.name, err)
		}
		if result != testCase.expected {
			t.Fatalf("expected %v for %s, got %v", testCase.expected, testCase.name, result)
		}
	}
}

// TestParseUpdateFrequencyRejectsUnknown verifies unknown values return errors.
func TestParseUpdateFrequencyRejectsUnknown(t *testing.T) {

	if _, err := parseUpdateFrequency("unknown"); err == nil {
		t.Fatalf("expected error for unknown update frequency")
	}
}
