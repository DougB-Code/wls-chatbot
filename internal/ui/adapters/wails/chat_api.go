// chat_api.go exposes chat endpoints to the frontend via the bridge.
// internal/ui/adapters/wails/chat_api.go
package wails

import (
	"fmt"

	chatfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/app/chat"
)

// CreateConversation creates a new conversation with the given settings.
func (b *Bridge) CreateConversation(providerName, model string) (*chatfeature.Conversation, error) {

	if b.app == nil || b.app.Conversations == nil {
		return nil, fmt.Errorf("chat orchestrator not configured")
	}
	conversation, err := b.app.Conversations.CreateConversation(providerName, model)
	if err != nil {
		return nil, err
	}
	return mapAppConversation(conversation), nil
}

// SetActiveConversation sets the active conversation by ID.
func (b *Bridge) SetActiveConversation(id string) {

	if b.app == nil || b.app.Conversations == nil {
		return
	}
	b.app.Conversations.SetActiveConversation(id)
}

// GetActiveConversation returns the currently active conversation.
func (b *Bridge) GetActiveConversation() *chatfeature.Conversation {

	if b.app == nil || b.app.Conversations == nil {
		return nil
	}
	return mapAppConversation(b.app.Conversations.GetActiveConversation())
}

// GetConversation returns a conversation by ID.
func (b *Bridge) GetConversation(id string) *chatfeature.Conversation {

	if b.app == nil || b.app.Conversations == nil {
		return nil
	}
	return mapAppConversation(b.app.Conversations.GetConversation(id))
}

// ListConversations returns summaries of all conversations.
func (b *Bridge) ListConversations() []chatfeature.ConversationSummary {

	if b.app == nil || b.app.Conversations == nil {
		return nil
	}
	return mapAppConversationSummaries(b.app.Conversations.ListConversations())
}

// ListDeletedConversations returns summaries of archived conversations.
func (b *Bridge) ListDeletedConversations() []chatfeature.ConversationSummary {

	if b.app == nil || b.app.Conversations == nil {
		return nil
	}
	return mapAppConversationSummaries(b.app.Conversations.ListDeletedConversations())
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
