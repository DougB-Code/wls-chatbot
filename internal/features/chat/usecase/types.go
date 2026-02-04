// re-export domain types and ports for the chat use case.
// internal/features/chat/usecase/types.go
package chat

import (
	chatdomain "github.com/MadeByDoug/wls-chatbot/internal/features/chat/domain"
	chatports "github.com/MadeByDoug/wls-chatbot/internal/features/chat/ports"
)

type Conversation = chatdomain.Conversation
type ConversationSettings = chatdomain.ConversationSettings
type ConversationSummary = chatdomain.ConversationSummary
type Message = chatdomain.Message
type MessageMetadata = chatdomain.MessageMetadata
type Role = chatdomain.Role
type Block = chatdomain.Block
type BlockType = chatdomain.BlockType

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
