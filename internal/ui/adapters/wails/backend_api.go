// backend_api.go exposes backend parity endpoints to the frontend bridge.
// internal/ui/adapters/wails/backend_api.go
package wails

import (
	"fmt"

	imageports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/ports"
	modelinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/interfaces"
)

// GenerateImage generates an image using the shared backend interface.
func (b *Bridge) GenerateImage(request imageports.GenerateImageRequest) (imageports.ImageBinaryResult, error) {

	if b.app == nil || b.app.Images == nil {
		return imageports.ImageBinaryResult{}, fmt.Errorf("backend interface not configured")
	}

	return b.app.Images.GenerateImage(b.ctxOrBackground(), request)
}

// EditImage edits an image using the shared backend interface.
func (b *Bridge) EditImage(request imageports.EditImageRequest) (imageports.ImageBinaryResult, error) {

	if b.app == nil || b.app.Images == nil {
		return imageports.ImageBinaryResult{}, fmt.Errorf("backend interface not configured")
	}

	return b.app.Images.EditImage(b.ctxOrBackground(), request)
}

// ListModels lists model catalog entries using the shared backend interface.
func (b *Bridge) ListModels(source string) ([]modelinterfaces.ModelSummary, error) {

	if b.app == nil || b.app.Models == nil {
		return nil, fmt.Errorf("backend interface not configured")
	}

	return b.app.Models.ListModels(b.ctxOrBackground(), modelinterfaces.ModelListFilter{Source: source})
}

// QueryModels lists model catalog entries using advanced capability filters.
func (b *Bridge) QueryModels(filter modelinterfaces.ModelListFilter) ([]modelinterfaces.ModelSummary, error) {

	if b.app == nil || b.app.Models == nil {
		return nil, fmt.Errorf("backend interface not configured")
	}

	return b.app.Models.ListModels(b.ctxOrBackground(), filter)
}

// ImportModels imports custom models from a local file path.
func (b *Bridge) ImportModels(filePath string) error {

	if b.app == nil || b.app.Models == nil {
		return fmt.Errorf("backend interface not configured")
	}
	return b.app.Models.ImportModels(b.ctxOrBackground(), modelinterfaces.ImportModelsRequest{FilePath: filePath})
}

// SyncModels imports custom models from the default app data path.
func (b *Bridge) SyncModels() (modelinterfaces.SyncModelsResult, error) {

	if b.app == nil || b.app.Models == nil {
		return modelinterfaces.SyncModelsResult{}, fmt.Errorf("backend interface not configured")
	}

	return b.app.Models.SyncModels(b.ctxOrBackground())
}
