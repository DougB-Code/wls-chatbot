// re-export port types for the provider use case.
// internal/features/settings/usecase/types.go
package provider

import (
	coreports "github.com/MadeByDoug/wls-chatbot/internal/core/ports"
	settingsports "github.com/MadeByDoug/wls-chatbot/internal/features/settings/ports"
)

type Provider = settingsports.Provider
type Config = settingsports.ProviderConfig
type ProviderMessage = settingsports.ProviderMessage
type ChatOptions = settingsports.ChatOptions
type Role = settingsports.Role
type Model = settingsports.Model
type Tool = settingsports.Tool
type Chunk = settingsports.Chunk
type UsageStats = settingsports.UsageStats

type Registry = settingsports.ProviderRegistry
type Cache = settingsports.ProviderCache
type CacheSnapshot = settingsports.ProviderCacheSnapshot
type CacheEntry = settingsports.ProviderCacheEntry
type SecretStore = settingsports.SecretStore
type Logger = coreports.Logger
type LogField = coreports.LogField
