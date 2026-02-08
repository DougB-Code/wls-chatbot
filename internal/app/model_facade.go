// model_facade.go adapts model service capabilities into app interfaces.
// internal/app/model_facade.go
package app

import (
	"context"
	"fmt"

	modelinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/interfaces"
)

// NewModelCatalog adapts model service operations into app interfaces.
func NewModelCatalog(models modelinterfaces.ProviderModelInterface) ModelCatalog {

	return &modelCatalog{models: models}
}

// modelCatalog exposes model service operations through app interfaces.
type modelCatalog struct {
	models modelinterfaces.ProviderModelInterface
}

// ListModels returns model summaries that satisfy the supplied filter.
func (m *modelCatalog) ListModels(ctx context.Context, filter modelinterfaces.ModelListFilter) ([]modelinterfaces.ModelSummary, error) {

	if m.models == nil {
		return nil, fmt.Errorf("app models: service not configured")
	}

	return m.models.ListModels(ctx, filter)
}

// ImportModels imports model definitions from disk.
func (m *modelCatalog) ImportModels(ctx context.Context, request modelinterfaces.ImportModelsRequest) error {

	if m.models == nil {
		return fmt.Errorf("app models: service not configured")
	}

	return m.models.ImportModels(ctx, request)
}

// SyncModels syncs model definitions from the default application path.
func (m *modelCatalog) SyncModels(ctx context.Context) (modelinterfaces.SyncModelsResult, error) {

	if m.models == nil {
		return modelinterfaces.SyncModelsResult{}, fmt.Errorf("app models: service not configured")
	}

	return m.models.SyncModels(ctx)
}
