// image_facade.go adapts image service capabilities into app interfaces.
// internal/app/image_facade.go
package app

import (
	"context"
	"fmt"

	imageports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/ports"
)

// NewImageOperations adapts image service operations into app interfaces.
func NewImageOperations(images imageports.ImageInterface) ImageOperations {

	return &imageOperations{images: images}
}

// imageOperations exposes image service operations through app interfaces.
type imageOperations struct {
	images imageports.ImageInterface
}

// GenerateImage produces an image using a configured provider.
func (o *imageOperations) GenerateImage(ctx context.Context, request imageports.GenerateImageRequest) (imageports.ImageBinaryResult, error) {

	if o.images == nil {
		return imageports.ImageBinaryResult{}, fmt.Errorf("app images: service not configured")
	}

	return o.images.GenerateImage(ctx, request)
}

// EditImage edits an image using a configured provider.
func (o *imageOperations) EditImage(ctx context.Context, request imageports.EditImageRequest) (imageports.ImageBinaryResult, error) {

	if o.images == nil {
		return imageports.ImageBinaryResult{}, fmt.Errorf("app images: service not configured")
	}

	return o.images.EditImage(ctx, request)
}
