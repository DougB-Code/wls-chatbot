// backend.go defines the shared backend interface for CLI and Wails adapters.
// internal/core/interfaces/backend.go
package interfaces

import (
	"context"

	provider "github.com/MadeByDoug/wls-chatbot/internal/features/settings/usecase"
)

// ProviderInterface defines provider capabilities shared across transports.
type ProviderInterface interface {
	GetProviders(ctx context.Context) ([]provider.Info, error)
	TestProvider(ctx context.Context, name string) error
}

// ImageInterface defines image generation capabilities shared across transports.
type ImageInterface interface {
	GenerateImage(ctx context.Context, request GenerateImageRequest) (ImageBinaryResult, error)
	EditImage(ctx context.Context, request EditImageRequest) (ImageBinaryResult, error)
}

// ModelInterface defines model catalog capabilities shared across transports.
type ModelInterface interface {
	ListModels(ctx context.Context, filter ModelListFilter) ([]ModelSummary, error)
	ImportModels(ctx context.Context, request ImportModelsRequest) error
	SyncModels(ctx context.Context) (SyncModelsResult, error)
}

// Backend defines the unified backend capability surface for adapters.
type Backend interface {
	ProviderInterface
	ImageInterface
	ModelInterface
}

// GenerateImageRequest contains image generation inputs.
type GenerateImageRequest struct {
	ProviderName   string `json:"providerName"`
	ModelName      string `json:"modelName,omitempty"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Style          string `json:"style,omitempty"`
	ResponseFormat string `json:"responseFormat,omitempty"`
	User           string `json:"user,omitempty"`
}

// EditImageRequest contains image edit inputs.
type EditImageRequest struct {
	ProviderName string `json:"providerName"`
	ModelName    string `json:"modelName,omitempty"`
	Prompt       string `json:"prompt"`
	ImagePath    string `json:"imagePath"`
	MaskPath     string `json:"maskPath,omitempty"`
	N            int    `json:"n,omitempty"`
	Size         string `json:"size,omitempty"`
}

// ImageBinaryResult contains binary image output metadata.
type ImageBinaryResult struct {
	Bytes         []byte `json:"bytes"`
	RevisedPrompt string `json:"revisedPrompt,omitempty"`
}

// ModelListFilter defines model query filters.
type ModelListFilter struct {
	Source                     string   `json:"source,omitempty"`
	RequiredInputModalities    []string `json:"requiredInputModalities,omitempty"`
	RequiredOutputModalities   []string `json:"requiredOutputModalities,omitempty"`
	RequiredCapabilityIDs      []string `json:"requiredCapabilityIds,omitempty"`
	RequiredSystemTags         []string `json:"requiredSystemTags,omitempty"`
	RequiresStreaming          *bool    `json:"requiresStreaming,omitempty"`
	RequiresToolCalling        *bool    `json:"requiresToolCalling,omitempty"`
	RequiresStructuredOutput   *bool    `json:"requiresStructuredOutput,omitempty"`
	RequiresVision             *bool    `json:"requiresVision,omitempty"`
}

// ModelCapabilities contains model feature and semantic capability metadata.
type ModelCapabilities struct {
	SupportsStreaming        bool     `json:"supportsStreaming"`
	SupportsToolCalling      bool     `json:"supportsToolCalling"`
	SupportsStructuredOutput bool     `json:"supportsStructuredOutput"`
	SupportsVision           bool     `json:"supportsVision"`
	InputModalities          []string `json:"inputModalities"`
	OutputModalities         []string `json:"outputModalities"`
	CapabilityIDs            []string `json:"capabilityIds"`
	SystemTags               []string `json:"systemTags,omitempty"`
}

// ModelSummary contains model listing fields used by adapters.
type ModelSummary struct {
	ID                string            `json:"id"`
	ModelID           string            `json:"modelId"`
	DisplayName       string            `json:"displayName"`
	ProviderName      string            `json:"providerName"`
	Source            string            `json:"source"`
	Approved          bool              `json:"approved"`
	AvailabilityState string            `json:"availabilityState"`
	ContextWindow     int               `json:"contextWindow"`
	CostTier          string            `json:"costTier"`
	Capabilities      ModelCapabilities `json:"capabilities"`
}

// ImportModelsRequest contains model import inputs.
type ImportModelsRequest struct {
	FilePath string `json:"filePath"`
}

// SyncModelsResult contains sync output metadata.
type SyncModelsResult struct {
	Path     string `json:"path"`
	Imported bool   `json:"imported"`
}
