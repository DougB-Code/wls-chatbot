// model_service.go provides model catalog backend operations.
// internal/core/backend/ai/model_service.go
package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/MadeByDoug/wls-chatbot/internal/core/datastore"
	aiinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/interfaces"
	"github.com/MadeByDoug/wls-chatbot/internal/platform"
)

// ModelCapabilitiesRecord defines catalog capability fields required by model service filters.
type ModelCapabilitiesRecord struct {
	SupportsStreaming        bool
	SupportsToolCalling      bool
	SupportsStructuredOutput bool
	SupportsVision           bool
	InputModalities          []string
	OutputModalities         []string
}

// ModelSummaryRecord defines catalog model summary fields required by model service operations.
type ModelSummaryRecord struct {
	ModelCapabilitiesRecord
	ID                string
	EndpointID        string
	ModelID           string
	DisplayName       string
	Source            string
	Approved          bool
	AvailabilityState string
	MetadataJSON      string
	CostTier          string
}

// EndpointRecord defines catalog endpoint fields required by model service operations.
type EndpointRecord struct {
	ID           string
	ProviderName string
}

// ModelCatalogOperations defines model catalog operations required by the model backend service.
type ModelCatalogOperations interface {
	ListModelSummaries(ctx context.Context) ([]ModelSummaryRecord, error)
	ListModelSystemTags(ctx context.Context) (map[string][]string, error)
	ListEndpoints(ctx context.Context) ([]EndpointRecord, error)
}

// ModelService handles model catalog operations for transport adapters.
type ModelService struct {
	catalog ModelCatalogOperations
	db      *sql.DB
	appName string
}

var _ aiinterfaces.ProviderModelInterface = (*ModelService)(nil)
var _ aiinterfaces.ProviderModelMutationInterface = (*ModelService)(nil)

// NewModelService creates a model backend service from catalog dependencies.
func NewModelService(catalog ModelCatalogOperations, db *sql.DB, appName string) *ModelService {

	return &ModelService{
		catalog: catalog,
		db:      db,
		appName: appName,
	}
}

