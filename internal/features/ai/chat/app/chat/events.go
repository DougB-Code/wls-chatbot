// events.go registers strongly typed chat event signals for frontend transport.
// internal/features/ai/chat/app/chat/events.go
package chat

import coreevents "github.com/MadeByDoug/wls-chatbot/internal/core/events"

// MessageEventPayload represents message creation and stream-start event payloads.
type MessageEventPayload struct {
	ConversationID string   `json:"conversationId"`
	MessageID      string   `json:"messageId"`
	Timestamp      int64    `json:"ts"`
	Message        *Message `json:"message,omitempty"`
}

// StreamChunkEventPayload represents streaming update, error, and completion payloads.
type StreamChunkEventPayload struct {
	ConversationID string           `json:"conversationId"`
	MessageID      string           `json:"messageId"`
	Timestamp      int64            `json:"ts"`
	BlockIndex     int              `json:"blockIndex"`
	Content        string           `json:"content"`
	IsDone         bool             `json:"isDone"`
	Metadata       *MessageMetadata `json:"metadata,omitempty"`
	Error          string           `json:"error,omitempty"`
	StatusCode     int              `json:"statusCode,omitempty"`
}

// ConversationTitleEventPayload represents conversation title update payloads.
type ConversationTitleEventPayload struct {
	ConversationID string `json:"conversationId"`
	Timestamp      int64  `json:"ts"`
	Title          string `json:"title"`
}

var (
	SignalMessageCreated    = coreevents.MustRegister[MessageEventPayload]("chat.message")
	SignalStreamStarted     = coreevents.MustRegister[MessageEventPayload]("chat.stream.start")
	SignalStreamChunk       = coreevents.MustRegister[StreamChunkEventPayload]("chat.stream.chunk")
	SignalStreamError       = coreevents.MustRegister[StreamChunkEventPayload]("chat.stream.error")
	SignalStreamCompleted   = coreevents.MustRegister[StreamChunkEventPayload]("chat.stream.complete")
	SignalConversationTitle = coreevents.MustRegister[ConversationTitleEventPayload]("chat.conversation.title")
)
