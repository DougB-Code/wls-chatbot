// expose chat endpoints to the frontend via the bridge.
// internal/core/adapters/wails/chat_api.go
package wails

import "github.com/MadeByDoug/wls-chatbot/internal/features/chat/usecase"

// CreateConversation creates a new conversation with the given settings.
func (b *Bridge) CreateConversation(providerName, model string) *chat.Conversation {

	return b.chat.CreateConversation(providerName, model)
}

// SetActiveConversation sets the active conversation by ID.
func (b *Bridge) SetActiveConversation(id string) {

	b.chat.SetActiveConversation(id)
}

// GetActiveConversation returns the currently active conversation.
func (b *Bridge) GetActiveConversation() *chat.Conversation {

	return b.chat.GetActiveConversation()
}

// GetConversation returns a conversation by ID.
func (b *Bridge) GetConversation(id string) *chat.Conversation {

	return b.chat.GetConversation(id)
}

// ListConversations returns summaries of all conversations.
func (b *Bridge) ListConversations() []chat.ConversationSummary {

	return b.chat.ListConversations()
}

// ListDeletedConversations returns summaries of archived conversations.
func (b *Bridge) ListDeletedConversations() []chat.ConversationSummary {

	return b.chat.ListDeletedConversations()
}

// UpdateConversationModel updates the model for a conversation.
func (b *Bridge) UpdateConversationModel(conversationID, model string) bool {

	return b.chat.UpdateConversationModel(conversationID, model)
}

// UpdateConversationProvider updates the provider for a conversation.
func (b *Bridge) UpdateConversationProvider(conversationID, provider string) bool {

	return b.chat.UpdateConversationProvider(conversationID, provider)
}

// DeleteConversation moves a conversation to the recycle bin.
func (b *Bridge) DeleteConversation(id string) bool {

	return b.chat.DeleteConversation(id)
}

// RestoreConversation restores a recycled conversation.
func (b *Bridge) RestoreConversation(id string) bool {

	return b.chat.RestoreConversation(id)
}

// PurgeConversation permanently deletes a conversation.
func (b *Bridge) PurgeConversation(id string) bool {

	return b.chat.PurgeConversation(id)
}
