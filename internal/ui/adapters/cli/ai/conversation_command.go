// conversation_command.go defines AI CLI adapters for conversation management workflows.
// internal/ui/adapters/cli/ai/conversation_command.go
package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// newConversationCommand creates the parent 'conversation' command.
func newConversationCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "conversation",
		Aliases: []string{"conv"},
		Short:   "Manage conversations",
	}
	cmd.AddCommand(newConversationListCommand(deps))
	cmd.AddCommand(newConversationListDeletedCommand(deps))
	cmd.AddCommand(newConversationCreateCommand(deps))
	cmd.AddCommand(newConversationGetCommand(deps))
	cmd.AddCommand(newConversationActiveCommand(deps))
	cmd.AddCommand(newConversationSetActiveCommand(deps))
	cmd.AddCommand(newConversationUpdateModelCommand(deps))
	cmd.AddCommand(newConversationUpdateProviderCommand(deps))
	cmd.AddCommand(newConversationDeleteCommand(deps))
	cmd.AddCommand(newConversationRestoreCommand(deps))
	cmd.AddCommand(newConversationPurgeCommand(deps))
	cmd.AddCommand(newConversationSendCommand(deps))
	cmd.AddCommand(newConversationStopCommand(deps))
	return cmd
}

// newConversationListCommand lists conversations.
func newConversationListCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List conversations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			summaries := applicationFacade.Conversations.ListConversations()
			if len(summaries) == 0 {
				fmt.Println("No conversations found.")
				return nil
			}

			fmt.Printf("%-36s %-20s %-10s\n", "ID", "TITLE", "MESSAGES")
			fmt.Println(strings.Repeat("-", 70))
			for _, summary := range summaries {
				title := summary.Title
				if len(title) > 20 {
					title = title[:17] + "..."
				}
				fmt.Printf("%-36s %-20s %-10d\n", summary.ID, title, summary.MessageCount)
			}
			return nil
		},
	}
	return cmd
}

// newConversationListDeletedCommand lists deleted conversations.
func newConversationListDeletedCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list-deleted",
		Short: "List deleted conversations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			summaries := applicationFacade.Conversations.ListDeletedConversations()
			if len(summaries) == 0 {
				fmt.Println("No deleted conversations found.")
				return nil
			}

			fmt.Printf("%-36s %-20s %-10s\n", "ID", "TITLE", "MESSAGES")
			fmt.Println(strings.Repeat("-", 70))
			for _, summary := range summaries {
				title := summary.Title
				if len(title) > 20 {
					title = title[:17] + "..."
				}
				fmt.Printf("%-36s %-20s %-10d\n", summary.ID, title, summary.MessageCount)
			}
			return nil
		},
	}
	return cmd
}

// newConversationCreateCommand creates a new conversation.
func newConversationCreateCommand(deps Dependencies) *cobra.Command {

	var providerName string
	var modelName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new conversation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			conversation, err := applicationFacade.Conversations.CreateConversation(providerName, modelName)
			if err != nil {
				return err
			}

			fmt.Printf("Created conversation: %s\n", conversation.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name")
	_ = cmd.MarkFlagRequired("provider")
	cmd.Flags().StringVar(&modelName, "model", "", "Model name")
	_ = cmd.MarkFlagRequired("model")
	return cmd
}

// newConversationGetCommand gets a conversation by ID.
func newConversationGetCommand(deps Dependencies) *cobra.Command {

	var id string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get conversation details",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			conversation := applicationFacade.Conversations.GetConversation(id)
			if conversation == nil {
				return fmt.Errorf("conversation not found: %s", id)
			}

			fmt.Printf("ID:       %s\n", conversation.ID)
			fmt.Printf("Title:    %s\n", conversation.Title)
			fmt.Printf("Provider: %s\n", conversation.Settings.Provider)
			fmt.Printf("Model:    %s\n", conversation.Settings.Model)
			fmt.Printf("Messages: %d\n", len(conversation.Messages))
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Conversation ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

// newConversationActiveCommand shows the active conversation.
func newConversationActiveCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "active",
		Short: "Show the active conversation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			conversation := applicationFacade.Conversations.GetActiveConversation()
			if conversation == nil {
				fmt.Println("No active conversation set.")
				return nil
			}

			fmt.Printf("Active conversation: %s (%s)\n", conversation.ID, conversation.Title)
			return nil
		},
	}
	return cmd
}

