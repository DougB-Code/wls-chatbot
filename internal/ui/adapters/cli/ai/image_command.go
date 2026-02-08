// image_command.go defines AI CLI adapters for image generation workflows.
// internal/ui/adapters/cli/ai/image_command.go
package ai

import (
	"context"
	"fmt"
	"os"

	appcontracts "github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
	"github.com/spf13/cobra"
)

// newImageCommand creates the parent 'image' command.
func newImageCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "image",
		Short: "Generate or edit images",
	}
	cmd.AddCommand(newImageGenerateCommand(deps))
	cmd.AddCommand(newImageEditCommand(deps))
	return cmd
}

// newImageGenerateCommand creates the 'image generate' command.
func newImageGenerateCommand(deps Dependencies) *cobra.Command {

	var providerName string
	var modelName string
	var prompt string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate an image",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			deps.BaseLogger.Info().Str("provider", providerName).Str("model", modelName).Msg("Generating image...")
			result, err := applicationFacade.Images.GenerateImage(context.Background(), appcontracts.GenerateImageRequest{
				ProviderName: providerName,
				ModelName:    modelName,
				Prompt:       prompt,
				N:            1,
			})
			if err != nil {
				return fmt.Errorf("generation failed: %w", err)
			}

			if outputPath != "" {
				if err := os.WriteFile(outputPath, result.Bytes, 0o644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				deps.BaseLogger.Info().Str("path", outputPath).Msg("Image saved")
			} else {
				deps.BaseLogger.Info().Int("bytes", len(result.Bytes)).Msg("Image generated (use --output to save)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name (e.g. gemini, openai)")
	_ = cmd.MarkFlagRequired("provider")
	cmd.Flags().StringVar(&modelName, "model", "", "Model name (optional)")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Image prompt")
	_ = cmd.MarkFlagRequired("prompt")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output path for the generated image")

	return cmd
}

// newImageEditCommand creates the 'image edit' command.
func newImageEditCommand(deps Dependencies) *cobra.Command {

	var providerName string
	var modelName string
	var prompt string
	var outputPath string
	var imagePath string
	var maskPath string

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit an image",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			deps.BaseLogger.Info().Str("provider", providerName).Str("model", modelName).Msg("Editing image...")
			result, err := applicationFacade.Images.EditImage(context.Background(), appcontracts.EditImageRequest{
				ProviderName: providerName,
				ModelName:    modelName,
				Prompt:       prompt,
				ImagePath:    imagePath,
				MaskPath:     maskPath,
				N:            1,
			})
			if err != nil {
				return fmt.Errorf("editing failed: %w", err)
			}

			if outputPath != "" {
				if err := os.WriteFile(outputPath, result.Bytes, 0o644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				deps.BaseLogger.Info().Str("path", outputPath).Msg("Image saved")
			} else {
				deps.BaseLogger.Info().Int("bytes", len(result.Bytes)).Msg("Image generated (use --output to save)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name (e.g. gemini, openai)")
	_ = cmd.MarkFlagRequired("provider")
	cmd.Flags().StringVar(&modelName, "model", "", "Model name (optional)")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Image prompt")
	_ = cmd.MarkFlagRequired("prompt")
	cmd.Flags().StringVar(&imagePath, "image", "", "Input image path")
	_ = cmd.MarkFlagRequired("image")
	cmd.Flags().StringVar(&maskPath, "mask", "", "Input mask path (optional)")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output path for the generated image")

	return cmd
}
