// expose chat streaming endpoints to the frontend via the bridge.
// internal/core/adapters/wails/stream.go
package wails

import (
	"github.com/MadeByDoug/wls-chatbot/internal/features/chat/usecase"
)

// SendMessage sends a user message and initiates a streaming response.
func (b *Bridge) SendMessage(conversationID, content string) (*chat.Message, error) {

	return b.chat.SendMessage(b.ctxOrBackground(), conversationID, content)
}

// StopStream cancels the currently running stream.
func (b *Bridge) StopStream() {

	b.chat.StopStream()
}
