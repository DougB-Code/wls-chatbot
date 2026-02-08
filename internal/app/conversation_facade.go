// conversation_facade.go adapts chat orchestration capabilities into app interfaces.
// internal/app/conversation_facade.go
package app

import (
	"context"
	"fmt"

	chatfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/app/chat"
)

// NewConversationManagement adapts chat orchestration operations into app operations.
func NewConversationManagement(orchestrator *chatfeature.Orchestrator) ConversationManagement {

	return &conversationManagement{orchestrator: orchestrator}
}

// conversationManagement exposes chat orchestration through app operations.
type conversationManagement struct {
	orchestrator *chatfeature.Orchestrator
}

// CreateConversation creates a new conversation.
func (m *conversationManagement) CreateConversation(providerName, model string) (*chatfeature.Conversation, error) {

	if m.orchestrator == nil {
		return nil, fmt.Errorf("app conversations: orchestrator not configured")
	}

	return m.orchestrator.CreateConversation(providerName, model)
}

// SetActiveConversation marks a conversation as active.
func (m *conversationManagement) SetActiveConversation(id string) {

	if m.orchestrator == nil {
		return
	}
	m.orchestrator.SetActiveConversation(id)
}

// GetActiveConversation returns the active conversation.
func (m *conversationManagement) GetActiveConversation() *chatfeature.Conversation {

	if m.orchestrator == nil {
		return nil
	}
	return m.orchestrator.GetActiveConversation()
}

// GetConversation returns a conversation by ID.
func (m *conversationManagement) GetConversation(id string) *chatfeature.Conversation {

	if m.orchestrator == nil {
		return nil
	}
	return m.orchestrator.GetConversation(id)
}

// ListConversations returns active conversation summaries.
func (m *conversationManagement) ListConversations() []chatfeature.ConversationSummary {

	if m.orchestrator == nil {
		return nil
	}
	return m.orchestrator.ListConversations()
}

// ListDeletedConversations returns archived conversation summaries.
func (m *conversationManagement) ListDeletedConversations() []chatfeature.ConversationSummary {

	if m.orchestrator == nil {
		return nil
	}
	return m.orchestrator.ListDeletedConversations()
}

// UpdateConversationModel updates a conversation model.
func (m *conversationManagement) UpdateConversationModel(conversationID, model string) bool {

	if m.orchestrator == nil {
		return false
	}
	return m.orchestrator.UpdateConversationModel(conversationID, model)
}

// UpdateConversationProvider updates a conversation provider.
func (m *conversationManagement) UpdateConversationProvider(conversationID, provider string) bool {

	if m.orchestrator == nil {
		return false
	}
	return m.orchestrator.UpdateConversationProvider(conversationID, provider)
}

// DeleteConversation archives a conversation.
func (m *conversationManagement) DeleteConversation(id string) bool {

	if m.orchestrator == nil {
		return false
	}
	return m.orchestrator.DeleteConversation(id)
}

// RestoreConversation restores an archived conversation.
func (m *conversationManagement) RestoreConversation(id string) bool {

	if m.orchestrator == nil {
		return false
	}
	return m.orchestrator.RestoreConversation(id)
}

// PurgeConversation permanently removes a conversation.
func (m *conversationManagement) PurgeConversation(id string) bool {

	if m.orchestrator == nil {
		return false
	}
	return m.orchestrator.PurgeConversation(id)
}

// SendMessage appends a message and triggers streaming.
func (m *conversationManagement) SendMessage(ctx context.Context, conversationID, content string) (*chatfeature.Message, error) {

	if m.orchestrator == nil {
		return nil, fmt.Errorf("app conversations: orchestrator not configured")
	}

	return m.orchestrator.SendMessage(ctx, conversationID, content)
}

// StopStream cancels the active stream.
func (m *conversationManagement) StopStream() {

	if m.orchestrator == nil {
		return
	}
	m.orchestrator.StopStream()
}
