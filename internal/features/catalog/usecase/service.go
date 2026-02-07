// service.go coordinates model catalog discovery, endpoints, and roles.
// internal/features/catalog/usecase/service.go
package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	coreports "github.com/MadeByDoug/wls-chatbot/internal/core/ports"
	"github.com/MadeByDoug/wls-chatbot/internal/features/catalog/adapters/catalogrepo"
	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/config"
	providerusecase "github.com/MadeByDoug/wls-chatbot/internal/features/settings/usecase"
)

// ProviderService describes provider operations required by the catalog.
type ProviderService interface {
	List() []providerusecase.Info
	RefreshResources(ctx context.Context, name string) error
	GetResources(name string) []providerusecase.Model
	TestConnection(ctx context.Context, name string) error
}

// Service manages catalog discovery and role assignments.
type Service struct {
	repo      *catalogrepo.Repository
	providers ProviderService
	cfg       config.AppConfig
	logger    coreports.Logger
}

// defaultRoleRecords defines app-provided role contracts seeded into the catalog.
var defaultRoleRecords = []catalogrepo.RoleRecord{
	{
		Name:                    "text_summarization",
		RequiredInputModalities: []string{"text"},
		RequiredOutputModalities: []string{"text"},
	},
	{
		Name:                    "video_transcription",
		RequiredInputModalities: []string{"video"},
		RequiredOutputModalities: []string{"text"},
	},
	{
		Name:                    "audio_transcription",
		RequiredInputModalities: []string{"audio"},
		RequiredOutputModalities: []string{"text"},
	},
}

// NewService creates a catalog service with required dependencies.
func NewService(repo *catalogrepo.Repository, providers ProviderService, cfg config.AppConfig, logger coreports.Logger) *Service {

	return &Service{
		repo:      repo,
		providers: providers,
		cfg:       cfg,
		logger:    logger,
	}
}

// RefreshAll refreshes models for all configured endpoints.
func (s *Service) RefreshAll(ctx context.Context) error {

	if s == nil {
		return fmt.Errorf("catalog service: missing dependencies")
	}
	if err := s.ensureDefaultRoles(ctx); err != nil {
		return err
	}

	providerInfos := s.indexProviders()
	refreshErrors := make([]error, 0)
	for _, providerConfig := range s.cfg.Providers {
		info := providerInfos[providerConfig.Name]
		if err := s.refreshProvider(ctx, providerConfig, info); err != nil {
			s.logWarn("Catalog refresh provider failed", err, coreports.LogField{Key: "provider", Value: providerConfig.Name})
			refreshErrors = append(refreshErrors, fmt.Errorf("%s: %w", providerConfig.Name, err))
		}
	}
	if len(refreshErrors) > 0 {
		return fmt.Errorf("catalog refresh failed for %d provider(s): %w", len(refreshErrors), errors.Join(refreshErrors...))
	}
	return nil
}

// RefreshEndpoint refreshes models for the endpoint's provider.
func (s *Service) RefreshEndpoint(ctx context.Context, endpointID string) error {

	if s == nil || s.repo == nil {
		return fmt.Errorf("catalog service: repo required")
	}

	endpoint, err := s.repo.GetEndpoint(ctx, endpointID)
	if err != nil {
		return err
	}
	if endpoint.ID == "" {
		return fmt.Errorf("catalog service: endpoint not found")
	}

	providerConfig, ok := s.findProviderConfig(endpoint.ProviderName)
	if !ok {
		return fmt.Errorf("catalog service: provider not found for endpoint")
	}
	providerInfos := s.indexProviders()
	info := providerInfos[providerConfig.Name]
	return s.refreshProvider(ctx, providerConfig, info)
}

