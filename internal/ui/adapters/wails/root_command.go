// root_command.go constructs the Wails UI adapter command.
// internal/ui/adapters/wails/root_command.go
package wails

import (
	"fmt"
	"io/fs"

	commonadapter "github.com/MadeByDoug/wls-chatbot/internal/ui/adapters/common"
	"github.com/spf13/cobra"
)

// Dependencies groups construction functions required by the Wails adapter command.
type Dependencies struct {
	*commonadapter.Dependencies
	Assets fs.FS
}

// NewCommand builds the Wails adapter command.
func NewCommand(deps Dependencies) *cobra.Command {

	dependencyErr := deps.validate()
	if dependencyErr != nil {
		return &cobra.Command{
			Use:          "ui",
			SilenceUsage: true,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return dependencyErr
			},
		}
	}

	cmd := &cobra.Command{
		Use:          "ui",
		Aliases:      []string{"wails"},
		Short:        "Launch the Wails UI",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := deps.Dependencies.ValidateResolved(); err != nil {
				return fmt.Errorf("wails adapter: %w", err)
			}
			deps.BaseLogger.Info().Msg("Starting Wails Lit Starter ChatBot UI...")
			if err := runUI(deps.BaseLogger, deps.Config, deps.DB, deps.Assets, deps.AppName, deps.KeyringServiceName); err != nil {
				return fmt.Errorf("run UI: %w", err)
			}
			return nil
		},
	}

	return cmd
}

// validate returns an error when required Wails adapter dependencies are missing.
func (d Dependencies) validate() error {

	if d.Dependencies == nil {
		return fmt.Errorf("wails adapter: common dependencies required")
	}
	if err := d.Dependencies.ValidateCore(); err != nil {
		return fmt.Errorf("wails adapter: %w", err)
	}
	if d.Assets == nil {
		return fmt.Errorf("wails adapter: assets FS required")
	}
	return nil
}
