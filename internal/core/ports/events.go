// define event payloads and emitter contracts.
// internal/core/ports/events.go
package ports

import "github.com/MadeByDoug/wls-chatbot/internal/features/chat/domain"

// ChatEvent is the wrapper for all events sent to the frontend.
type ChatEvent struct {
	Type           string      `json:"type"`
	ConversationID string      `json:"conversationId"`
	MessageID      string      `json:"messageId,omitempty"`
	Timestamp      int64       `json:"ts"`
	Payload        interface{} `json:"payload,omitempty"`
}

// StreamChunkPayload is the payload for stream chunk events.
type StreamChunkPayload struct {
	BlockIndex int                   `json:"blockIndex"`
	Content    string                `json:"content"`
	IsDone     bool                  `json:"isDone"`
	Metadata   *chat.MessageMetadata `json:"metadata,omitempty"`
	Error      string                `json:"error,omitempty"`
	StatusCode int                   `json:"statusCode,omitempty"`
}

// ConversationTitlePayload carries a conversation title update.
type ConversationTitlePayload struct {
	Title string `json:"title"`
}

// ToastPayload describes a toast notification.
type ToastPayload struct {
	Type    string `json:"type,omitempty"`
	Title   string `json:"title,omitempty"`
	Message string `json:"message,omitempty"`
}

// ChatEmitter publishes chat-related events to the frontend.
type ChatEmitter interface {
	EmitChatEvent(event ChatEvent)
}

// ProviderEmitter publishes provider-related events to the frontend.
type ProviderEmitter interface {
	EmitProvidersUpdated()
}

// CatalogEmitter publishes catalog-related events to the frontend.
type CatalogEmitter interface {
	EmitCatalogUpdated()
}

// ToastEmitter publishes toast notifications to the frontend.
type ToastEmitter interface {
	EmitToast(payload ToastPayload)
}

// Emitter bundles all event emitters used by policies.
type Emitter interface {
	ChatEmitter
	ProviderEmitter
	CatalogEmitter
	ToastEmitter
}
