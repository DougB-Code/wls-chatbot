// expose chat endpoints to the frontend via the bridge.
// internal/adapters/wails/chat_api.go
package wails

import "github.com/MadeByDoug/wls-chatbot/internal/core/usecase/chat"

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

// UpdateConversationModel updates the model for a conversation.
func (b *Bridge) UpdateConversationModel(conversationID, model string) bool {

	return b.chat.UpdateConversationModel(conversationID, model)
}
