// backend_api.go exposes backend parity endpoints to the frontend bridge.
// internal/core/adapters/wails/backend_api.go
package wails

import (
	"fmt"

	coreinterfaces "github.com/MadeByDoug/wls-chatbot/internal/core/interfaces"
)

// GenerateImage generates an image using the shared backend interface.
func (b *Bridge) GenerateImage(request coreinterfaces.GenerateImageRequest) (coreinterfaces.ImageBinaryResult, error) {

	if b.backend == nil {
		return coreinterfaces.ImageBinaryResult{}, fmt.Errorf("backend interface not configured")
	}
	return b.backend.GenerateImage(b.ctxOrBackground(), request)
}

// EditImage edits an image using the shared backend interface.
func (b *Bridge) EditImage(request coreinterfaces.EditImageRequest) (coreinterfaces.ImageBinaryResult, error) {

	if b.backend == nil {
		return coreinterfaces.ImageBinaryResult{}, fmt.Errorf("backend interface not configured")
	}
	return b.backend.EditImage(b.ctxOrBackground(), request)
}

// ListModels lists model catalog entries using the shared backend interface.
func (b *Bridge) ListModels(source string) ([]coreinterfaces.ModelSummary, error) {

	if b.backend == nil {
		return nil, fmt.Errorf("backend interface not configured")
	}
	return b.backend.ListModels(b.ctxOrBackground(), coreinterfaces.ModelListFilter{
		Source: source,
	})
}

// QueryModels lists model catalog entries using advanced capability filters.
func (b *Bridge) QueryModels(filter coreinterfaces.ModelListFilter) ([]coreinterfaces.ModelSummary, error) {

	if b.backend == nil {
		return nil, fmt.Errorf("backend interface not configured")
	}
	return b.backend.ListModels(b.ctxOrBackground(), filter)
}

// ImportModels imports custom models from a local file path.
func (b *Bridge) ImportModels(filePath string) error {

	if b.backend == nil {
		return fmt.Errorf("backend interface not configured")
	}
	return b.backend.ImportModels(b.ctxOrBackground(), coreinterfaces.ImportModelsRequest{
		FilePath: filePath,
	})
}

// SyncModels imports custom models from the default app data path.
func (b *Bridge) SyncModels() (coreinterfaces.SyncModelsResult, error) {

	if b.backend == nil {
		return coreinterfaces.SyncModelsResult{}, fmt.Errorf("backend interface not configured")
	}
	return b.backend.SyncModels(b.ctxOrBackground())
}
