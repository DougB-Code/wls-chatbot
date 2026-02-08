// chat_facade.go adapts stateless chat service capabilities into app contracts.
// internal/app/chat_facade.go
package app

import (
	"context"
	"fmt"

	"github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
	chatfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/ports"
)

// NewChatCompletion adapts chat service operations into app contracts.
func NewChatCompletion(chat chatfeature.ChatInterface) ChatCompletion {

	return &chatCompletion{chat: chat}
}

// chatCompletion exposes chat service operations through app contracts.
type chatCompletion struct {
	chat chatfeature.ChatInterface
}

// Chat streams chat completion chunks.
func (c *chatCompletion) Chat(ctx context.Context, request contracts.ChatRequest) (<-chan contracts.ChatChunk, error) {

	if c.chat == nil {
		return nil, fmt.Errorf("app chat: service not configured")
	}

	chunks, err := c.chat.Chat(ctx, chatfeature.ChatRequest{
		ProviderName: request.ProviderName,
		ModelName:    request.ModelName,
		Messages:     mapChatMessages(request.Messages),
		Options: chatfeature.ChatOptions{
			Temperature: request.Options.Temperature,
			MaxTokens:   request.Options.MaxTokens,
			Stream:      request.Options.Stream,
			StopWords:   request.Options.StopWords,
			Tools:       mapChatTools(request.Options.Tools),
		},
	})
	if err != nil {
		return nil, err
	}

	mapped := make(chan contracts.ChatChunk)
	go func() {
		defer close(mapped)
		for chunk := range chunks {
			mapped <- mapChatChunk(chunk)
		}
	}()
	return mapped, nil
}

// mapChatMessages converts app chat messages into feature chat messages.
func mapChatMessages(messages []contracts.ChatMessage) []chatfeature.ChatMessage {

	if len(messages) == 0 {
		return nil
	}

	mapped := make([]chatfeature.ChatMessage, 0, len(messages))
	for _, message := range messages {
		mapped = append(mapped, chatfeature.ChatMessage{
			Role:    chatfeature.ChatRole(message.Role),
			Content: message.Content,
		})
	}
	return mapped
}

// mapChatTools converts app chat tools into feature chat tools.
func mapChatTools(tools []contracts.ChatTool) []chatfeature.ChatTool {

	if len(tools) == 0 {
		return nil
	}

	mapped := make([]chatfeature.ChatTool, 0, len(tools))
	for _, tool := range tools {
		mapped = append(mapped, chatfeature.ChatTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}
	return mapped
}

// mapChatChunk converts feature chat chunks into app chat chunks.
func mapChatChunk(chunk chatfeature.ChatChunk) contracts.ChatChunk {

	var usage *contracts.ChatUsage
	if chunk.Usage != nil {
		usage = &contracts.ChatUsage{
			InputTokens:  chunk.Usage.InputTokens,
			OutputTokens: chunk.Usage.OutputTokens,
			TotalTokens:  chunk.Usage.TotalTokens,
		}
	}

	toolCalls := make([]contracts.ChatToolCall, 0, len(chunk.ToolCalls))
	for _, call := range chunk.ToolCalls {
		toolCalls = append(toolCalls, contracts.ChatToolCall{
			ID:        call.ID,
			Name:      call.Name,
			Arguments: call.Arguments,
		})
	}
	if len(toolCalls) == 0 {
		toolCalls = nil
	}

	return contracts.ChatChunk{
		Content:      chunk.Content,
		Model:        chunk.Model,
		ToolCalls:    toolCalls,
		FinishReason: chunk.FinishReason,
		Usage:        usage,
		Error:        chunk.Error,
	}
}
