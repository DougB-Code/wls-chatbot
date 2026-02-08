// chat_facade.go adapts stateless chat service capabilities into app interfaces.
// internal/app/chat_facade.go
package app

import (
	"context"
	"fmt"

	chatports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/ports"
)

// NewChatCompletion adapts chat service operations into app interfaces.
func NewChatCompletion(chat chatports.ChatInterface) ChatCompletion {

	return &chatCompletion{chat: chat}
}

// chatCompletion exposes chat service operations through app interfaces.
type chatCompletion struct {
	chat chatports.ChatInterface
}

// Chat streams chat completion chunks.
func (c *chatCompletion) Chat(ctx context.Context, request chatports.ChatRequest) (<-chan chatports.ChatChunk, error) {

	if c.chat == nil {
		return nil, fmt.Errorf("app chat: service not configured")
	}

	return c.chat.Chat(ctx, request)
}
