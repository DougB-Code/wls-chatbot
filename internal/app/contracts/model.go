// model.go defines canonical model catalog DTOs for the application facade.
// internal/app/contracts/model.go
package contracts

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

// ModelSummary contains model listing fields used by transport adapters.
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
