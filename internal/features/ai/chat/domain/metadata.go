// metadata.go defines metadata captured from provider responses.
// internal/features/chat/domain/metadata.go
package domain

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
