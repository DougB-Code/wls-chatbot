// alias settings port types for provider adapters.
// internal/features/settings/adapters/provider/types.go
package provider

import (
	coreports "github.com/MadeByDoug/wls-chatbot/internal/core/ports"
	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/ports"
)

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
type CredentialField = ports.CredentialField
type ProviderCredentials = ports.ProviderCredentials
type CacheEntry = ports.ProviderCacheEntry
type CacheSnapshot = ports.ProviderCacheSnapshot
type Logger = coreports.Logger
type LogField = coreports.LogField

const (
	RoleUser      = ports.RoleUser
	RoleAssistant = ports.RoleAssistant
	RoleSystem    = ports.RoleSystem
	RoleTool      = ports.RoleTool
	CredentialAPIKey    = ports.CredentialAPIKey
	CredentialAccountID = ports.CredentialAccountID
	CredentialGatewayID = ports.CredentialGatewayID
	CredentialToken     = ports.CredentialToken
	CredentialCloudflareToken = ports.CredentialCloudflareToken
)
