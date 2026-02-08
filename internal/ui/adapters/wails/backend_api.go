// backend_api.go exposes backend parity endpoints to the frontend bridge.
// internal/ui/adapters/wails/backend_api.go
package wails

import (
	"fmt"

	appcontracts "github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
	imageinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/ports"
	modelinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/interfaces"
)

// GenerateImage generates an image using the shared backend interface.
func (b *Bridge) GenerateImage(request imageinterfaces.GenerateImageRequest) (imageinterfaces.ImageBinaryResult, error) {

	if b.app == nil || b.app.Images == nil {
		return imageinterfaces.ImageBinaryResult{}, fmt.Errorf("backend interface not configured")
	}

	result, err := b.app.Images.GenerateImage(b.ctxOrBackground(), appcontracts.GenerateImageRequest{
		ProviderName:   request.ProviderName,
		ModelName:      request.ModelName,
		Prompt:         request.Prompt,
		N:              request.N,
		Size:           request.Size,
		Quality:        request.Quality,
		Style:          request.Style,
		ResponseFormat: request.ResponseFormat,
		User:           request.User,
	})
	if err != nil {
		return imageinterfaces.ImageBinaryResult{}, err
	}
	return imageinterfaces.ImageBinaryResult{
		Bytes:         result.Bytes,
		RevisedPrompt: result.RevisedPrompt,
	}, nil
}

// EditImage edits an image using the shared backend interface.
func (b *Bridge) EditImage(request imageinterfaces.EditImageRequest) (imageinterfaces.ImageBinaryResult, error) {

	if b.app == nil || b.app.Images == nil {
		return imageinterfaces.ImageBinaryResult{}, fmt.Errorf("backend interface not configured")
	}

	result, err := b.app.Images.EditImage(b.ctxOrBackground(), appcontracts.EditImageRequest{
		ProviderName: request.ProviderName,
		ModelName:    request.ModelName,
		Prompt:       request.Prompt,
		ImagePath:    request.ImagePath,
		MaskPath:     request.MaskPath,
		N:            request.N,
		Size:         request.Size,
	})
	if err != nil {
		return imageinterfaces.ImageBinaryResult{}, err
	}
	return imageinterfaces.ImageBinaryResult{
		Bytes:         result.Bytes,
		RevisedPrompt: result.RevisedPrompt,
	}, nil
}

// ListModels lists model catalog entries using the shared backend interface.
func (b *Bridge) ListModels(source string) ([]modelinterfaces.ModelSummary, error) {

	if b.app == nil || b.app.Models == nil {
		return nil, fmt.Errorf("backend interface not configured")
	}

	summaries, err := b.app.Models.ListModels(b.ctxOrBackground(), appcontracts.ModelListFilter{Source: source})
	if err != nil {
		return nil, err
	}
	return mapAppModelSummaries(summaries), nil
}

// QueryModels lists model catalog entries using advanced capability filters.
func (b *Bridge) QueryModels(filter modelinterfaces.ModelListFilter) ([]modelinterfaces.ModelSummary, error) {

	if b.app == nil || b.app.Models == nil {
		return nil, fmt.Errorf("backend interface not configured")
	}

	summaries, err := b.app.Models.ListModels(b.ctxOrBackground(), appcontracts.ModelListFilter{
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
	return mapAppModelSummaries(summaries), nil
}

// ImportModels imports custom models from a local file path.
func (b *Bridge) ImportModels(filePath string) error {

	if b.app == nil || b.app.Models == nil {
		return fmt.Errorf("backend interface not configured")
	}
	return b.app.Models.ImportModels(b.ctxOrBackground(), appcontracts.ImportModelsRequest{FilePath: filePath})
}

// SyncModels imports custom models from the default app data path.
func (b *Bridge) SyncModels() (modelinterfaces.SyncModelsResult, error) {

	if b.app == nil || b.app.Models == nil {
		return modelinterfaces.SyncModelsResult{}, fmt.Errorf("backend interface not configured")
	}

	result, err := b.app.Models.SyncModels(b.ctxOrBackground())
	if err != nil {
		return modelinterfaces.SyncModelsResult{}, err
	}
	return modelinterfaces.SyncModelsResult{Path: result.Path, Imported: result.Imported}, nil
}

// mapAppModelSummaries converts app model DTOs into Wails model DTOs.
func mapAppModelSummaries(summaries []appcontracts.ModelSummary) []modelinterfaces.ModelSummary {

	if len(summaries) == 0 {
		return nil
	}

	mapped := make([]modelinterfaces.ModelSummary, 0, len(summaries))
	for _, summary := range summaries {
		mapped = append(mapped, modelinterfaces.ModelSummary{
			ID:                summary.ID,
			ModelID:           summary.ModelID,
			DisplayName:       summary.DisplayName,
			ProviderName:      summary.ProviderName,
			Source:            summary.Source,
			Approved:          summary.Approved,
			AvailabilityState: summary.AvailabilityState,
			ContextWindow:     summary.ContextWindow,
			CostTier:          summary.CostTier,
			Capabilities: modelinterfaces.ModelCapabilities{
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
	return mapped
}
