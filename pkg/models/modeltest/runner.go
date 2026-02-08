// runner.go orchestrates test execution across providers and capabilities.
// pkg/models/modeltest/runner.go
package modeltest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// RunnerConfig configures the test runner.
type RunnerConfig struct {
	OutputDir    string       // Directory for golden files
	Mode         string       // "live" or "mock"
	Providers    []string     // Providers to test (empty = all)
	Capabilities []Capability // Capabilities to test (empty = all)
	Parallel     int          // Max parallel tests (0 = sequential)
	Timeout      time.Duration
}

// DefaultRunnerConfig returns sensible defaults.
func DefaultRunnerConfig() RunnerConfig {
	return RunnerConfig{
		OutputDir: "testdata/golden",
		Mode:      "mock",
		Parallel:  1,
		Timeout:   30 * time.Second,
	}
}

// ProviderConfig describes a provider for testing.
type ProviderConfig struct {
	Name    string       `yaml:"name"`
	Type    ProviderType `yaml:"type"`
	BaseURL string       `yaml:"base_url"`
	APIKey  string       `yaml:"-"` // From environment
	Models  []ModelConfig `yaml:"models"`
}

// ModelConfig describes a model for testing.
type ModelConfig struct {
	ID           string       `yaml:"id"`
	Capabilities []Capability `yaml:"capabilities"`
}

// TestPlan contains all tests to run.
type TestPlan struct {
	Providers []ProviderConfig
}

// Runner executes tests according to a plan.
type Runner struct {
	config RunnerConfig
	plan   *TestPlan
	index  *GoldenIndex
}

// NewRunner creates a test runner.
func NewRunner(config RunnerConfig) *Runner {
	return &Runner{
		config: config,
	}
}

// LoadPlan loads the test plan from a YAML file.
func (r *Runner) LoadPlan(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read plan file: %w", err)
	}

	var plan TestPlan
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return fmt.Errorf("parse plan: %w", err)
	}

	// Load API keys from environment
	for i := range plan.Providers {
		envVar := fmt.Sprintf("%s_API_KEY", toEnvName(plan.Providers[i].Name))
		plan.Providers[i].APIKey = os.Getenv(envVar)
	}

	r.plan = &plan
	return nil
}

// LoadEmbeddedCatalogPlan derives and loads a test plan from the canonical model catalog.
func (r *Runner) LoadEmbeddedCatalogPlan() error {

	plan, err := LoadEmbeddedCatalogPlan()
	if err != nil {
		return err
	}
	r.plan = plan
	return nil
}

// LoadGoldenFiles loads golden files for mock mode.
func (r *Runner) LoadGoldenFiles() error {
	files, err := LoadGoldenFiles(r.config.OutputDir)
	if err != nil {
		if os.IsNotExist(err) {
			r.index = NewGoldenIndex(nil)
			return nil
		}
		return err
	}
	r.index = NewGoldenIndex(files)
	return nil
}

// Run executes all tests in the plan.
func (r *Runner) Run(ctx context.Context) ([]TestResult, error) {
	if r.plan == nil {
		return nil, fmt.Errorf("no test plan loaded")
	}

	if r.config.Mode == "mock" {
		if err := r.LoadGoldenFiles(); err != nil {
			return nil, fmt.Errorf("load golden files: %w", err)
		}
	}

	var results []TestResult
	var mu sync.Mutex

	// Create work items
	type workItem struct {
		provider ProviderConfig
		model    ModelConfig
		cap      Capability
	}
	var work []workItem

	for _, prov := range r.plan.Providers {
		if !r.shouldTestProvider(prov.Name) {
			continue
		}
		for _, model := range prov.Models {
			for _, cap := range model.Capabilities {
				if !r.shouldTestCapability(cap) {
					continue
				}
				work = append(work, workItem{prov, model, cap})
			}
		}
	}

	// Execute tests
	workers := r.config.Parallel
	if workers <= 0 {
		workers = 1
	}

	workCh := make(chan workItem, len(work))
	for _, w := range work {
		workCh <- w
	}
	close(workCh)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for w := range workCh {
				result := r.runTest(ctx, w.provider, w.model, w.cap)
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return results, nil
}

// runTest executes a single test.
func (r *Runner) runTest(ctx context.Context, prov ProviderConfig, model ModelConfig, cap Capability) TestResult {
	ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	client := NewClient(ClientConfig{
		ProviderName: prov.Name,
		ProviderType: prov.Type,
		BaseURL:      prov.BaseURL,
		APIKey:       prov.APIKey,
	})

	var transport http.RoundTripper
	if r.config.Mode == "live" {
		recorder := NewRecordingTransport(nil)
		transport = recorder
		client.SetTransport(transport)

		tester := GetTester(cap)
		if tester == nil {
			return TestResult{
				Provider:   prov.Name,
				Capability: cap,
				Model:      model.ID,
				Error:      "unknown capability",
			}
		}

		result := tester.Test(ctx, client, model.ID)

		// Save recording as golden file
		if result.Success && len(recorder.Recordings()) > 0 {
			golden := NewGoldenFile(prov.Name, string(cap), model.ID, recorder.Recordings()[0])
			if err := golden.Save(r.config.OutputDir); err != nil {
				result.Error = fmt.Sprintf("save golden file: %v", err)
			}
		}

		return result
	}

	// Mock mode
	mockTransport := NewMockTransport(r.index, prov.Name, model.ID)
	client.SetTransport(mockTransport)

	tester := GetTester(cap)
	if tester == nil {
		return TestResult{
			Provider:   prov.Name,
			Capability: cap,
			Model:      model.ID,
			Error:      "unknown capability",
		}
	}

	return tester.Test(ctx, client, model.ID)
}

// shouldTestProvider checks if a provider should be tested.
func (r *Runner) shouldTestProvider(name string) bool {
	if len(r.config.Providers) == 0 {
		return true
	}
	for _, p := range r.config.Providers {
		if p == name {
			return true
		}
	}
	return false
}

// shouldTestCapability checks if a capability should be tested.
func (r *Runner) shouldTestCapability(cap Capability) bool {
	if len(r.config.Capabilities) == 0 {
		return true
	}
	for _, c := range r.config.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}

// toEnvName converts a provider name to environment variable format.
func toEnvName(name string) string {
	result := ""
	for _, c := range name {
		if c >= 'a' && c <= 'z' {
			result += string(c - 32) // uppercase
		} else if c >= 'A' && c <= 'Z' {
			result += string(c)
		} else if c >= '0' && c <= '9' {
			result += string(c)
		} else {
			result += "_"
		}
	}
	return result
}

// Report contains aggregated test results.
type Report struct {
	Timestamp  time.Time    `json:"timestamp"`
	TotalTests int          `json:"total_tests"`
	Passed     int          `json:"passed"`
	Failed     int          `json:"failed"`
	Results    []TestResult `json:"results"`
}

// GenerateReport creates a summary report from test results.
func GenerateReport(results []TestResult) Report {
	report := Report{
		Timestamp:  time.Now(),
		TotalTests: len(results),
		Results:    results,
	}
	for _, r := range results {
		if r.Success {
			report.Passed++
		} else {
			report.Failed++
		}
	}
	return report
}
