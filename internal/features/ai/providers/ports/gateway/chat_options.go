// chat_options.go defines chat request configuration contracts.
// internal/features/ai/providers/ports/gateway/chat_options.go
package gateway

// ChatOptions configures a chat completion request.
type ChatOptions struct {
	Model       string   `json:"model"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int      `json:"maxTokens,omitempty"`
	Stream      bool     `json:"stream"`
	Tools       []Tool   `json:"tools,omitempty"`
	StopWords   []string `json:"stopWords,omitempty"`
}