// TestEndpoint performs a connectivity test for the endpoint's provider.
func (s *Service) TestEndpoint(ctx context.Context, endpointID string) error {

	if s == nil || s.repo == nil {
		return fmt.Errorf("catalog service: repo required")
	}

	endpoint, err := s.repo.GetEndpoint(ctx, endpointID)
	if err != nil {
		return err
	}
	if endpoint.ID == "" {
		return fmt.Errorf("catalog service: endpoint not found")
	}

	if err := s.providers.TestConnection(ctx, endpoint.ProviderName); err != nil {
		if statusErr := s.repo.UpdateEndpointStatus(ctx, endpointID, time.Now().UnixMilli(), false, err.Error()); statusErr != nil {
			s.logWarn("Catalog endpoint status update failed", statusErr, coreports.LogField{Key: "endpoint", Value: endpointID})
		}
		return &CatalogError{Code: ErrorCodeProviderAuthFailure, Message: err.Error(), Cause: err}
	}
	if statusErr := s.repo.UpdateEndpointStatus(ctx, endpointID, time.Now().UnixMilli(), true, ""); statusErr != nil {
		s.logWarn("Catalog endpoint status update failed", statusErr, coreports.LogField{Key: "endpoint", Value: endpointID})
	}
	return nil
}

// GetOverview returns catalog providers, endpoints, and roles.
func (s *Service) GetOverview(ctx context.Context) (CatalogOverview, error) {

	if s == nil || s.repo == nil {
		return CatalogOverview{}, fmt.Errorf("catalog service: repo required")
	}
	if err := s.ensureDefaultRoles(ctx); err != nil {
		return CatalogOverview{}, err
	}
	if err := s.ensureConnectedProviderCatalog(ctx); err != nil {
		return CatalogOverview{}, err
	}

	providers, err := s.repo.ListProviders(ctx)
	if err != nil {
		return CatalogOverview{}, err
	}
	endpoints, err := s.repo.ListEndpoints(ctx)
	if err != nil {
		return CatalogOverview{}, err
	}
	models, err := s.repo.ListModelSummaries(ctx)
	if err != nil {
		return CatalogOverview{}, err
	}
	roles, err := s.repo.ListRoles(ctx)
	if err != nil {
		return CatalogOverview{}, err
	}
	assignments, err := s.repo.ListRoleAssignments(ctx)
	if err != nil {
		return CatalogOverview{}, err
	}
	modelLabels, err := s.repo.ListModelLabels(ctx)
	if err != nil {
		return CatalogOverview{}, err
	}

	modelByEndpoint := make(map[string][]ModelSummary)
	for _, model := range models {
		modelByEndpoint[model.EndpointID] = append(modelByEndpoint[model.EndpointID], mapModelSummary(model))
	}
	for endpointID := range modelByEndpoint {
		sort.Slice(modelByEndpoint[endpointID], func(i, j int) bool {
			return modelByEndpoint[endpointID][i].ModelID < modelByEndpoint[endpointID][j].ModelID
		})
	}

	providerMap := make(map[string]catalogrepo.ProviderRecord)
	for _, provider := range providers {
		providerMap[provider.ID] = provider
	}

	endpointSummaries := make([]EndpointSummary, 0, len(endpoints))
	for _, endpoint := range endpoints {
		provider := providerMap[endpoint.ProviderID]
		endpointSummaries = append(endpointSummaries, EndpointSummary{
			ID:                  endpoint.ID,
			ProviderID:          endpoint.ProviderID,
			ProviderName:        endpoint.ProviderName,
			ProviderDisplayName: provider.DisplayName,
			DisplayName:         endpoint.DisplayName,
			AdapterType:         endpoint.AdapterType,
			BaseURL:             endpoint.BaseURL,
			RouteKind:           endpoint.RouteKind,
			OriginProvider:      endpoint.OriginProvider,
			OriginRouteLabel:    endpoint.OriginRouteLabel,
			LastTestAt:          endpoint.LastTestAt,
			LastTestOK:          endpoint.LastTestOK,
			LastError:           endpoint.LastError,
			Models:              modelByEndpoint[endpoint.ID],
		})
	}

	roleAssignments := make(map[string][]RoleAssignmentSummary)
	for _, assignment := range assignments {
		roleAssignments[assignment.RoleID] = append(roleAssignments[assignment.RoleID], RoleAssignmentSummary{
			RoleID:              assignment.RoleID,
			ModelCatalogEntryID: assignment.ModelCatalogEntryID,
			ModelLabel:          modelLabels[assignment.ModelCatalogEntryID],
			AssignedBy:          assignment.AssignedBy,
			CreatedAt:           assignment.CreatedAt,
			Enabled:             assignment.Enabled,
		})
	}

	roleSummaries := make([]RoleSummary, 0, len(roles))
	for _, role := range roles {
		roleSummaries = append(roleSummaries, RoleSummary{
			ID:   role.ID,
			Name: role.Name,
			Requirements: RoleRequirements{
				RequiredInputModalities:  role.RequiredInputModalities,
				RequiredOutputModalities: role.RequiredOutputModalities,
				RequiresStreaming:        role.RequiresStreaming,
				RequiresToolCalling:      role.RequiresToolCalling,
				RequiresStructuredOutput: role.RequiresStructuredOutput,
				RequiresVision:           role.RequiresVision,
			},
			Constraints: RoleConstraints{
				MaxCostTier:        role.MaxCostTier,
				MaxLatencyTier:     role.MaxLatencyTier,
				MinReliabilityTier: role.MinReliabilityTier,
			},
			Assignments: roleAssignments[role.ID],
		})
	}

	providerSummaries := make([]ProviderSummary, 0, len(providers))
	for _, provider := range providers {
		providerSummaries = append(providerSummaries, ProviderSummary{
			ID:              provider.ID,
			Name:            provider.Name,
			DisplayName:     provider.DisplayName,
			AdapterType:     provider.AdapterType,
			TrustMode:       provider.TrustMode,
			BaseURL:         provider.BaseURL,
			LastTestAt:      provider.LastTestAt,
			LastTestOK:      provider.LastTestOK,
			LastError:       provider.LastError,
			LastDiscoveryAt: provider.LastDiscoveryAt,
		})
	}

	return CatalogOverview{
		Providers: providerSummaries,
		Endpoints: endpointSummaries,
		Roles:     roleSummaries,
	}, nil
}

