// model_command.go defines AI CLI adapters for model catalog workflows.
// internal/ui/adapters/cli/ai/model_command.go
package ai

import (
	"context"
	"fmt"
	"strings"

	modelinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/ports"
	"github.com/spf13/cobra"
)

// newModelCommand creates the 'model' command with subcommands.
func newModelCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "model",
		Aliases: []string{"models"},
		Short:   "Manage model catalog",
	}
	cmd.AddCommand(newModelListCommand(deps))
	cmd.AddCommand(newModelImportCommand(deps))
	cmd.AddCommand(newModelSyncCommand(deps))
	return cmd
}

// newModelListCommand lists models in the catalog.
func newModelListCommand(deps Dependencies) *cobra.Command {

	var source string
	var requiredInputModalities []string
	var requiredOutputModalities []string
	var requiredCapabilityIDs []string
	var requiredSystemTags []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List models in the catalog",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			summaries, err := applicationFacade.Models.ListModels(context.Background(), modelinterfaces.ModelListFilter{
				Source:                   source,
				RequiredInputModalities:  requiredInputModalities,
				RequiredOutputModalities: requiredOutputModalities,
				RequiredCapabilityIDs:    requiredCapabilityIDs,
				RequiredSystemTags:       requiredSystemTags,
			})
			if err != nil {
				return err
			}

			fmt.Printf("%-40s %-15s %-12s %-10s\n", "MODEL ID", "PROVIDER", "SOURCE", "APPROVED")
			fmt.Println(strings.Repeat("-", 80))
			for _, summary := range summaries {
				approved := "no"
				if summary.Approved {
					approved = "yes"
				}
				fmt.Printf("%-40s %-15s %-12s %-10s\n", summary.ModelID, summary.ProviderName, summary.Source, approved)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&source, "source", "", "Filter by source (seed, user, discovered)")
	cmd.Flags().StringSliceVar(&requiredInputModalities, "requires-input-modality", nil, "Require one or more input modalities (repeat flag)")
	cmd.Flags().StringSliceVar(&requiredOutputModalities, "requires-output-modality", nil, "Require one or more output modalities (repeat flag)")
	cmd.Flags().StringSliceVar(&requiredCapabilityIDs, "requires-capability", nil, "Require one or more semantic capability IDs (repeat flag)")
	cmd.Flags().StringSliceVar(&requiredSystemTags, "requires-system-tag", nil, "Require one or more model system tags (repeat flag)")
	return cmd
}

// newModelImportCommand imports custom models from a YAML file.
func newModelImportCommand(deps Dependencies) *cobra.Command {

	var filePath string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import custom models from a YAML file",
		Long:  "Import custom models from a YAML file. Format matches models.yaml structure.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			if err := applicationFacade.Models.ImportModels(context.Background(), modelinterfaces.ImportModelsRequest{
				FilePath: filePath,
			}); err != nil {
				return err
			}

			fmt.Println("Custom models imported successfully.")
			return nil
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "Path to the custom models YAML file")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

// newModelSyncCommand re-syncs custom models from the default location.
func newModelSyncCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync custom models from app config directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			result, err := applicationFacade.Models.SyncModels(context.Background())
			if err != nil {
				return err
			}
			if !result.Imported {
				fmt.Printf("No custom models file found at %s\n", result.Path)
				return nil
			}

			fmt.Printf("Custom models synced from %s\n", result.Path)
			return nil
		},
	}
	return cmd
}
