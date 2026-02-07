// service.go implements the shared backend interface for CLI and Wails adapters.
// internal/core/backend/service.go
package backend

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MadeByDoug/wls-chatbot/internal/core/adapters/datastore"
	coreinterfaces "github.com/MadeByDoug/wls-chatbot/internal/core/interfaces"
	"github.com/MadeByDoug/wls-chatbot/internal/features/catalog/adapters/catalogrepo"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/providers/interfaces/gateway"
	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/config"
	provider "github.com/MadeByDoug/wls-chatbot/internal/features/settings/usecase"
)

// providerOperations defines provider actions required by the backend service.
type providerOperations interface {
	GetProviders() []provider.Info
	TestProvider(ctx context.Context, name string) error
	GenerateImage(ctx context.Context, name string, options provider.ImageGenerationOptions) (*provider.ImageResult, error)
	EditImage(ctx context.Context, name string, options provider.ImageEditOptions) (*provider.ImageResult, error)
}

// catalogOperations defines catalog actions required by the backend service.
type catalogOperations interface {
	ListModelSummaries(ctx context.Context) ([]catalogrepo.ModelSummaryRecord, error)
	ListEndpoints(ctx context.Context) ([]catalogrepo.EndpointRecord, error)
	ListModelSystemTags(ctx context.Context) (map[string][]string, error)
}

// Service provides a single backend capability surface for transport adapters.
type Service struct {
	providers providerOperations
	catalog   catalogOperations
	db        *sql.DB
	appName   string
}

// New creates a backend service from feature-level dependencies.
func New(providers providerOperations, catalog catalogOperations, db *sql.DB, appName string) *Service {

	return &Service{
		providers: providers,
		catalog:   catalog,
		db:        db,
		appName:   appName,
	}
}

var _ coreinterfaces.Backend = (*Service)(nil)

// GetProviders returns configured provider statuses.
func (s *Service) GetProviders(context.Context) ([]provider.Info, error) {

	if s.providers == nil {
		return nil, fmt.Errorf("backend service: providers not configured")
	}
	return s.providers.GetProviders(), nil
}

// TestProvider checks connectivity for a provider.
func (s *Service) TestProvider(ctx context.Context, name string) error {

	if s.providers == nil {
		return fmt.Errorf("backend service: providers not configured")
	}
	return s.providers.TestProvider(ctx, name)
}

// GenerateImage produces an image using a configured provider.
func (s *Service) GenerateImage(ctx context.Context, request coreinterfaces.GenerateImageRequest) (coreinterfaces.ImageBinaryResult, error) {

	if s.providers == nil {
		return coreinterfaces.ImageBinaryResult{}, fmt.Errorf("backend service: providers not configured")
	}

	result, err := s.providers.GenerateImage(ctx, request.ProviderName, provider.ImageGenerationOptions{
		Model:          request.ModelName,
		Prompt:         request.Prompt,
		N:              maxCount(request.N),
		Size:           request.Size,
		Quality:        request.Quality,
		Style:          request.Style,
		ResponseFormat: request.ResponseFormat,
		User:           request.User,
	})
	if err != nil {
		return coreinterfaces.ImageBinaryResult{}, err
	}

	return firstImageBinaryResult(ctx, result)
}

// EditImage edits an image using a configured provider.
func (s *Service) EditImage(ctx context.Context, request coreinterfaces.EditImageRequest) (coreinterfaces.ImageBinaryResult, error) {

	if s.providers == nil {
		return coreinterfaces.ImageBinaryResult{}, fmt.Errorf("backend service: providers not configured")
	}

	result, err := s.providers.EditImage(ctx, request.ProviderName, provider.ImageEditOptions{
		Model:  request.ModelName,
		Image:  request.ImagePath,
		Mask:   request.MaskPath,
		Prompt: request.Prompt,
		N:      maxCount(request.N),
		Size:   request.Size,
	})
	if err != nil {
		return coreinterfaces.ImageBinaryResult{}, err
	}

	return firstImageBinaryResult(ctx, result)
}

