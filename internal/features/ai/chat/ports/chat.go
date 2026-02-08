// chat.go defines chat completion transport contracts.
// internal/core/interfaces/ai/chat.go
package ports

import "context"

// ChatInterface defines chat completion capabilities shared across transports.
type ChatInterface interface {
	Chat(ctx context.Context, request ChatRequest) (<-chan ChatChunk, error)
}

// ChatRequest contains inputs for a chat completion request.
type ChatRequest struct {
	ProviderName string        `json:"providerName"`
	ModelName    string        `json:"modelName"`
	Messages     []ChatMessage `json:"messages"`
	Options      ChatOptions   `json:"options,omitempty"`
}

// ChatMessage represents a single chat message payload.
type ChatMessage struct {
	Role    ChatRole `json:"role"`
	Content string   `json:"content"`
}

// ChatRole represents the sender role for chat messages.
type ChatRole string

const (
	ChatRoleUser      ChatRole = "user"
	ChatRoleAssistant ChatRole = "assistant"
	ChatRoleSystem    ChatRole = "system"
	ChatRoleTool      ChatRole = "tool"
)

// ChatOptions configures chat request behavior.
type ChatOptions struct {
	Temperature float64    `json:"temperature,omitempty"`
	MaxTokens   int        `json:"maxTokens,omitempty"`
	Stream      bool       `json:"stream,omitempty"`
	StopWords   []string   `json:"stopWords,omitempty"`
	Tools       []ChatTool `json:"tools,omitempty"`
}

// ChatTool describes a tool available for model tool-calling.
type ChatTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// ChatChunk represents one streamed chat response chunk.
type ChatChunk struct {
	Content      string         `json:"content,omitempty"`
	Model        string         `json:"model,omitempty"`
	ToolCalls    []ChatToolCall `json:"toolCalls,omitempty"`
	FinishReason string         `json:"finishReason,omitempty"`
	Usage        *ChatUsage     `json:"usage,omitempty"`
	Error        string         `json:"error,omitempty"`
}

// ChatToolCall represents a model-requested tool invocation.
type ChatToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ChatUsage contains token accounting for a chat request.
type ChatUsage struct {
	InputTokens  int `json:"inputTokens,omitempty"`
	OutputTokens int `json:"outputTokens,omitempty"`
	TotalTokens  int `json:"totalTokens,omitempty"`
}
