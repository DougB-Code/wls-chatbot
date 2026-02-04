// alias settings port types for provider adapters.
// internal/features/settings/adapters/provider/types.go
package provider

import "github.com/MadeByDoug/wls-chatbot/internal/features/settings/ports"

type Provider = ports.Provider
type Config = ports.ProviderConfig
type Model = ports.Model
type ChatOptions = ports.ChatOptions
type ProviderMessage = ports.ProviderMessage
type Role = ports.Role
type Tool = ports.Tool
type Chunk = ports.Chunk
type ToolCall = ports.ToolCall
type UsageStats = ports.UsageStats
type CacheEntry = ports.ProviderCacheEntry
type CacheSnapshot = ports.ProviderCacheSnapshot

const (
	RoleUser      = ports.RoleUser
	RoleAssistant = ports.RoleAssistant
	RoleSystem    = ports.RoleSystem
	RoleTool      = ports.RoleTool
)
