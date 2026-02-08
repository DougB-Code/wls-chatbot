// messages.go defines provider message role and payload contracts.
// internal/features/providers/interfaces/gateway/messages.go
package gateway

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