// ensureConnectedProviderCatalog backfills catalog endpoints for connected providers.
func (s *Service) ensureConnectedProviderCatalog(ctx context.Context) error {

	if s == nil || s.repo == nil {
		return fmt.Errorf("catalog service: repo required")
	}

	endpoints, err := s.repo.ListEndpoints(ctx)
	if err != nil {
		return err
	}

	hasEndpointByProvider := make(map[string]struct{}, len(endpoints))
	for _, endpoint := range endpoints {
		hasEndpointByProvider[endpoint.ProviderName] = struct{}{}
	}

	providerInfos := s.indexProviders()
	for _, providerConfig := range s.cfg.Providers {
		info := providerInfos[providerConfig.Name]
		if !info.IsConnected {
			continue
		}
		if _, exists := hasEndpointByProvider[providerConfig.Name]; exists {
			continue
		}

		if err := s.refreshProvider(ctx, providerConfig, info); err != nil {
			s.logWarn("Catalog connected provider bootstrap failed", err, coreports.LogField{Key: "provider", Value: providerConfig.Name})
			continue
		}
		hasEndpointByProvider[providerConfig.Name] = struct{}{}
	}

	return nil
}

// SaveRole creates or updates a role definition.
func (s *Service) SaveRole(ctx context.Context, summary RoleSummary) (RoleSummary, error) {

	if s == nil || s.repo == nil {
		return RoleSummary{}, fmt.Errorf("catalog service: repo required")
	}

	record, err := s.repo.UpsertRole(ctx, catalogrepo.RoleRecord{
		ID:                       summary.ID,
		Name:                     summary.Name,
		RequiresStreaming:        summary.Requirements.RequiresStreaming,
		RequiresToolCalling:      summary.Requirements.RequiresToolCalling,
		RequiresStructuredOutput: summary.Requirements.RequiresStructuredOutput,
		RequiresVision:           summary.Requirements.RequiresVision,
		MaxCostTier:              summary.Constraints.MaxCostTier,
		MaxLatencyTier:           summary.Constraints.MaxLatencyTier,
		MinReliabilityTier:       summary.Constraints.MinReliabilityTier,
		RequiredInputModalities:  summary.Requirements.RequiredInputModalities,
		RequiredOutputModalities: summary.Requirements.RequiredOutputModalities,
	})
	if err != nil {
		return RoleSummary{}, err
	}

	return RoleSummary{
		ID:   record.ID,
		Name: record.Name,
		Requirements: RoleRequirements{
			RequiredInputModalities:  record.RequiredInputModalities,
			RequiredOutputModalities: record.RequiredOutputModalities,
			RequiresStreaming:        record.RequiresStreaming,
			RequiresToolCalling:      record.RequiresToolCalling,
			RequiresStructuredOutput: record.RequiresStructuredOutput,
			RequiresVision:           record.RequiresVision,
		},
		Constraints: RoleConstraints{
			MaxCostTier:        record.MaxCostTier,
			MaxLatencyTier:     record.MaxLatencyTier,
			MinReliabilityTier: record.MinReliabilityTier,
		},
	}, nil
}

