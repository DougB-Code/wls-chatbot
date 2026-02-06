// types.go defines model catalog and role data transfer objects.
// internal/features/catalog/usecase/types.go
package catalog

// ProviderSummary represents provider metadata for the catalog.
type ProviderSummary struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DisplayName     string `json:"displayName"`
	AdapterType     string `json:"adapterType"`
	TrustMode       string `json:"trustMode"`
	BaseURL         string `json:"baseUrl"`
	LastTestAt      int64  `json:"lastTestAt"`
	LastTestOK      bool   `json:"lastTestOk"`
	LastError       string `json:"lastError,omitempty"`
	LastDiscoveryAt int64  `json:"lastDiscoveryAt"`
}

// EndpointSummary represents an endpoint with its discovered models.
type EndpointSummary struct {
	ID                  string         `json:"id"`
	ProviderID          string         `json:"providerId"`
	ProviderName        string         `json:"providerName"`
	ProviderDisplayName string         `json:"providerDisplayName"`
	DisplayName         string         `json:"displayName"`
	AdapterType         string         `json:"adapterType"`
	BaseURL             string         `json:"baseUrl"`
	RouteKind           string         `json:"routeKind"`
	OriginProvider      string         `json:"originProvider"`
	OriginRouteLabel    string         `json:"originRouteLabel"`
	LastTestAt          int64          `json:"lastTestAt"`
	LastTestOK          bool           `json:"lastTestOk"`
	LastError           string         `json:"lastError,omitempty"`
	Models              []ModelSummary `json:"models"`
}

// ModelSummary describes a catalog model and intrinsic capabilities.
type ModelSummary struct {
	ID                       string   `json:"id"`
	EndpointID               string   `json:"endpointId"`
	ModelID                  string   `json:"modelId"`
	DisplayName              string   `json:"displayName"`
	AvailabilityState        string   `json:"availabilityState"`
	ContextWindow            int      `json:"contextWindow"`
	CostTier                 string   `json:"costTier"`
	SupportsStreaming        bool     `json:"supportsStreaming"`
	SupportsToolCalling      bool     `json:"supportsToolCalling"`
	SupportsStructuredOutput bool     `json:"supportsStructuredOutput"`
	SupportsVision           bool     `json:"supportsVision"`
	InputModalities          []string `json:"inputModalities"`
	OutputModalities         []string `json:"outputModalities"`
}

// RoleRequirements describes required modalities and features for a role.
type RoleRequirements struct {
	RequiredInputModalities  []string `json:"requiredInputModalities"`
	RequiredOutputModalities []string `json:"requiredOutputModalities"`
	RequiresStreaming        bool     `json:"requiresStreaming"`
	RequiresToolCalling      bool     `json:"requiresToolCalling"`
	RequiresStructuredOutput bool     `json:"requiresStructuredOutput"`
	RequiresVision           bool     `json:"requiresVision"`
}

// RoleConstraints describe optional constraints for a role.
type RoleConstraints struct {
	MaxCostTier        string `json:"maxCostTier,omitempty"`
	MaxLatencyTier     string `json:"maxLatencyTier,omitempty"`
	MinReliabilityTier string `json:"minReliabilityTier,omitempty"`
}

// RoleSummary describes a role and its assignments.
type RoleSummary struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Requirements RoleRequirements        `json:"requirements"`
	Constraints  RoleConstraints         `json:"constraints"`
	Assignments  []RoleAssignmentSummary `json:"assignments"`
}

// RoleAssignmentSummary describes a role assignment entry.
type RoleAssignmentSummary struct {
	RoleID              string `json:"roleId"`
	ModelCatalogEntryID string `json:"modelCatalogEntryId"`
	ModelLabel          string `json:"modelLabel"`
	AssignedBy          string `json:"assignedBy"`
	CreatedAt           int64  `json:"createdAt"`
	Enabled             bool   `json:"enabled"`
}

// RoleAssignmentResult describes assignment validation errors.
type RoleAssignmentResult struct {
	MissingModalities []string `json:"missingModalities,omitempty"`
	MissingFeatures   []string `json:"missingFeatures,omitempty"`
}

// CatalogOverview combines endpoints and roles for the settings UI.
type CatalogOverview struct {
	Providers []ProviderSummary `json:"providers"`
	Endpoints []EndpointSummary `json:"endpoints"`
	Roles     []RoleSummary     `json:"roles"`
}
