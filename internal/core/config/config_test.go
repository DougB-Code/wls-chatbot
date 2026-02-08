// config_test.go verifies application configuration helpers.
// internal/core/config/app/config_test.go
package config

import (
	"errors"
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

// TestDefaultConfigFromCatalogBuildsProviders verifies default config is derived from canonical model catalog.
func TestDefaultConfigFromCatalogBuildsProviders(t *testing.T) {

	cfg, err := defaultConfigFromCatalog()
	if err != nil {
		t.Fatalf("default config from catalog: %v", err)
	}
	if len(cfg.Providers) == 0 {
		t.Fatalf("expected providers from model catalog")
	}

	for _, provider := range cfg.Providers {
		if provider.Name == "" {
			t.Fatalf("expected provider name")
		}
		if provider.Type == "" {
			t.Fatalf("expected provider type for %s", provider.Name)
		}
		if len(provider.Models) == 0 {
			t.Fatalf("expected models for provider %s", provider.Name)
		}
		if provider.DefaultModel == "" {
			t.Fatalf("expected default model for provider %s", provider.Name)
		}
	}
}

// TestLoadConfigBootstrapsFromCatalog verifies missing config is seeded from canonical model catalog.
func TestLoadConfigBootstrapsFromCatalog(t *testing.T) {

	store := &stubStore{
		loadErr: ErrConfigNotFound,
	}

	cfg, err := LoadConfig(store)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(cfg.Providers) == 0 {
		t.Fatalf("expected providers from catalog seed")
	}
	if !store.saveCalled {
		t.Fatalf("expected save to be called")
	}
}

// stubStore implements Store for config tests.
type stubStore struct {
	cfg        AppConfig
	loadErr    error
	saveErr    error
	saveCalled bool
}

func (s *stubStore) Load() (AppConfig, error) {

	if s.loadErr != nil {
		return AppConfig{}, s.loadErr
	}
	return s.cfg, nil
}

func (s *stubStore) Save(cfg AppConfig) error {

	s.saveCalled = true
	s.cfg = cfg
	return s.saveErr
}

// TestLoadConfigPropagatesLoadErrors verifies non-not-found load errors are returned.
func TestLoadConfigPropagatesLoadErrors(t *testing.T) {

	store := &stubStore{
		loadErr: errors.New("boom"),
	}

	if _, err := LoadConfig(store); err == nil {
		t.Fatalf("expected load error")
	}
}