// DeleteRole removes a role.
func (s *Service) DeleteRole(ctx context.Context, roleID string) error {

	if s == nil || s.repo == nil {
		return fmt.Errorf("catalog service: repo required")
	}
	return s.repo.DeleteRole(ctx, roleID)
}

// AssignRole assigns a model to a role with validation.
func (s *Service) AssignRole(ctx context.Context, roleID, modelEntryID, assignedBy string) (RoleAssignmentResult, error) {

	if s == nil || s.repo == nil {
		return RoleAssignmentResult{}, fmt.Errorf("catalog service: repo required")
	}

	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return RoleAssignmentResult{}, err
	}
	if role.ID == "" {
		return RoleAssignmentResult{}, fmt.Errorf("catalog service: role not found")
	}

	capabilities, err := s.repo.GetModelCapabilities(ctx, modelEntryID)
	if err != nil {
		return RoleAssignmentResult{}, err
	}
	if capabilities.InputModalities == nil && capabilities.OutputModalities == nil {
		return RoleAssignmentResult{}, fmt.Errorf("catalog service: model capabilities missing")
	}

	missingModalities := diffModalities(role.RequiredInputModalities, capabilities.InputModalities)
	missingModalities = append(missingModalities, diffModalities(role.RequiredOutputModalities, capabilities.OutputModalities)...)

	missingFeatures := []string{}
	if role.RequiresStreaming && !capabilities.SupportsStreaming {
		missingFeatures = append(missingFeatures, "streaming")
	}
	if role.RequiresToolCalling && !capabilities.SupportsToolCalling {
		missingFeatures = append(missingFeatures, "tool_calling")
	}
	if role.RequiresStructuredOutput && !capabilities.SupportsStructuredOutput {
		missingFeatures = append(missingFeatures, "structured_output")
	}
	if role.RequiresVision && !capabilities.SupportsVision {
		missingFeatures = append(missingFeatures, "vision")
	}

	result := RoleAssignmentResult{MissingModalities: missingModalities, MissingFeatures: missingFeatures}
	if len(missingModalities) > 0 || len(missingFeatures) > 0 {
		message := "model does not satisfy role requirements"
		if len(missingModalities) > 0 || len(missingFeatures) > 0 {
			messageParts := []string{}
			if len(missingModalities) > 0 {
				messageParts = append(messageParts, fmt.Sprintf("missing modalities: %s", strings.Join(missingModalities, ", ")))
			}
			if len(missingFeatures) > 0 {
				messageParts = append(messageParts, fmt.Sprintf("missing features: %s", strings.Join(missingFeatures, ", ")))
			}
			message = fmt.Sprintf("%s (%s)", message, strings.Join(messageParts, "; "))
		}
		return result, &CatalogError{Code: ErrorCodeRoleValidation, Message: message}
	}

	if strings.TrimSpace(assignedBy) == "" {
		assignedBy = "user"
	}
	err = s.repo.UpsertRoleAssignment(ctx, catalogrepo.RoleAssignmentRecord{
		RoleID:              roleID,
		ModelCatalogEntryID: modelEntryID,
		AssignedBy:          assignedBy,
		CreatedAt:           time.Now().UnixMilli(),
		Enabled:             true,
	})
	if err != nil {
		return RoleAssignmentResult{}, err
	}
	return result, nil
}

// UnassignRole removes an assignment.
func (s *Service) UnassignRole(ctx context.Context, roleID, modelEntryID string) error {

	if s == nil || s.repo == nil {
		return fmt.Errorf("catalog service: repo required")
	}
	return s.repo.DeleteRoleAssignment(ctx, roleID, modelEntryID)
}

