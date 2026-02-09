// types.go re-exports chat domain and port contracts into the app boundary.
// internal/features/ai/chat/app/chat/types.go
package chat

import (
	chatdomain "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/domain"
	chatports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/ports"
)

type Conversation = chatdomain.Conversation
type ConversationSettings = chatdomain.ConversationSettings
type ConversationSummary = chatdomain.ConversationSummary
type Message = chatdomain.Message
type MessageMetadata = chatdomain.MessageMetadata
type Role = chatdomain.Role
type Block = chatdomain.Block
type BlockType = chatdomain.BlockType
type Artifact = chatdomain.Artifact
type ActionExecution = chatdomain.ActionExecution
type ActionStatus = chatdomain.ActionStatus

type Repository = chatports.ChatRepository

const (
	RoleUser      = chatdomain.RoleUser
	RoleAssistant = chatdomain.RoleAssistant
	RoleSystem    = chatdomain.RoleSystem
	RoleTool      = chatdomain.RoleTool
)

const (
	BlockTypeText     = chatdomain.BlockTypeText
	BlockTypeCode     = chatdomain.BlockTypeCode
	BlockTypeArtifact = chatdomain.BlockTypeArtifact
	BlockTypeThinking = chatdomain.BlockTypeThinking
	BlockTypeAction   = chatdomain.BlockTypeAction
	BlockTypeError    = chatdomain.BlockTypeError
	BlockTypeImage    = chatdomain.BlockTypeImage
)

const (
	ActionStatusPending   = chatdomain.ActionStatusPending
	ActionStatusApproved  = chatdomain.ActionStatusApproved
	ActionStatusRejected  = chatdomain.ActionStatusRejected
	ActionStatusRunning   = chatdomain.ActionStatusRunning
	ActionStatusCompleted = chatdomain.ActionStatusCompleted
	ActionStatusFailed    = chatdomain.ActionStatusFailed
)

// NewConversation builds a new conversation from domain settings.
func NewConversation(settings ConversationSettings) *Conversation {

	return chatdomain.NewConversation(settings)
}

// NewMessage builds a new domain message.
func NewMessage(conversationID string, role Role, content string) *Message {

	return chatdomain.NewMessage(conversationID, role, content)
}

// NewStreamingMessage builds a streaming placeholder message.
func NewStreamingMessage(conversationID string, role Role) *Message {

	return chatdomain.NewStreamingMessage(conversationID, role)
}

// StatusCodeFromErr extracts HTTP-like error status codes when available.
func StatusCodeFromErr(err error) int {

	return chatdomain.StatusCodeFromErr(err)
}
