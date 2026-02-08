// chat_service.go provides chat backend operations.
// internal/core/backend/ai/chat_service.go
package chat

import (
	"context"
	"fmt"

	aiinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/ports"
)

// ChatService handles chat operations for transport adapters.
type ChatService struct{}

var _ aiinterfaces.ChatInterface = (*ChatService)(nil)

// NewChatService creates a chat backend service.
func NewChatService() *ChatService {

	return &ChatService{}
}

// Chat streams chat completion chunks.
func (s *ChatService) Chat(context.Context, aiinterfaces.ChatRequest) (<-chan aiinterfaces.ChatChunk, error) {

	return nil, fmt.Errorf("backend service: chat not configured")
}