func (s *Service) refreshProvider(ctx context.Context, providerConfig config.ProviderConfig, info providerusecase.Info) error {

	if s.repo == nil {
		return fmt.Errorf("catalog service: repo required")
	}

	record, err := s.repo.EnsureProvider(ctx, catalogrepo.ProviderRecord{
		Name:        providerConfig.Name,
		DisplayName: providerConfig.DisplayName,
		AdapterType: providerConfig.Type,
		TrustMode:   "user_managed",
		BaseURL:     providerConfig.BaseURL,
	})
	if err != nil {
		return err
	}

	if !info.IsConnected {
		return nil
	}

	if err := s.providers.RefreshResources(ctx, providerConfig.Name); err != nil {
		if statusErr := s.repo.UpdateProviderStatus(ctx, record.ID, time.Now().UnixMilli(), false, err.Error(), record.LastDiscoveryAt); statusErr != nil {
			s.logWarn("Catalog provider status update failed", statusErr, coreports.LogField{Key: "provider", Value: record.Name})
		}
		return &CatalogError{Code: ErrorCodeDiscoveryFailure, Message: err.Error(), Cause: err}
	}

	resources := s.providers.GetResources(providerConfig.Name)
	endpoints := s.buildEndpoints(providerConfig, info, resources, record)

	for _, endpoint := range endpoints {
		if err := s.syncEndpointModels(ctx, endpoint, resources); err != nil {
			s.logWarn("Catalog sync endpoint failed", err, coreports.LogField{Key: "endpoint", Value: endpoint.ID})
		}
	}

	if statusErr := s.repo.UpdateProviderStatus(ctx, record.ID, time.Now().UnixMilli(), true, "", time.Now().UnixMilli()); statusErr != nil {
		s.logWarn("Catalog provider status update failed", statusErr, coreports.LogField{Key: "provider", Value: record.Name})
	}
	return nil
}

func (s *Service) buildEndpoints(providerConfig config.ProviderConfig, info providerusecase.Info, resources []providerusecase.Model, provider catalogrepo.ProviderRecord) []catalogrepo.EndpointRecord {

	now := time.Now().UnixMilli()
	authJSON := s.buildEndpointAuth(info)

	endpointInputs, err := s.repo.LoadProviderInputs(providerConfig.Name)
	if err != nil {
		s.logWarn("Catalog provider inputs load failed", err, coreports.LogField{Key: "provider", Value: providerConfig.Name})
	}
	baseURL := providerConfig.BaseURL
	if providerConfig.Type == "cloudflare" {
		baseURL = resolveCloudflareBaseURL(providerConfig.BaseURL, endpointInputs)
	}

	endpoints := []catalogrepo.EndpointRecord{}
	if providerConfig.Type == "cloudflare" {
		workersModels, gatewayModels := splitCloudflareModels(resources)
		if len(workersModels) > 0 {
			endpoints = append(endpoints, catalogrepo.EndpointRecord{
				ProviderID:       provider.ID,
				ProviderName:     provider.Name,
				DisplayName:      "Workers AI",
				AdapterType:      providerConfig.Type,
				BaseURL:          baseURL,
				RouteKind:        "hosted",
				OriginProvider:   "cloudflare_workers_ai",
				OriginRouteLabel: "workers-ai",
				AuthJSON:         authJSON,
				LastTestAt:       now,
				LastTestOK:       info.Status != nil && info.Status.OK,
			})
		}
		if len(gatewayModels) > 0 {
			endpoints = append(endpoints, catalogrepo.EndpointRecord{
				ProviderID:       provider.ID,
				ProviderName:     provider.Name,
				DisplayName:      "Gateway Route",
				AdapterType:      providerConfig.Type,
				BaseURL:          baseURL,
				RouteKind:        "gateway_route",
				OriginProvider:   "other",
				OriginRouteLabel: "gateway",
				AuthJSON:         authJSON,
				LastTestAt:       now,
				LastTestOK:       info.Status != nil && info.Status.OK,
			})
		}
		return s.upsertEndpoints(endpoints)
	}

	routeKind := "direct"
	originProvider := ""
	if providerConfig.Type == "openrouter" {
		routeKind = "gateway_route"
		originProvider = "openrouter"
	}

	endpoints = append(endpoints, catalogrepo.EndpointRecord{
		ProviderID:       provider.ID,
		ProviderName:     provider.Name,
		DisplayName:      "Primary",
		AdapterType:      providerConfig.Type,
		BaseURL:          baseURL,
		RouteKind:        routeKind,
		OriginProvider:   originProvider,
		OriginRouteLabel: "",
		AuthJSON:         authJSON,
		LastTestAt:       now,
		LastTestOK:       info.Status != nil && info.Status.OK,
	})

	return s.upsertEndpoints(endpoints)
}

