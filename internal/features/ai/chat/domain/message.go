// message.go defines chat message entities and content blocks.
// internal/features/ai/chat/domain/message.go
package domain

import (
	"time"
)

// Role represents the sender of a message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// BlockType represents the type of content in a message block.
type BlockType string

const (
	BlockTypeText     BlockType = "text"
	BlockTypeCode     BlockType = "code"
	BlockTypeArtifact BlockType = "artifact"
	BlockTypeThinking BlockType = "thinking"
	BlockTypeAction   BlockType = "action"
	BlockTypeError    BlockType = "error"
	BlockTypeImage    BlockType = "image"
)

// ActionStatus represents the status of a tool action.
type ActionStatus string

const (
	ActionStatusPending   ActionStatus = "pending"
	ActionStatusApproved  ActionStatus = "approved"
	ActionStatusRejected  ActionStatus = "rejected"
	ActionStatusRunning   ActionStatus = "running"
	ActionStatusCompleted ActionStatus = "completed"
	ActionStatusFailed    ActionStatus = "failed"
)

// ActionExecution represents a tool call and its execution state.
type ActionExecution struct {
	ID          string                 `json:"id"`
	ToolName    string                 `json:"toolName"`
	Description string                 `json:"description"`
	Args        map[string]interface{} `json:"args"`
	Status      ActionStatus           `json:"status"`
	Result      string                 `json:"result,omitempty"`
	StartedAt   int64                  `json:"startedAt,omitempty"`
	CompletedAt int64                  `json:"completedAt,omitempty"`
}

// Artifact represents a generated document, code file, or other content.
type Artifact struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"` // document, code, diagram, image, data
	Content   string `json:"content"`
	Language  string `json:"language,omitempty"`
	Version   int    `json:"version"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

// Block represents a content block within a message.
type Block struct {
	Type        BlockType        `json:"type"`
	Content     string           `json:"content"`
	Language    string           `json:"language,omitempty"`
	Artifact    *Artifact        `json:"artifact,omitempty"`
	Action      *ActionExecution `json:"action,omitempty"`
	IsCollapsed bool             `json:"isCollapsed,omitempty"`
}

// Message represents a single message in a conversation.
type Message struct {
	ID             string           `json:"id"`
	ConversationID string           `json:"conversationId"`
	Role           Role             `json:"role"`
	Blocks         []Block          `json:"blocks"`
	Timestamp      int64            `json:"timestamp"`
	IsStreaming    bool             `json:"isStreaming,omitempty"`
	Metadata       *MessageMetadata `json:"metadata,omitempty"`
}

// NewMessage creates a new message with the given role and content.
func NewMessage(conversationID string, role Role, content string) *Message {

	return &Message{
		ID:             generateID(),
		ConversationID: conversationID,
		Role:           role,
		Blocks: []Block{
			{
				Type:    BlockTypeText,
				Content: content,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewStreamingMessage creates a new message that will receive streaming content.
func NewStreamingMessage(conversationID string, role Role) *Message {

	return &Message{
		ID:             generateID(),
		ConversationID: conversationID,
		Role:           role,
		Blocks:         []Block{},
		Timestamp:      time.Now().UnixMilli(),
		IsStreaming:    true,
	}
}
