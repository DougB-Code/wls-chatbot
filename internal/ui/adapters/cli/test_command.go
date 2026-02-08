// test_command.go defines CLI adapters for model test workflows.
// internal/ui/adapters/cli/test_command.go
package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MadeByDoug/wls-chatbot/pkg/models/modeltest"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// newTestCommand creates the parent 'test' command with subcommands.
func newTestCommand(log zerolog.Logger) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests",
	}
	cmd.AddCommand(newTestGoldenCommand(log))
	return cmd
}

// newTestGoldenCommand creates the 'test golden' command for model capability testing.
func newTestGoldenCommand(log zerolog.Logger) *cobra.Command {

	var mode string
	var outputDir string
	var providersStr string
	var capsStr string
	var parallel int
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "golden",
		Short: "Run model capability tests with golden file capture",
		Long: `Run capability tests against LLM providers and generate/validate golden files.

In 'live' mode, real API calls are made and golden files are updated.
In 'mock' mode, golden files are replayed without API calls.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config := modeltest.RunnerConfig{
				OutputDir: outputDir,
				Mode:      mode,
				Parallel:  parallel,
				Timeout:   timeout,
			}

			if providersStr != "" {
				config.Providers = strings.Split(providersStr, ",")
			}
			if capsStr != "" {
				for _, capability := range strings.Split(capsStr, ",") {
					config.Capabilities = append(config.Capabilities, modeltest.Capability(capability))
				}
			}

			runner := modeltest.NewRunner(config)
			if err := runner.LoadEmbeddedCatalogPlan(); err != nil {
				return fmt.Errorf("load embedded catalog plan: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			log.Info().Str("mode", mode).Msg("Running model capability tests...")
			results, err := runner.Run(ctx)
			if err != nil {
				return err
			}

			report := modeltest.GenerateReport(results)
			fmt.Printf("\n=== Test Report ===\n")
			fmt.Printf("Total: %d  Passed: %d  Failed: %d\n\n", report.TotalTests, report.Passed, report.Failed)

			for _, result := range report.Results {
				status := "✓"
				if !result.Success {
					status = "✗"
				}
				fmt.Printf("%s %s/%s/%s\n", status, result.Provider, result.Capability, result.Model)
				if result.Error != "" {
					fmt.Printf("    Error: %s\n", result.Error)
				}
			}

			if report.Failed > 0 {
				return fmt.Errorf("%d tests failed", report.Failed)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "mock", "Test mode: live or mock")
	cmd.Flags().StringVar(&outputDir, "output", "testdata/golden", "Golden file output directory")
	cmd.Flags().StringVar(&providersStr, "providers", "", "Comma-separated list of providers to test")
	cmd.Flags().StringVar(&capsStr, "capabilities", "", "Comma-separated list of capabilities to test")
	cmd.Flags().IntVar(&parallel, "parallel", 1, "Number of parallel tests")
	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "Timeout per test")

	return cmd
}