// ListModels returns model summaries with capability and modality metadata.
func (s *Service) ListModels(ctx context.Context, filter coreinterfaces.ModelListFilter) ([]coreinterfaces.ModelSummary, error) {

	if s.catalog == nil {
		return nil, fmt.Errorf("backend service: catalog not configured")
	}

	summaries, err := s.catalog.ListModelSummaries(ctx)
	if err != nil {
		return nil, err
	}
	endpoints, err := s.catalog.ListEndpoints(ctx)
	if err != nil {
		return nil, err
	}
	systemTagsByEntryID, err := s.catalog.ListModelSystemTags(ctx)
	if err != nil {
		return nil, err
	}

	providerByEndpoint := make(map[string]string, len(endpoints))
	for _, endpoint := range endpoints {
		providerByEndpoint[endpoint.ID] = endpoint.ProviderName
	}

	sourceFilter := strings.TrimSpace(filter.Source)
	models := make([]coreinterfaces.ModelSummary, 0, len(summaries))
	for _, summary := range summaries {
		if sourceFilter != "" && !strings.EqualFold(summary.Source, sourceFilter) {
			continue
		}
		profile := buildCapabilityProfile(summary, systemTagsByEntryID[summary.ID])
		if !matchesModelFilter(profile, filter) {
			continue
		}
		models = append(models, coreinterfaces.ModelSummary{
			ID:                summary.ID,
			ModelID:           summary.ModelID,
			DisplayName:       summary.DisplayName,
			ProviderName:      providerByEndpoint[summary.EndpointID],
			Source:            summary.Source,
			Approved:          summary.Approved,
			AvailabilityState: summary.AvailabilityState,
			ContextWindow:     parseContextWindow(summary.MetadataJSON),
			CostTier:          normalizeCostTier(summary.CostTier),
			Capabilities:      profile,
		})
	}

	return models, nil
}

// buildCapabilityProfile builds semantic model capabilities from catalog metadata.
func buildCapabilityProfile(summary catalogrepo.ModelSummaryRecord, systemTags []string) coreinterfaces.ModelCapabilities {

	allSystemTags := append([]string{}, systemTags...)
	allSystemTags = append(allSystemTags, parseSystemTagsFromMetadata(summary.MetadataJSON)...)

	capabilityIDs := make([]string, 0)
	for _, id := range deriveBaseCapabilityIDs(summary) {
		capabilityIDs = append(capabilityIDs, id)
	}
	for _, id := range deriveCapabilityIDsFromTags(allSystemTags) {
		capabilityIDs = append(capabilityIDs, id)
	}
	for _, id := range parseCapabilityIDsFromMetadata(summary.MetadataJSON) {
		capabilityIDs = append(capabilityIDs, id)
	}

	return coreinterfaces.ModelCapabilities{
		SupportsStreaming:        summary.SupportsStreaming,
		SupportsToolCalling:      summary.SupportsToolCalling,
		SupportsStructuredOutput: summary.SupportsStructuredOutput,
		SupportsVision:           summary.SupportsVision,
		InputModalities:          normalizeValues(summary.InputModalities),
		OutputModalities:         normalizeValues(summary.OutputModalities),
		CapabilityIDs:            normalizeValues(capabilityIDs),
		SystemTags:               normalizeValues(allSystemTags),
	}
}

// matchesModelFilter returns true when a model satisfies capability and modality requirements.
func matchesModelFilter(profile coreinterfaces.ModelCapabilities, filter coreinterfaces.ModelListFilter) bool {

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

	if !containsAll(profile.InputModalities, filter.RequiredInputModalities) {
		return false
	}
	if !containsAll(profile.OutputModalities, filter.RequiredOutputModalities) {
		return false
	}
	if !containsAll(profile.CapabilityIDs, filter.RequiredCapabilityIDs) {
		return false
	}
	if !containsAll(profile.SystemTags, filter.RequiredSystemTags) {
		return false
	}

	return true
}

// deriveBaseCapabilityIDs infers semantic capabilities from modality/feature combinations.
func deriveBaseCapabilityIDs(summary catalogrepo.ModelSummaryRecord) []string {

	ids := make([]string, 0, 4)
	inputs := normalizeValues(summary.InputModalities)
	outputs := normalizeValues(summary.OutputModalities)

	if containsValue(outputs, "image") {
		ids = append(ids, string(providergateway.CapabilityGenerateImage))
	}
	if containsValue(outputs, "text") && containsAny(inputs, []string{"image", "video", "audio"}) {
		ids = append(ids, string(providergateway.CapabilityChatMultimodalToText))
	} else if containsValue(outputs, "text") && containsValue(inputs, "text") {
		ids = append(ids, string(providergateway.CapabilityChatText))
	}
	if summary.SupportsToolCalling {
		ids = append(ids, string(providergateway.CapabilityAgentToolUse))
	}

	return ids
}