// newConversationSetActiveCommand sets the active conversation.
func newConversationSetActiveCommand(deps Dependencies) *cobra.Command {

	var id string

	cmd := &cobra.Command{
		Use:   "set-active",
		Short: "Set the active conversation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			applicationFacade.Conversations.SetActiveConversation(id)
			fmt.Printf("Active conversation set to %s.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Conversation ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

// newConversationUpdateModelCommand updates a conversation's model.
func newConversationUpdateModelCommand(deps Dependencies) *cobra.Command {

	var id string
	var modelName string

	cmd := &cobra.Command{
		Use:   "update-model",
		Short: "Update conversation model",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			if !applicationFacade.Conversations.UpdateConversationModel(id, modelName) {
				return fmt.Errorf("failed to update model for conversation: %s", id)
			}

			fmt.Printf("Model updated to %s.\n", modelName)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Conversation ID")
	_ = cmd.MarkFlagRequired("id")
	cmd.Flags().StringVar(&modelName, "model", "", "Model name")
	_ = cmd.MarkFlagRequired("model")
	return cmd
}

// newConversationUpdateProviderCommand updates a conversation's provider.
func newConversationUpdateProviderCommand(deps Dependencies) *cobra.Command {

	var id string
	var providerName string

	cmd := &cobra.Command{
		Use:   "update-provider",
		Short: "Update conversation provider",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			if !applicationFacade.Conversations.UpdateConversationProvider(id, providerName) {
				return fmt.Errorf("failed to update provider for conversation: %s", id)
			}

			fmt.Printf("Provider updated to %s.\n", providerName)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Conversation ID")
	_ = cmd.MarkFlagRequired("id")
	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name")
	_ = cmd.MarkFlagRequired("provider")
	return cmd
}

// newConversationDeleteCommand deletes a conversation.
func newConversationDeleteCommand(deps Dependencies) *cobra.Command {

	var id string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a conversation (soft delete)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			if !applicationFacade.Conversations.DeleteConversation(id) {
				return fmt.Errorf("failed to delete conversation: %s", id)
			}

			fmt.Printf("Conversation %s deleted.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Conversation ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

// newConversationRestoreCommand restores a deleted conversation.
func newConversationRestoreCommand(deps Dependencies) *cobra.Command {

	var id string

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore a deleted conversation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			if !applicationFacade.Conversations.RestoreConversation(id) {
				return fmt.Errorf("failed to restore conversation: %s", id)
			}

			fmt.Printf("Conversation %s restored.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Conversation ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

// newConversationPurgeCommand permanently deletes a conversation.
func newConversationPurgeCommand(deps Dependencies) *cobra.Command {

	var id string

	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Permanently delete a conversation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			if !applicationFacade.Conversations.PurgeConversation(id) {
				return fmt.Errorf("failed to purge conversation: %s", id)
			}

			fmt.Printf("Conversation %s permanently deleted.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Conversation ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

// newConversationSendCommand sends a message to a conversation.
func newConversationSendCommand(deps Dependencies) *cobra.Command {

	var id string
	var content string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a message to a conversation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			message, err := applicationFacade.Conversations.SendMessage(context.Background(), id, content)
			if err != nil {
				return err
			}

			content := ""
			if len(message.Blocks) > 0 {
				content = message.Blocks[0].Content
			}
			fmt.Printf("Message sent. Response:\n%s\n", content)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Conversation ID")
	_ = cmd.MarkFlagRequired("id")
	cmd.Flags().StringVar(&content, "content", "", "Message content")
	_ = cmd.MarkFlagRequired("content")
	return cmd
}

// newConversationStopCommand stops any active stream.
func newConversationStopCommand(deps Dependencies) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the active stream",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			applicationFacade, err := loadApp(deps)
			if err != nil {
				return err
			}

			applicationFacade.Conversations.StopStream()
			fmt.Println("Stream stopped.")
			return nil
		},
	}
	return cmd
}
