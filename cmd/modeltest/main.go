// Package main provides a standalone CLI for LLM model testing.
// This binary can be distributed independently for community model validation.
// cmd/modeltest/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MadeByDoug/wls-chatbot/pkg/models/modeltest"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modeltest",
		Short: "LLM model capability testing tool",
		Long: `A standalone tool for testing LLM provider capabilities and generating golden files.

Golden files capture request/response pairs for mock-based regression testing.
This tool can run in two modes:
  - live:  Make real API calls and update golden files
  - mock:  Replay golden files without API calls`,
	}

	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newValidateCommand())
	return cmd
}

func newRunCommand() *cobra.Command {
	var config modeltest.RunnerConfig
	var providersStr string
	var capsStr string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run capability tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			if providersStr != "" {
				config.Providers = strings.Split(providersStr, ",")
			}
			if capsStr != "" {
				for _, c := range strings.Split(capsStr, ",") {
					config.Capabilities = append(config.Capabilities, modeltest.Capability(c))
				}
			}

			runner := modeltest.NewRunner(config)
			if err := runner.LoadEmbeddedCatalogPlan(); err != nil {
				return fmt.Errorf("load embedded catalog plan: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			fmt.Printf("Running tests in %s mode...\n", config.Mode)
			results, err := runner.Run(ctx)
			if err != nil {
				return err
			}

			report := modeltest.GenerateReport(results)
			printReport(report)

			if report.Failed > 0 {
				return fmt.Errorf("%d tests failed", report.Failed)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&config.Mode, "mode", "mock", "Test mode: live or mock")
	cmd.Flags().StringVar(&config.OutputDir, "output", "testdata/golden", "Golden file output directory")
	cmd.Flags().StringVar(&providersStr, "providers", "", "Comma-separated list of providers to test")
	cmd.Flags().StringVar(&capsStr, "capabilities", "", "Comma-separated list of capabilities to test")
	cmd.Flags().IntVar(&config.Parallel, "parallel", 1, "Number of parallel tests")
	cmd.Flags().DurationVar(&config.Timeout, "timeout", 30*time.Second, "Timeout per test")

	return cmd
}

func newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate a test plan or golden file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Try loading as golden file first
			if strings.HasSuffix(path, ".json") {
				golden, err := modeltest.LoadGoldenFile(path)
				if err != nil {
					return fmt.Errorf("invalid golden file: %w", err)
				}
				fmt.Printf("✓ Valid golden file\n")
				fmt.Printf("  Provider:   %s\n", golden.Metadata.Provider)
				fmt.Printf("  Capability: %s\n", golden.Metadata.Capability)
				fmt.Printf("  Model:      %s\n", golden.Metadata.Model)
				fmt.Printf("  Timestamp:  %s\n", golden.Metadata.Timestamp.Format(time.RFC3339))
				return nil
			}

			if err := validatePlanFile(path); err != nil {
				return fmt.Errorf("invalid plan file: %w", err)
			}
			fmt.Printf("✓ Valid test plan\n")
			return nil
		},
	}
	return cmd
}

func printReport(report modeltest.Report) {
	fmt.Printf("\n=== Test Report ===\n")
	fmt.Printf("Timestamp: %s\n", report.Timestamp.Format(time.RFC3339))
	fmt.Printf("Total: %d  Passed: %d  Failed: %d\n\n", report.TotalTests, report.Passed, report.Failed)

	for _, r := range report.Results {
		status := "✓"
		if !r.Success {
			status = "✗"
		}
		fmt.Printf("%s %s/%s/%s\n", status, r.Provider, r.Capability, r.Model)
		if r.Error != "" {
			fmt.Printf("    Error: %s\n", r.Error)
		}
	}
}

// Export report as JSON if needed
func exportReportJSON(report modeltest.Report, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func validatePlanFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read plan file: %w", err)
	}

	var plan modeltest.TestPlan
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return fmt.Errorf("parse plan: %w", err)
	}

	return nil
}
