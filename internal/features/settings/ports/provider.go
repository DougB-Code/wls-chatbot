// define contracts for AI model providers.
// internal/features/settings/ports/provider.go
package ports

import (
	"context"

	coreports "github.com/MadeByDoug/wls-chatbot/internal/core/ports"
)

// Model represents an AI model.
type Model struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	ContextWindow     int    `json:"contextWindow"`
	SupportsStreaming bool   `json:"supportsStreaming"`
	SupportsTools     bool   `json:"supportsTools"`
	SupportsVision    bool   `json:"supportsVision"`
}

// ProviderCredentials stores credential values by field name.
type ProviderCredentials map[string]string

// CredentialField describes an input required to configure a provider.
type CredentialField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Secret      bool   `json:"secret"`
	Placeholder string `json:"placeholder,omitempty"`
	Help        string `json:"help,omitempty"`
}

const (
	CredentialAPIKey    = "api_key"
	CredentialAccountID = "account_id"
	CredentialGatewayID = "gateway_id"
	CredentialToken     = "token"
	CredentialCloudflareToken = "cloudflare_api_token"
)

// ProviderConfig holds provider configuration.
type ProviderConfig struct {
	Name         string  `json:"name"`
	DisplayName  string  `json:"displayName"`
	APIKey       string  `json:"apiKey,omitempty"`
	BaseURL      string  `json:"baseUrl,omitempty"`
	DefaultModel string  `json:"defaultModel"`
	Models       []Model `json:"models"`
	Credentials  ProviderCredentials `json:"credentials,omitempty"`
	Logger       coreports.Logger `json:"-"`
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
	CredentialFields() []CredentialField
	Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error)
	TestConnection(ctx context.Context) error
	Configure(config ProviderConfig) error
}
