// re-export port types for the provider use case.
// internal/core/usecase/provider/types.go
package provider

import "github.com/MadeByDoug/wls-chatbot/internal/core/ports"

type Provider = ports.Provider
type Config = ports.ProviderConfig
type ProviderMessage = ports.ProviderMessage
type ChatOptions = ports.ChatOptions
type Role = ports.Role
type Model = ports.Model
type Tool = ports.Tool
type Chunk = ports.Chunk
type UsageStats = ports.UsageStats

type Registry = ports.ProviderRegistry
type Cache = ports.ProviderCache
type CacheSnapshot = ports.ProviderCacheSnapshot
type CacheEntry = ports.ProviderCacheEntry
type SecretStore = ports.SecretStore
type Logger = ports.Logger
type LogField = ports.LogField
