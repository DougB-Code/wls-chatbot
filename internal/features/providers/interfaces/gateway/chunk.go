// chunk.go defines streaming chunk response contracts.
// internal/features/providers/interfaces/gateway/chunk.go
package gateway

// Chunk represents a piece of a streaming response.
type Chunk struct {
	Content      string      `json:"content,omitempty"`
	Model        string      `json:"model,omitempty"`
	ToolCalls    []ToolCall  `json:"toolCalls,omitempty"`
	FinishReason string      `json:"finishReason,omitempty"`
	Usage        *UsageStats `json:"usage,omitempty"`
	Error        error       `json:"-"`
}
