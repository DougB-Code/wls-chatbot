// conversation_facade.go adapts chat orchestration capabilities into app conversation contracts.
// internal/app/conversation_facade.go
package app

import (
	"context"
	"fmt"

	"github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
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
func (m *conversationManagement) CreateConversation(providerName, model string) (*contracts.Conversation, error) {

	if m.orchestrator == nil {
		return nil, fmt.Errorf("app conversations: orchestrator not configured")
	}

	conversation, err := m.orchestrator.CreateConversation(providerName, model)
	if err != nil {
		return nil, err
	}
	return mapConversation(conversation), nil
}

// SetActiveConversation marks a conversation as active.
func (m *conversationManagement) SetActiveConversation(id string) {

	if m.orchestrator == nil {
		return
	}
	m.orchestrator.SetActiveConversation(id)
}

// GetActiveConversation returns the active conversation.
func (m *conversationManagement) GetActiveConversation() *contracts.Conversation {

	if m.orchestrator == nil {
		return nil
	}
	return mapConversation(m.orchestrator.GetActiveConversation())
}

// GetConversation returns a conversation by ID.
func (m *conversationManagement) GetConversation(id string) *contracts.Conversation {

	if m.orchestrator == nil {
		return nil
	}
	return mapConversation(m.orchestrator.GetConversation(id))
}

// ListConversations returns active conversation summaries.
func (m *conversationManagement) ListConversations() []contracts.ConversationSummary {

	if m.orchestrator == nil {
		return nil
	}
	return mapConversationSummaries(m.orchestrator.ListConversations())
}

// ListDeletedConversations returns archived conversation summaries.
func (m *conversationManagement) ListDeletedConversations() []contracts.ConversationSummary {

	if m.orchestrator == nil {
		return nil
	}
	return mapConversationSummaries(m.orchestrator.ListDeletedConversations())
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
func (m *conversationManagement) SendMessage(ctx context.Context, conversationID, content string) (*contracts.Message, error) {

	if m.orchestrator == nil {
		return nil, fmt.Errorf("app conversations: orchestrator not configured")
	}

	message, err := m.orchestrator.SendMessage(ctx, conversationID, content)
	if err != nil {
		return nil, err
	}
	return mapMessage(message), nil
}

// StopStream cancels the active stream.
func (m *conversationManagement) StopStream() {

	if m.orchestrator == nil {
		return
	}
	m.orchestrator.StopStream()
}

// mapConversation converts feature conversation DTOs into app conversation DTOs.
func mapConversation(conversation *chatfeature.Conversation) *contracts.Conversation {

	if conversation == nil {
		return nil
	}

	return &contracts.Conversation{
		ID:    conversation.ID,
		Title: conversation.Title,
		Messages: func() []*contracts.Message {
			if len(conversation.Messages) == 0 {
				return nil
			}
			mapped := make([]*contracts.Message, 0, len(conversation.Messages))
			for _, message := range conversation.Messages {
				mapped = append(mapped, mapMessage(message))
			}
			return mapped
		}(),
		Settings: contracts.ConversationSettings{
			Provider:     conversation.Settings.Provider,
			Model:        conversation.Settings.Model,
			Temperature:  conversation.Settings.Temperature,
			MaxTokens:    conversation.Settings.MaxTokens,
			SystemPrompt: conversation.Settings.SystemPrompt,
		},
		CreatedAt:  conversation.CreatedAt,
		UpdatedAt:  conversation.UpdatedAt,
		IsArchived: conversation.IsArchived,
	}
}

// mapConversationSummaries converts feature conversation summaries into app conversation summaries.
func mapConversationSummaries(summaries []chatfeature.ConversationSummary) []contracts.ConversationSummary {

	if len(summaries) == 0 {
		return nil
	}

	mapped := make([]contracts.ConversationSummary, 0, len(summaries))
	for _, summary := range summaries {
		mapped = append(mapped, contracts.ConversationSummary{
			ID:           summary.ID,
			Title:        summary.Title,
			LastMessage:  summary.LastMessage,
			MessageCount: summary.MessageCount,
			UpdatedAt:    summary.UpdatedAt,
		})
	}
	return mapped
}

// mapMessage converts feature message DTOs into app message DTOs.
func mapMessage(message *chatfeature.Message) *contracts.Message {

	if message == nil {
		return nil
	}

	return &contracts.Message{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Role:           contracts.Role(message.Role),
		Blocks:         mapBlocks(message.Blocks),
		Timestamp:      message.Timestamp,
		IsStreaming:    message.IsStreaming,
		Metadata:       mapMessageMetadata(message.Metadata),
	}
}

// mapBlocks converts feature message blocks into app message blocks.
func mapBlocks(blocks []chatfeature.Block) []contracts.Block {

	if len(blocks) == 0 {
		return nil
	}

	mapped := make([]contracts.Block, 0, len(blocks))
	for _, block := range blocks {
		mapped = append(mapped, contracts.Block{
			Type:        contracts.BlockType(block.Type),
			Content:     block.Content,
			Language:    block.Language,
			Artifact:    mapArtifact(block.Artifact),
			Action:      mapActionExecution(block.Action),
			IsCollapsed: block.IsCollapsed,
		})
	}
	return mapped
}

// mapArtifact converts feature artifact DTOs into app artifact DTOs.
func mapArtifact(artifact *chatfeature.Artifact) *contracts.Artifact {

	if artifact == nil {
		return nil
	}

	return &contracts.Artifact{
		ID:        artifact.ID,
		Name:      artifact.Name,
		Type:      artifact.Type,
		Content:   artifact.Content,
		Language:  artifact.Language,
		Version:   artifact.Version,
		CreatedAt: artifact.CreatedAt,
		UpdatedAt: artifact.UpdatedAt,
	}
}

// mapActionExecution converts feature action DTOs into app action DTOs.
func mapActionExecution(action *chatfeature.ActionExecution) *contracts.ActionExecution {

	if action == nil {
		return nil
	}

	mappedArgs := make(map[string]interface{}, len(action.Args))
	for key, value := range action.Args {
		mappedArgs[key] = value
	}
	if len(mappedArgs) == 0 {
		mappedArgs = nil
	}

	return &contracts.ActionExecution{
		ID:          action.ID,
		ToolName:    action.ToolName,
		Description: action.Description,
		Args:        mappedArgs,
		Status:      contracts.ActionStatus(action.Status),
		Result:      action.Result,
		StartedAt:   action.StartedAt,
		CompletedAt: action.CompletedAt,
	}
}

// mapMessageMetadata converts feature message metadata into app message metadata.
func mapMessageMetadata(metadata *chatfeature.MessageMetadata) *contracts.MessageMetadata {

	if metadata == nil {
		return nil
	}

	return &contracts.MessageMetadata{
		Provider:     metadata.Provider,
		Model:        metadata.Model,
		TokensIn:     metadata.TokensIn,
		TokensOut:    metadata.TokensOut,
		TokensTotal:  metadata.TokensTotal,
		LatencyMs:    metadata.LatencyMs,
		FinishReason: metadata.FinishReason,
		StatusCode:   metadata.StatusCode,
		ErrorMessage: metadata.ErrorMessage,
	}
}
