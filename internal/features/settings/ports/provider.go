// define contracts for AI model providers.
// internal/features/settings/ports/provider.go
package ports

import "context"

// Model represents an AI model.
type Model struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	ContextWindow     int    `json:"contextWindow"`
	SupportsStreaming bool   `json:"supportsStreaming"`
	SupportsTools     bool   `json:"supportsTools"`
	SupportsVision    bool   `json:"supportsVision"`
}

// ProviderConfig holds provider configuration.
type ProviderConfig struct {
	Name         string  `json:"name"`
	DisplayName  string  `json:"displayName"`
	APIKey       string  `json:"apiKey,omitempty"`
	BaseURL      string  `json:"baseUrl,omitempty"`
	DefaultModel string  `json:"defaultModel"`
	Models       []Model `json:"models"`
}

// ChatOptions configures a chat completion request.
type ChatOptions struct {
	Model       string   `json:"model"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int      `json:"maxTokens,omitempty"`
	Stream      bool     `json:"stream"`
	Tools       []Tool   `json:"tools,omitempty"`
	StopWords   []string `json:"stopWords,omitempty"`
}

// Role represents the sender of a provider message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// ProviderMessage represents a provider-ready chat message.
type ProviderMessage struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Tool represents a function the AI can call.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Chunk represents a piece of a streaming response.
type Chunk struct {
	Content      string      `json:"content,omitempty"`
	Model        string      `json:"model,omitempty"`
	ToolCalls    []ToolCall  `json:"toolCalls,omitempty"`
	FinishReason string      `json:"finishReason,omitempty"`
	Usage        *UsageStats `json:"usage,omitempty"`
	Error        error       `json:"-"`
}

// ToolCall represents an AI request to execute a tool.
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// UsageStats contains token usage information.
type UsageStats struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// Provider is the interface for AI model providers.
type Provider interface {
	Name() string
	DisplayName() string
	Models() []Model
	Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error)
	TestConnection(ctx context.Context) error
	Configure(config ProviderConfig) error
}
