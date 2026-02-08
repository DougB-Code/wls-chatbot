// conversation.go defines canonical conversation and message DTOs for the application facade.
// internal/app/contracts/conversation.go
package contracts

// ConversationSettings holds the configuration for a conversation.
type ConversationSettings struct {
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	Temperature  float64 `json:"temperature,omitempty"`
	MaxTokens    int     `json:"maxTokens,omitempty"`
	SystemPrompt string  `json:"systemPrompt,omitempty"`
}

// Conversation represents a chat conversation.
type Conversation struct {
	ID         string               `json:"id"`
	Title      string               `json:"title"`
	Messages   []*Message           `json:"messages"`
	Settings   ConversationSettings `json:"settings"`
	CreatedAt  int64                `json:"createdAt"`
	UpdatedAt  int64                `json:"updatedAt"`
	IsArchived bool                 `json:"isArchived"`
}

// ConversationSummary is a lightweight representation for listing.
type ConversationSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	LastMessage  string `json:"lastMessage,omitempty"`
	MessageCount int    `json:"messageCount"`
	UpdatedAt    int64  `json:"updatedAt"`
}

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
	Type      string `json:"type"`
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

// MessageMetadata contains information about message generation.
type MessageMetadata struct {
	Provider     string `json:"provider,omitempty"`
	Model        string `json:"model,omitempty"`
	TokensIn     int    `json:"tokensIn,omitempty"`
	TokensOut    int    `json:"tokensOut,omitempty"`
	TokensTotal  int    `json:"tokensTotal,omitempty"`
	LatencyMs    int64  `json:"latencyMs,omitempty"`
	FinishReason string `json:"finishReason,omitempty"`
	StatusCode   int    `json:"statusCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
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
