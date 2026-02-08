// chat_command.go defines AI CLI adapters for chat workflows.
// internal/ui/adapters/cli/ai/chat_command.go
package ai

import (
	"context"
	"fmt"

	appcontracts "github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
	"github.com/spf13/cobra"
)

// newChatCommand creates the parent 'chat' command.
func newChatCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Run chat completions",
	}
	cmd.AddCommand(newChatSendCommand(deps))
	return cmd
}

// newChatSendCommand creates the 'chat send' command.
func newChatSendCommand(deps Dependencies) *cobra.Command {

	var providerName string
	var modelName string
	var prompt string
	var systemPrompt string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a chat prompt",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			messages := make([]appcontracts.ChatMessage, 0, 2)
			if systemPrompt != "" {
				messages = append(messages, appcontracts.ChatMessage{
					Role:    appcontracts.ChatRoleSystem,
					Content: systemPrompt,
				})
			}
			messages = append(messages, appcontracts.ChatMessage{
				Role:    appcontracts.ChatRoleUser,
				Content: prompt,
			})

			chunks, err := applicationFacade.Chat.Chat(context.Background(), appcontracts.ChatRequest{
				ProviderName: providerName,
				ModelName:    modelName,
				Messages:     messages,
				Options: appcontracts.ChatOptions{
					Stream: true,
				},
			})
			if err != nil {
				return err
			}

			for chunk := range chunks {
				if chunk.Error != "" {
					return fmt.Errorf("%s", chunk.Error)
				}
				if chunk.Content == "" {
					continue
				}
				fmt.Print(chunk.Content)
			}
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name")
	_ = cmd.MarkFlagRequired("provider")
	cmd.Flags().StringVar(&modelName, "model", "", "Model name")
	_ = cmd.MarkFlagRequired("model")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Prompt text")
	_ = cmd.MarkFlagRequired("prompt")
	cmd.Flags().StringVar(&systemPrompt, "system", "", "Optional system prompt")
	return cmd
}
