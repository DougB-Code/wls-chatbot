// stream.go exposes chat streaming endpoints to the frontend via the bridge.
// internal/ui/adapters/wails/stream.go
package wails

import (
	"fmt"

	chatfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/app/chat"
)

// SendMessage sends a user message and initiates a streaming response.
func (b *Bridge) SendMessage(conversationID, content string) (*chatfeature.Message, error) {

	if b.app == nil || b.app.Conversations == nil {
		return nil, fmt.Errorf("chat orchestrator not configured")
	}
	message, err := b.app.Conversations.SendMessage(b.ctxOrBackground(), conversationID, content)
	if err != nil {
		return nil, err
	}
	return mapAppMessage(message), nil
}

// StopStream cancels the currently running stream.
func (b *Bridge) StopStream() {

	if b.app == nil || b.app.Conversations == nil {
		return
	}
	b.app.Conversations.StopStream()
}
