// chat_api.go exposes chat endpoints to the frontend via the bridge.
// internal/ui/adapters/wails/chat_api.go
package wails

import (
	"fmt"

	chatdomain "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/domain"
)

// CreateConversation creates a new conversation with the given settings.
func (b *Bridge) CreateConversation(providerName, model string) (*chatdomain.Conversation, error) {

	if b.app == nil || b.app.Conversations == nil {
		return nil, fmt.Errorf("chat orchestrator not configured")
	}
	return b.app.Conversations.CreateConversation(providerName, model)
}

// SetActiveConversation sets the active conversation by ID.
func (b *Bridge) SetActiveConversation(id string) {

	if b.app == nil || b.app.Conversations == nil {
		return
	}
	b.app.Conversations.SetActiveConversation(id)
}

// GetActiveConversation returns the currently active conversation.
func (b *Bridge) GetActiveConversation() *chatdomain.Conversation {

	if b.app == nil || b.app.Conversations == nil {
		return nil
	}
	return b.app.Conversations.GetActiveConversation()
}

// GetConversation returns a conversation by ID.
func (b *Bridge) GetConversation(id string) *chatdomain.Conversation {

	if b.app == nil || b.app.Conversations == nil {
		return nil
	}
	return b.app.Conversations.GetConversation(id)
}

// ListConversations returns summaries of all conversations.
func (b *Bridge) ListConversations() []chatdomain.ConversationSummary {

	if b.app == nil || b.app.Conversations == nil {
		return nil
	}
	return b.app.Conversations.ListConversations()
}

// ListDeletedConversations returns summaries of archived conversations.
func (b *Bridge) ListDeletedConversations() []chatdomain.ConversationSummary {

	if b.app == nil || b.app.Conversations == nil {
		return nil
	}
	return b.app.Conversations.ListDeletedConversations()
}

// UpdateConversationModel updates the model for a conversation.
func (b *Bridge) UpdateConversationModel(conversationID, model string) bool {

	if b.app == nil || b.app.Conversations == nil {
		return false
	}
	return b.app.Conversations.UpdateConversationModel(conversationID, model)
}

// UpdateConversationProvider updates the provider for a conversation.
func (b *Bridge) UpdateConversationProvider(conversationID, provider string) bool {

	if b.app == nil || b.app.Conversations == nil {
		return false
	}
	return b.app.Conversations.UpdateConversationProvider(conversationID, provider)
}

// DeleteConversation moves a conversation to the recycle bin.
func (b *Bridge) DeleteConversation(id string) bool {

	if b.app == nil || b.app.Conversations == nil {
		return false
	}
	return b.app.Conversations.DeleteConversation(id)
}

// RestoreConversation restores a recycled conversation.
func (b *Bridge) RestoreConversation(id string) bool {

	if b.app == nil || b.app.Conversations == nil {
		return false
	}
	return b.app.Conversations.RestoreConversation(id)
}

// PurgeConversation permanently deletes a conversation.
func (b *Bridge) PurgeConversation(id string) bool {

	if b.app == nil || b.app.Conversations == nil {
		return false
	}
	return b.app.Conversations.PurgeConversation(id)
}
