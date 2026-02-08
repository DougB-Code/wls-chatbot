// model.go defines model transport contracts for backend adapters.
// internal/core/interfaces/ai/model.go
package ai

import "context"

// ProviderModelMutationInterface defines provider-model CRUD capabilities.
type ProviderModelMutationInterface interface {
	AddProviderModel(ctx context.Context, request AddProviderModelRequest) error
	UpdateProviderModel(ctx context.Context, request UpdateProviderModelRequest) error
	RemoveProviderModel(ctx context.Context, providerName string, modelID string) error
	UpdateProviderModelCapabilities(ctx context.Context, request UpdateProviderModelCapabilitiesRequest) error
}

// ProviderModelInterface defines model catalog capabilities shared across transports.
type ProviderModelInterface interface {
	ListModels(ctx context.Context, filter ModelListFilter) ([]ModelSummary, error)
	ImportModels(ctx context.Context, request ImportModelsRequest) error
	SyncModels(ctx context.Context) (SyncModelsResult, error)
}

// AddProviderModelRequest contains inputs for appending a model to a provider.
type AddProviderModelRequest struct {
	ProviderName string        `json:"providerName"`
	Model        ProviderModel `json:"model"`
}

// UpdateProviderModelRequest contains mutable model fields for a provider model.
type UpdateProviderModelRequest struct {
	ProviderName string              `json:"providerName"`
	ModelID      string              `json:"modelId"`
	Model        ProviderModelUpdate `json:"model"`
}

// UpdateProviderModelCapabilitiesRequest contains mutable capability fields for a provider model.
type UpdateProviderModelCapabilitiesRequest struct {
	ProviderName string                          `json:"providerName"`
	ModelID      string                          `json:"modelId"`
	Capabilities ProviderModelCapabilitiesUpdate `json:"capabilities"`
}

// ProviderModelUpdate describes mutable provider-model attributes.
type ProviderModelUpdate struct {
	Name          *string `json:"name,omitempty"`
	ContextWindow *int    `json:"contextWindow,omitempty"`
}

// ProviderModelCapabilitiesUpdate describes mutable provider-model capability flags.
type ProviderModelCapabilitiesUpdate struct {
	SupportsStreaming *bool `json:"supportsStreaming,omitempty"`
	SupportsTools     *bool `json:"supportsTools,omitempty"`
	SupportsVision    *bool `json:"supportsVision,omitempty"`
}

// ProviderModel describes a model exposed by a provider.
type ProviderModel struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	ContextWindow     int    `json:"contextWindow"`
	SupportsStreaming bool   `json:"supportsStreaming"`
	SupportsTools     bool   `json:"supportsTools"`
	SupportsVision    bool   `json:"supportsVision"`
}

// ModelListFilter defines model query filters.
type ModelListFilter struct {
	Source                   string   `json:"source,omitempty"`
	RequiredInputModalities  []string `json:"requiredInputModalities,omitempty"`
	RequiredOutputModalities []string `json:"requiredOutputModalities,omitempty"`
	RequiredCapabilityIDs    []string `json:"requiredCapabilityIds,omitempty"`
	RequiredSystemTags       []string `json:"requiredSystemTags,omitempty"`
	RequiresStreaming        *bool    `json:"requiresStreaming,omitempty"`
	RequiresToolCalling      *bool    `json:"requiresToolCalling,omitempty"`
	RequiresStructuredOutput *bool    `json:"requiresStructuredOutput,omitempty"`
	RequiresVision           *bool    `json:"requiresVision,omitempty"`
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