func (s *Service) upsertEndpoints(endpoints []catalogrepo.EndpointRecord) []catalogrepo.EndpointRecord {

	ctx := context.Background()
	results := make([]catalogrepo.EndpointRecord, 0, len(endpoints))
	for _, endpoint := range endpoints {
		record, err := s.repo.UpsertEndpoint(ctx, endpoint)
		if err != nil {
			s.logWarn("Catalog endpoint upsert failed", err, coreports.LogField{Key: "endpoint", Value: endpoint.DisplayName})
			continue
		}
		results = append(results, record)
	}
	return results
}

func (s *Service) syncEndpointModels(ctx context.Context, endpoint catalogrepo.EndpointRecord, resources []providerusecase.Model) error {

	existing, err := s.repo.ListModelEntriesByEndpoint(ctx, endpoint.ID)
	if err != nil {
		return err
	}

	existingByModel := make(map[string]catalogrepo.ModelEntryRecord, len(existing))
	for _, entry := range existing {
		existingByModel[entry.ModelID] = entry
	}

	models := resources
	if endpoint.RouteKind == "hosted" && endpoint.OriginProvider == "cloudflare_workers_ai" {
		workersModels, _ := splitCloudflareModels(resources)
		models = workersModels
	}
	if endpoint.RouteKind == "gateway_route" && endpoint.AdapterType == "cloudflare" {
		_, gatewayModels := splitCloudflareModels(resources)
		models = gatewayModels
	}

	now := time.Now().UnixMilli()
	seen := make(map[string]struct{})
	for _, model := range models {
		entry := existingByModel[model.ID]
		entry.EndpointID = endpoint.ID
		entry.ModelID = model.ID
		entry.DisplayName = model.Name
		if entry.FirstSeenAt == 0 {
			entry.FirstSeenAt = now
		}
		entry.LastSeenAt = now
		entry.AvailabilityState = "available"
		entry.Approved = true
		entry.MissedRefreshes = 0
		metadata, _ := json.Marshal(model)
		if len(metadata) > 0 {
			entry.MetadataJSON = string(metadata)
		}

		stored, err := s.repo.UpsertModelEntry(ctx, entry)
		if err != nil {
			return err
		}
		if err := s.ensureCapabilities(ctx, stored, model); err != nil {
			return err
		}
		seen[model.ID] = struct{}{}
	}

	for _, entry := range existing {
		if _, ok := seen[entry.ModelID]; ok {
			continue
		}
		missed := entry.MissedRefreshes + 1
		state := entry.AvailabilityState
		if missed >= 2 {
			state = "unknown"
		}
		if err := s.repo.UpdateMissingModelEntry(ctx, entry.ID, missed, state); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ensureCapabilities(ctx context.Context, entry catalogrepo.ModelEntryRecord, model providerusecase.Model) error {

	inputModalities := []string{"text"}
	if model.SupportsVision {
		inputModalities = append(inputModalities, "image")
	}
	outputModalities := []string{"text"}

	if err := s.repo.EnsureModelCapabilities(
		ctx,
		entry.ID,
		model.SupportsStreaming,
		model.SupportsTools,
		false,
		model.SupportsVision,
		inputModalities,
		outputModalities,
		"declared",
		time.Now().UnixMilli(),
	); err != nil {
		return err
	}

	if err := s.repo.EnsureModelSystemProfile(
		ctx,
		entry.ID,
		"unknown",
		"unknown",
		"unknown",
		"summarized",
		time.Now().UnixMilli(),
	); err != nil {
		return err
	}

	return s.repo.EnsureModelUserAddenda(ctx, entry.ID)
}

func (s *Service) buildEndpointAuth(info providerusecase.Info) string {

	if info.CredentialFields == nil {
		return ""
	}

	fields := make([]map[string]interface{}, 0, len(info.CredentialFields))
	for _, field := range info.CredentialFields {
		fields = append(fields, map[string]interface{}{
			"name":     field.Name,
			"required": field.Required,
			"secret":   field.Secret,
		})
	}

	data := map[string]interface{}{
		"credential_fields": fields,
	}
	authJSON, err := catalogrepo.MarshalAuthJSON(data)
	if err != nil {
		return ""
	}
	return authJSON
}

func (s *Service) indexProviders() map[string]providerusecase.Info {

	infos := s.providers.List()
	index := make(map[string]providerusecase.Info, len(infos))
	for _, info := range infos {
		index[info.Name] = info
	}
	return index
}

func (s *Service) findProviderConfig(name string) (config.ProviderConfig, bool) {

	for _, providerConfig := range s.cfg.Providers {
		if providerConfig.Name == name {
			return providerConfig, true
		}
	}
	return config.ProviderConfig{}, false
}

// ensureDefaultRoles inserts app-provided roles when they are missing.
func (s *Service) ensureDefaultRoles(ctx context.Context) error {

	if s == nil || s.repo == nil {
		return fmt.Errorf("catalog service: repo required")
	}

	for _, role := range defaultRoleRecords {
		existing, err := s.repo.GetRoleByName(ctx, role.Name)
		if err != nil {
			return err
		}
		if existing.ID != "" {
			continue
		}
		if _, err := s.repo.UpsertRole(ctx, role); err != nil {
			return err
		}
	}

	return nil
}

func resolveCloudflareBaseURL(baseURL string, inputs map[string]string) string {

	trimmed := strings.TrimSpace(baseURL)
	if trimmed != "" {
		return trimmed
	}
	accountID := strings.TrimSpace(inputs["account_id"])
	gatewayID := strings.TrimSpace(inputs["gateway_id"])
	if accountID == "" || gatewayID == "" {
		return ""
	}
	return fmt.Sprintf("https://gateway.ai.cloudflare.com/v1/%s/%s/compat", accountID, gatewayID)
}

func splitCloudflareModels(models []providerusecase.Model) ([]providerusecase.Model, []providerusecase.Model) {

	var workers []providerusecase.Model
	var gateway []providerusecase.Model
	for _, model := range models {
		if strings.HasPrefix(model.ID, "@cf/") {
			workers = append(workers, model)
		} else {
			gateway = append(gateway, model)
		}
	}
	return workers, gateway
}

func diffModalities(required, supported []string) []string {

	if len(required) == 0 {
		return nil
	}
	supportedSet := make(map[string]struct{}, len(supported))
	for _, modality := range supported {
		supportedSet[modality] = struct{}{}
	}
	var missing []string
	for _, modality := range required {
		if _, ok := supportedSet[modality]; !ok {
			missing = append(missing, modality)
		}
	}
	return missing
}

func mapModelSummary(record catalogrepo.ModelSummaryRecord) ModelSummary {

	contextWindow := parseContextWindow(record.MetadataJSON)
	costTier := strings.TrimSpace(record.CostTier)
	if costTier == "" {
		costTier = "unknown"
	}
	return ModelSummary{
		ID:                       record.ID,
		EndpointID:               record.EndpointID,
		ModelID:                  record.ModelID,
		DisplayName:              record.DisplayName,
		AvailabilityState:        record.AvailabilityState,
		ContextWindow:            contextWindow,
		CostTier:                 costTier,
		SupportsStreaming:        record.SupportsStreaming,
		SupportsToolCalling:      record.SupportsToolCalling,
		SupportsStructuredOutput: record.SupportsStructuredOutput,
		SupportsVision:           record.SupportsVision,
		InputModalities:          record.InputModalities,
		OutputModalities:         record.OutputModalities,
	}
}

// parseContextWindow extracts the context window from metadata JSON.
func parseContextWindow(metadata string) int {

	if strings.TrimSpace(metadata) == "" {
		return 0
	}
	var decoded struct {
		ContextWindow int `json:"contextWindow"`
	}
	if err := json.Unmarshal([]byte(metadata), &decoded); err != nil {
		return 0
	}
	return decoded.ContextWindow
}

func (s *Service) logWarn(message string, err error, fields ...coreports.LogField) {

	if s.logger == nil {
		return
	}
	s.logger.Warn(message, err, fields...)
}