// deriveCapabilityIDsFromTags maps model system tags to semantic capability ids.
func deriveCapabilityIDsFromTags(systemTags []string) []string {

	tagToCapabilityID := map[string]string{
		"image_gen":                 string(providergateway.CapabilityGenerateImage),
		"image_edit":                "vision.edit.image",
		"image_editing":             "vision.edit.image",
		"vision_segmentation_image": string(providergateway.CapabilityVisionSegmentationImage),
		"vision_segmentation_video": string(providergateway.CapabilityVisionSegmentationVideo),
		"image_segmentation":        string(providergateway.CapabilityVisionSegmentationImage),
		"vision_segmentation":       string(providergateway.CapabilityVisionSegmentationImage),
		"audio_transcription":       string(providergateway.CapabilitySpeechASR),
		"speech_asr":                string(providergateway.CapabilitySpeechASR),
		"speech_tts":                string(providergateway.CapabilitySpeechTTS),
		"moderation":                string(providergateway.CapabilitySafetyModeration),
		"embedding":                 string(providergateway.CapabilityRetrievalEmbedText),
		"multimodal_embedding":      string(providergateway.CapabilityRetrievalEmbedMultimodal),
		"rerank":                    string(providergateway.CapabilityRankRerank),
	}

	ids := make([]string, 0, len(systemTags))
	for _, tag := range normalizeValues(systemTags) {
		if capabilityID, ok := tagToCapabilityID[tag]; ok {
			ids = append(ids, capabilityID)
		}
	}
	return ids
}

// parseCapabilityIDsFromMetadata reads semantic capability ids from known metadata keys.
func parseCapabilityIDsFromMetadata(metadata string) []string {

	if strings.TrimSpace(metadata) == "" {
		return nil
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(metadata), &decoded); err != nil {
		return nil
	}

	candidateKeys := []string{
		"capabilityIds",
		"capability_ids",
		"semanticCapabilities",
		"semantic_capabilities",
	}

	ids := make([]string, 0)
	for _, key := range candidateKeys {
		parsed, ok := parseStringSlice(decoded[key])
		if !ok {
			continue
		}
		ids = append(ids, parsed...)
	}
	return ids
}

// parseSystemTagsFromMetadata reads model system tags from known metadata keys.
func parseSystemTagsFromMetadata(metadata string) []string {

	if strings.TrimSpace(metadata) == "" {
		return nil
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(metadata), &decoded); err != nil {
		return nil
	}

	candidateKeys := []string{
		"systemTags",
		"system_tags",
		"tags",
	}

	tags := make([]string, 0)
	for _, key := range candidateKeys {
		parsed, ok := parseStringSlice(decoded[key])
		if !ok {
			continue
		}
		tags = append(tags, parsed...)
	}
	return tags
}

// parseContextWindow extracts model context window from metadata.
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

// normalizeCostTier normalizes empty cost tier values.
func normalizeCostTier(costTier string) string {

	trimmed := strings.TrimSpace(costTier)
	if trimmed == "" {
		return "unknown"
	}
	return trimmed
}

// normalizeValues trims, lowercases, de-duplicates, and sorts values.
func normalizeValues(values []string) []string {

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
	sort.Strings(normalized)
	return normalized
}

// containsAll checks whether every required value exists in the available set.
func containsAll(available []string, required []string) bool {

	requiredNormalized := normalizeValues(required)
	if len(requiredNormalized) == 0 {
		return true
	}

	availableSet := make(map[string]struct{}, len(available))
	for _, value := range normalizeValues(available) {
		availableSet[value] = struct{}{}
	}
	for _, value := range requiredNormalized {
		if _, ok := availableSet[value]; !ok {
			return false
		}
	}
	return true
}

