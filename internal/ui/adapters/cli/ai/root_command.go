// root_command.go constructs the AI CLI adapter command tree.
// internal/ui/adapters/cli/ai/root_command.go
package ai

import "github.com/spf13/cobra"

// NewCommand builds the AI CLI adapter command and all AI CLI subcommands.
func NewCommand(deps Dependencies) *cobra.Command {

	dependencyErr := deps.validate()
	if dependencyErr != nil {
		return &cobra.Command{
			Use:          "ai",
			SilenceUsage: true,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return dependencyErr
			},
		}
	}

	cmd := &cobra.Command{
		Use:          "ai",
		Short:        "Run AI workflows",
		SilenceUsage: true,
	}

	cmd.AddCommand(newProviderCommand(deps))
	cmd.AddCommand(newModelCommand(deps))
	cmd.AddCommand(newImageCommand(deps))
	cmd.AddCommand(newChatCommand(deps))
	cmd.AddCommand(newConversationCommand(deps))

	return cmd
}
