// model_facade.go adapts model service capabilities into app contracts.
// internal/app/model_facade.go
package app

import (
	"context"
	"fmt"

	"github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
	modelfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/interfaces"
)

// NewModelCatalog adapts model service operations into app contracts.
func NewModelCatalog(models modelfeature.ProviderModelInterface) ModelCatalog {

	return &modelCatalog{models: models}
}

// modelCatalog exposes model service operations through app contracts.
type modelCatalog struct {
	models modelfeature.ProviderModelInterface
}

// ListModels returns model summaries that satisfy the supplied filter.
func (m *modelCatalog) ListModels(ctx context.Context, filter contracts.ModelListFilter) ([]contracts.ModelSummary, error) {

	if m.models == nil {
		return nil, fmt.Errorf("app models: service not configured")
	}

	summaries, err := m.models.ListModels(ctx, modelfeature.ModelListFilter{
		Source:                   filter.Source,
		RequiredInputModalities:  filter.RequiredInputModalities,
		RequiredOutputModalities: filter.RequiredOutputModalities,
		RequiredCapabilityIDs:    filter.RequiredCapabilityIDs,
		RequiredSystemTags:       filter.RequiredSystemTags,
		RequiresStreaming:        filter.RequiresStreaming,
		RequiresToolCalling:      filter.RequiresToolCalling,
		RequiresStructuredOutput: filter.RequiresStructuredOutput,
		RequiresVision:           filter.RequiresVision,
	})
	if err != nil {
		return nil, err
	}

	mapped := make([]contracts.ModelSummary, 0, len(summaries))
	for _, summary := range summaries {
		mapped = append(mapped, contracts.ModelSummary{
			ID:                summary.ID,
			ModelID:           summary.ModelID,
			DisplayName:       summary.DisplayName,
			ProviderName:      summary.ProviderName,
			Source:            summary.Source,
			Approved:          summary.Approved,
			AvailabilityState: summary.AvailabilityState,
			ContextWindow:     summary.ContextWindow,
			CostTier:          summary.CostTier,
			Capabilities: contracts.ModelCapabilities{
				SupportsStreaming:        summary.Capabilities.SupportsStreaming,
				SupportsToolCalling:      summary.Capabilities.SupportsToolCalling,
				SupportsStructuredOutput: summary.Capabilities.SupportsStructuredOutput,
				SupportsVision:           summary.Capabilities.SupportsVision,
				InputModalities:          summary.Capabilities.InputModalities,
				OutputModalities:         summary.Capabilities.OutputModalities,
				CapabilityIDs:            summary.Capabilities.CapabilityIDs,
				SystemTags:               summary.Capabilities.SystemTags,
			},
		})
	}
	return mapped, nil
}

// ImportModels imports model definitions from disk.
func (m *modelCatalog) ImportModels(ctx context.Context, request contracts.ImportModelsRequest) error {

	if m.models == nil {
		return fmt.Errorf("app models: service not configured")
	}

	return m.models.ImportModels(ctx, modelfeature.ImportModelsRequest{FilePath: request.FilePath})
}

// SyncModels syncs model definitions from the default application path.
func (m *modelCatalog) SyncModels(ctx context.Context) (contracts.SyncModelsResult, error) {

	if m.models == nil {
		return contracts.SyncModelsResult{}, fmt.Errorf("app models: service not configured")
	}

	result, err := m.models.SyncModels(ctx)
	if err != nil {
		return contracts.SyncModelsResult{}, err
	}
	return contracts.SyncModelsResult{
		Path:     result.Path,
		Imported: result.Imported,
	}, nil
}
