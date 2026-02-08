// root_command.go constructs the CLI adapter command tree.
// internal/ui/adapters/cli/root_command.go
package cli

import (
	aicli "github.com/MadeByDoug/wls-chatbot/internal/ui/adapters/cli/ai"
	"github.com/spf13/cobra"
)

// NewCommand builds the CLI adapter command and all CLI subcommands.
func NewCommand(deps Dependencies) *cobra.Command {

	dependencyErr := deps.validate()
	if dependencyErr != nil {
		return &cobra.Command{
			Use:          "cli",
			SilenceUsage: true,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return dependencyErr
			},
		}
	}

	cmd := &cobra.Command{
		Use:          "cli",
		Short:        "Run CLI workflows",
		SilenceUsage: true,
	}

	cmd.AddCommand(aicli.NewCommand(aicli.Dependencies{
		Dependencies: deps.Dependencies,
	}))
	cmd.AddCommand(newTestCommand(deps.BaseLogger))

	return cmd
}
