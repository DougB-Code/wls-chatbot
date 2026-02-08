// chat_mappers.go maps app conversation contracts into Wails chat transport DTOs.
// internal/ui/adapters/wails/chat_mappers.go
package wails

import (
	appcontracts "github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
	chatfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/app/chat"
)

// mapAppConversation converts app conversation DTOs into Wails chat conversation DTOs.
func mapAppConversation(conversation *appcontracts.Conversation) *chatfeature.Conversation {

	if conversation == nil {
		return nil
	}

	return &chatfeature.Conversation{
		ID:    conversation.ID,
		Title: conversation.Title,
		Messages: func() []*chatfeature.Message {
			if len(conversation.Messages) == 0 {
				return nil
			}
			mapped := make([]*chatfeature.Message, 0, len(conversation.Messages))
			for _, message := range conversation.Messages {
				mapped = append(mapped, mapAppMessage(message))
			}
			return mapped
		}(),
		Settings: chatfeature.ConversationSettings{
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

// mapAppConversationSummaries converts app conversation summaries into Wails chat summaries.
func mapAppConversationSummaries(summaries []appcontracts.ConversationSummary) []chatfeature.ConversationSummary {

	if len(summaries) == 0 {
		return nil
	}

	mapped := make([]chatfeature.ConversationSummary, 0, len(summaries))
	for _, summary := range summaries {
		mapped = append(mapped, chatfeature.ConversationSummary{
			ID:           summary.ID,
			Title:        summary.Title,
			LastMessage:  summary.LastMessage,
			MessageCount: summary.MessageCount,
			UpdatedAt:    summary.UpdatedAt,
		})
	}
	return mapped
}

// mapAppMessage converts app message DTOs into Wails chat message DTOs.
func mapAppMessage(message *appcontracts.Message) *chatfeature.Message {

	if message == nil {
		return nil
	}

	return &chatfeature.Message{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Role:           chatfeature.Role(message.Role),
		Blocks:         mapAppBlocks(message.Blocks),
		Timestamp:      message.Timestamp,
		IsStreaming:    message.IsStreaming,
		Metadata:       mapAppMessageMetadata(message.Metadata),
	}
}

// mapAppBlocks converts app message blocks into Wails chat blocks.
func mapAppBlocks(blocks []appcontracts.Block) []chatfeature.Block {

	if len(blocks) == 0 {
		return nil
	}

	mapped := make([]chatfeature.Block, 0, len(blocks))
	for _, block := range blocks {
		mapped = append(mapped, chatfeature.Block{
			Type:        chatfeature.BlockType(block.Type),
			Content:     block.Content,
			Language:    block.Language,
			Artifact:    mapAppArtifact(block.Artifact),
			Action:      mapAppActionExecution(block.Action),
			IsCollapsed: block.IsCollapsed,
		})
	}
	return mapped
}

// mapAppArtifact converts app artifact DTOs into Wails chat artifact DTOs.
func mapAppArtifact(artifact *appcontracts.Artifact) *chatfeature.Artifact {

	if artifact == nil {
		return nil
	}

	return &chatfeature.Artifact{
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

// mapAppActionExecution converts app action DTOs into Wails chat action DTOs.
func mapAppActionExecution(action *appcontracts.ActionExecution) *chatfeature.ActionExecution {

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

	return &chatfeature.ActionExecution{
		ID:          action.ID,
		ToolName:    action.ToolName,
		Description: action.Description,
		Args:        mappedArgs,
		Status:      chatfeature.ActionStatus(action.Status),
		Result:      action.Result,
		StartedAt:   action.StartedAt,
		CompletedAt: action.CompletedAt,
	}
}

// mapAppMessageMetadata converts app message metadata into Wails chat metadata.
func mapAppMessageMetadata(metadata *appcontracts.MessageMetadata) *chatfeature.MessageMetadata {

	if metadata == nil {
		return nil
	}

	return &chatfeature.MessageMetadata{
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
