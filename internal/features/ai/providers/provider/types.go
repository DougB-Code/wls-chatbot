// types.go re-exports provider contracts for the settings provider app service.
// internal/features/settings/app/provider/types.go
package provider

import (
	coreports "github.com/MadeByDoug/wls-chatbot/internal/core/logger"
	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/interfaces/core"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/interfaces/gateway"
)

type Provider = providercore.Provider
type Config = providercore.ProviderConfig
type ProviderMessage = providergateway.ProviderMessage
type ChatOptions = providergateway.ChatOptions
type Role = providergateway.Role
type Model = providercore.Model
type Tool = providergateway.Tool
type Chunk = providergateway.Chunk
type UsageStats = providergateway.UsageStats
type ImageGenerationOptions = providergateway.ImageGenerationOptions
type ImageEditOptions = providergateway.ImageEditOptions
type ImageResult = providergateway.ImageResult
type ImageData = providergateway.ImageData
type CredentialField = providercore.CredentialField
type ProviderCredentials = providercore.ProviderCredentials

type Registry = providercore.ProviderRegistry
type Cache = providercore.ProviderCache
type CacheSnapshot = providercore.ProviderCacheSnapshot
type CacheEntry = providercore.ProviderCacheEntry
type SecretStore = providercore.SecretStore
type InputsStore = providercore.ProviderInputsStore
type Logger = coreports.Logger
type LogField = coreports.LogField
