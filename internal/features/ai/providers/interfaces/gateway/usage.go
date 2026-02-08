// usage.go defines gateway token usage contracts.
// internal/features/providers/interfaces/gateway/usage.go
package gateway

// UsageStats contains token usage information.
type UsageStats struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}