// containsAny checks whether any candidate value exists in available values.
func containsAny(values []string, candidates []string) bool {

	normalizedValues := normalizeValues(values)
	for _, candidate := range normalizeValues(candidates) {
		if containsValue(normalizedValues, candidate) {
			return true
		}
	}
	return false
}

// containsValue checks whether a normalized value exists in available values.
func containsValue(values []string, value string) bool {

	needle := strings.ToLower(strings.TrimSpace(value))
	for _, candidate := range values {
		if strings.EqualFold(candidate, needle) {
			return true
		}
	}
	return false
}

// parseStringSlice converts interface{} arrays into trimmed string slices.
func parseStringSlice(raw interface{}) ([]string, bool) {

	array, ok := raw.([]interface{})
	if !ok {
		return nil, false
	}

	values := make([]string, 0, len(array))
	for _, item := range array {
		value, ok := item.(string)
		if !ok {
			continue
		}
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}
	if len(values) == 0 {
		return nil, false
	}
	return values, true
}

// ImportModels imports custom models from a file path.
func (s *Service) ImportModels(_ context.Context, request coreinterfaces.ImportModelsRequest) error {

	if s.db == nil {
		return fmt.Errorf("backend service: database not configured")
	}
	filePath := strings.TrimSpace(request.FilePath)
	if filePath == "" {
		return fmt.Errorf("custom models file path required")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read custom models file: %w", err)
	}
	if err := datastore.SeedModels(s.db, data); err != nil {
		return fmt.Errorf("import custom models: %w", err)
	}
	return nil
}

// SyncModels imports custom models from the default app data location.
func (s *Service) SyncModels(context.Context) (coreinterfaces.SyncModelsResult, error) {

	if s.db == nil {
		return coreinterfaces.SyncModelsResult{}, fmt.Errorf("backend service: database not configured")
	}

	appDataDir, err := config.ResolveAppDataDir(s.appName)
	if err != nil {
		return coreinterfaces.SyncModelsResult{}, err
	}
	path := filepath.Join(appDataDir, "custom-models.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return coreinterfaces.SyncModelsResult{
				Path:     path,
				Imported: false,
			}, nil
		}
		return coreinterfaces.SyncModelsResult{}, fmt.Errorf("read custom models file: %w", err)
	}
	if err := datastore.SeedModels(s.db, data); err != nil {
		return coreinterfaces.SyncModelsResult{}, fmt.Errorf("sync custom models: %w", err)
	}

	return coreinterfaces.SyncModelsResult{
		Path:     path,
		Imported: true,
	}, nil
}

// maxCount normalizes optional image count values.
func maxCount(count int) int {

	if count <= 0 {
		return 1
	}
	return count
}

// firstImageBinaryResult extracts and resolves the first image payload.
func firstImageBinaryResult(ctx context.Context, result *provider.ImageResult) (coreinterfaces.ImageBinaryResult, error) {

	if result == nil || len(result.Data) == 0 {
		return coreinterfaces.ImageBinaryResult{}, fmt.Errorf("no image data returned")
	}

	imageData := result.Data[0]
	bytes, err := resolveImageBytes(ctx, imageData)
	if err != nil {
		return coreinterfaces.ImageBinaryResult{}, err
	}

	return coreinterfaces.ImageBinaryResult{
		Bytes:         bytes,
		RevisedPrompt: imageData.RevisedPrompt,
	}, nil
}

// resolveImageBytes resolves either base64 or URL image payloads into bytes.
func resolveImageBytes(ctx context.Context, imageData provider.ImageData) ([]byte, error) {

	if ctx == nil {
		ctx = context.Background()
	}

	if strings.TrimSpace(imageData.B64JSON) != "" {
		bytes, err := base64.StdEncoding.DecodeString(imageData.B64JSON)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 image: %w", err)
		}
		return bytes, nil
	}

	if strings.TrimSpace(imageData.URL) != "" {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, imageData.URL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create image request: %w", err)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, fmt.Errorf("failed to download image from URL: %w", err)
		}
		defer func() { _ = response.Body.Close() }()

		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to download image, status: %d", response.StatusCode)
		}

		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read image body: %w", err)
		}
		return bytes, nil
	}

	return nil, fmt.Errorf("provider returned neither base64 nor URL for image")
}