// ListModels returns model summaries filtered by requested capabilities.
func (s *ModelService) ListModels(ctx context.Context, filter aiinterfaces.ModelListFilter) ([]aiinterfaces.ModelSummary, error) {

	if s.catalog == nil {
		return nil, fmt.Errorf("backend service: model catalog not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	records, err := s.catalog.ListModelSummaries(ctx)
	if err != nil {
		return nil, err
	}
	systemTagsByEntryID, err := s.catalog.ListModelSystemTags(ctx)
	if err != nil {
		return nil, err
	}
	endpoints, err := s.catalog.ListEndpoints(ctx)
	if err != nil {
		return nil, err
	}

	providerByEndpointID := make(map[string]string, len(endpoints))
	for _, endpoint := range endpoints {
		providerByEndpointID[endpoint.ID] = endpoint.ProviderName
	}

	summaries := make([]aiinterfaces.ModelSummary, 0, len(records))
	for _, record := range records {
		if !matchesSourceFilter(record.Source, filter.Source) {
			continue
		}

		profile := buildCapabilityProfile(record, systemTagsByEntryID[record.ID])
		if !matchesModelFilter(profile, filter) {
			continue
		}

		summaries = append(summaries, aiinterfaces.ModelSummary{
			ID:                record.ID,
			ModelID:           record.ModelID,
			DisplayName:       firstNonEmpty(record.DisplayName, record.ModelID),
			ProviderName:      providerByEndpointID[record.EndpointID],
			Source:            record.Source,
			Approved:          record.Approved,
			AvailabilityState: record.AvailabilityState,
			ContextWindow:     parseContextWindowFromMetadata(record.MetadataJSON),
			CostTier:          record.CostTier,
			Capabilities: aiinterfaces.ModelCapabilities{
				SupportsStreaming:        profile.SupportsStreaming,
				SupportsToolCalling:      profile.SupportsToolCalling,
				SupportsStructuredOutput: profile.SupportsStructuredOutput,
				SupportsVision:           profile.SupportsVision,
				InputModalities:          profile.InputModalities,
				OutputModalities:         profile.OutputModalities,
				CapabilityIDs:            profile.CapabilityIDs,
				SystemTags:               profile.SystemTags,
			},
		})
	}

	return summaries, nil
}

// ImportModels imports models from a local YAML file into the catalog datastore.
func (s *ModelService) ImportModels(_ context.Context, request aiinterfaces.ImportModelsRequest) error {

	if s.db == nil {
		return fmt.Errorf("backend service: datastore not configured")
	}

	path := strings.TrimSpace(request.FilePath)
	if path == "" {
		return fmt.Errorf("backend service: import models requires file path")
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("backend service: read model file: %w", err)
	}

	if err := datastore.SeedModels(s.db, payload); err != nil {
		return fmt.Errorf("backend service: import models: %w", err)
	}

	return nil
}

// SyncModels imports models from the default app custom-models file when present.
func (s *ModelService) SyncModels(ctx context.Context) (aiinterfaces.SyncModelsResult, error) {

	if strings.TrimSpace(s.appName) == "" {
		return aiinterfaces.SyncModelsResult{}, fmt.Errorf("backend service: app name not configured")
	}

	appDataDir, err := platform.ResolveAppDataDir(s.appName)
	if err != nil {
		return aiinterfaces.SyncModelsResult{}, fmt.Errorf("backend service: resolve app data dir: %w", err)
	}

	customModelsPath := filepath.Join(appDataDir, "custom-models.yaml")
	if _, err := os.Stat(customModelsPath); err != nil {
		if os.IsNotExist(err) {
			return aiinterfaces.SyncModelsResult{
				Path:     customModelsPath,
				Imported: false,
			}, nil
		}
		return aiinterfaces.SyncModelsResult{}, fmt.Errorf("backend service: stat custom models file: %w", err)
	}

	if err := s.ImportModels(ctx, aiinterfaces.ImportModelsRequest{FilePath: customModelsPath}); err != nil {
		return aiinterfaces.SyncModelsResult{}, err
	}

	return aiinterfaces.SyncModelsResult{
		Path:     customModelsPath,
		Imported: true,
	}, nil
}

// AddProviderModel adds a model to a provider catalog entry set.
func (*ModelService) AddProviderModel(context.Context, aiinterfaces.AddProviderModelRequest) error {

	return fmt.Errorf("backend service: add provider model not configured")
}

// UpdateProviderModel updates mutable model fields for a provider model.
func (*ModelService) UpdateProviderModel(context.Context, aiinterfaces.UpdateProviderModelRequest) error {

	return fmt.Errorf("backend service: update provider model not configured")
}

// RemoveProviderModel removes a model from a provider catalog entry set.
func (*ModelService) RemoveProviderModel(context.Context, string, string) error {

	return fmt.Errorf("backend service: remove provider model not configured")
}

// UpdateProviderModelCapabilities updates mutable capability fields for a provider model.
func (*ModelService) UpdateProviderModelCapabilities(context.Context, aiinterfaces.UpdateProviderModelCapabilitiesRequest) error {

	return fmt.Errorf("backend service: update provider model capabilities not configured")
}

// modelCapabilityProfile contains model capabilities normalized for filtering.
type modelCapabilityProfile struct {
	SupportsStreaming        bool
	SupportsToolCalling      bool
	SupportsStructuredOutput bool
	SupportsVision           bool
	InputModalities          []string
	OutputModalities         []string
	CapabilityIDs            []string
	SystemTags               []string
}

// semanticCapabilityBySystemTag maps provider/system tags to semantic capability identifiers.
var semanticCapabilityBySystemTag = map[string][]string{
	"image_edit":                []string{"vision.edit.image"},
	"vision_segmentation_image": []string{"vision.segmentation.promptable_image"},
	"speech_asr":                []string{"speech.asr"},
	"speech_tts":                []string{"speech.tts"},
}

// buildCapabilityProfile builds a normalized model capability profile used for filtering.
func buildCapabilityProfile(record ModelSummaryRecord, systemTags []string) modelCapabilityProfile {

	normalizedSystemTags := uniqueNormalized(systemTags)
	capabilityIDs := uniqueNormalized(append(
		parseCapabilityIDsFromMetadata(record.MetadataJSON),
		deriveCapabilityIDsFromTags(normalizedSystemTags)...,
	))

	return modelCapabilityProfile{
		SupportsStreaming:        record.SupportsStreaming,
		SupportsToolCalling:      record.SupportsToolCalling,
		SupportsStructuredOutput: record.SupportsStructuredOutput,
		SupportsVision:           record.SupportsVision,
		InputModalities:          uniqueNormalized(record.InputModalities),
		OutputModalities:         uniqueNormalized(record.OutputModalities),
		CapabilityIDs:            capabilityIDs,
		SystemTags:               normalizedSystemTags,
	}
}

// matchesModelFilter reports whether a model capability profile satisfies all filter requirements.
func matchesModelFilter(profile modelCapabilityProfile, filter aiinterfaces.ModelListFilter) bool {

	if len(filter.RequiredInputModalities) > 0 && !containsAll(profile.InputModalities, filter.RequiredInputModalities) {
		return false
	}
	if len(filter.RequiredOutputModalities) > 0 && !containsAll(profile.OutputModalities, filter.RequiredOutputModalities) {
		return false
	}
	if len(filter.RequiredCapabilityIDs) > 0 && !containsAll(profile.CapabilityIDs, filter.RequiredCapabilityIDs) {
		return false
	}
	if len(filter.RequiredSystemTags) > 0 && !containsAll(profile.SystemTags, filter.RequiredSystemTags) {
		return false
	}

	if filter.RequiresStreaming != nil && profile.SupportsStreaming != *filter.RequiresStreaming {
		return false
	}
	if filter.RequiresToolCalling != nil && profile.SupportsToolCalling != *filter.RequiresToolCalling {
		return false
	}
	if filter.RequiresStructuredOutput != nil && profile.SupportsStructuredOutput != *filter.RequiresStructuredOutput {
		return false
	}
	if filter.RequiresVision != nil && profile.SupportsVision != *filter.RequiresVision {
		return false
	}

	return true
}

// deriveCapabilityIDsFromTags maps known model tags into semantic capability identifiers.
func deriveCapabilityIDsFromTags(systemTags []string) []string {

	derived := make([]string, 0)
	for _, tag := range uniqueNormalized(systemTags) {
		derived = append(derived, semanticCapabilityBySystemTag[tag]...)
	}
	return uniqueNormalized(derived)
}

// parseCapabilityIDsFromMetadata extracts semantic capability identifiers from model metadata JSON.
func parseCapabilityIDsFromMetadata(metadata string) []string {

	parsed := parseMetadataObject(metadata)
	if parsed == nil {
		return nil
	}

	capabilityIDs := make([]string, 0)
	for _, key := range []string{"capabilityIds", "capability_ids", "semantic_capabilities"} {
		capabilityIDs = append(capabilityIDs, readStringSlice(parsed[key])...)
	}

	return uniqueNormalized(capabilityIDs)
}

// parseSystemTagsFromMetadata extracts model system tags from model metadata JSON.
func parseSystemTagsFromMetadata(metadata string) []string {

	parsed := parseMetadataObject(metadata)
	if parsed == nil {
		return nil
	}

	tags := make([]string, 0)
	for _, key := range []string{"systemTags", "system_tags", "tags"} {
		tags = append(tags, readStringSlice(parsed[key])...)
	}

	return uniqueNormalized(tags)
}

// containsAll reports whether every value in needles appears in haystack.
func containsAll(haystack []string, needles []string) bool {

	haystackSet := make(map[string]struct{}, len(haystack))
	for _, value := range uniqueNormalized(haystack) {
		haystackSet[value] = struct{}{}
	}

	for _, value := range uniqueNormalized(needles) {
		if _, ok := haystackSet[value]; !ok {
			return false
		}
	}
	return true
}

// matchesSourceFilter reports whether a source value matches the optional source filter.
func matchesSourceFilter(source string, requested string) bool {

	trimmed := strings.TrimSpace(requested)
	if trimmed == "" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(source), trimmed)
}

// parseContextWindowFromMetadata reads optional context window hints from metadata.
func parseContextWindowFromMetadata(metadata string) int {

	parsed := parseMetadataObject(metadata)
	if parsed == nil {
		return 0
	}

	for _, key := range []string{"contextWindow", "context_window", "context_length"} {
		switch value := parsed[key].(type) {
		case float64:
			if value > 0 {
				return int(value)
			}
		case int:
			if value > 0 {
				return value
			}
		}
	}

	return 0
}

// parseMetadataObject parses metadata JSON into a map.
func parseMetadataObject(metadata string) map[string]interface{} {

	if strings.TrimSpace(metadata) == "" {
		return nil
	}

	parsed := make(map[string]interface{})
	if err := json.Unmarshal([]byte(metadata), &parsed); err != nil {
		return nil
	}
	return parsed
}

// readStringSlice converts metadata values into a normalized string slice.
func readStringSlice(value interface{}) []string {

	switch typed := value.(type) {
	case []string:
		return uniqueNormalized(typed)
	case []interface{}:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				items = append(items, text)
			}
		}
		return uniqueNormalized(items)
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return nil
		}
		return []string{strings.ToLower(trimmed)}
	default:
		return nil
	}
}

// uniqueNormalized returns sorted unique lowercase non-empty values.
func uniqueNormalized(values []string) []string {

	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	if len(normalized) == 0 {
		return nil
	}
	slices.Sort(normalized)
	return normalized
}

// firstNonEmpty returns the first non-empty trimmed value.
func firstNonEmpty(values ...string) string {

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
