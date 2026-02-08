// image_facade.go adapts image service capabilities into app contracts.
// internal/app/image_facade.go
package app

import (
	"context"
	"fmt"

	"github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
	imagefeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/ports"
)

// NewImageOperations adapts image service operations into app contracts.
func NewImageOperations(images imagefeature.ImageInterface) ImageOperations {

	return &imageOperations{images: images}
}

// imageOperations exposes image service operations through app contracts.
type imageOperations struct {
	images imagefeature.ImageInterface
}

// GenerateImage produces an image using a configured provider.
func (o *imageOperations) GenerateImage(ctx context.Context, request contracts.GenerateImageRequest) (contracts.ImageBinaryResult, error) {

	if o.images == nil {
		return contracts.ImageBinaryResult{}, fmt.Errorf("app images: service not configured")
	}

	result, err := o.images.GenerateImage(ctx, imagefeature.GenerateImageRequest{
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
		return contracts.ImageBinaryResult{}, err
	}
	return contracts.ImageBinaryResult{
		Bytes:         result.Bytes,
		RevisedPrompt: result.RevisedPrompt,
	}, nil
}

// EditImage edits an image using a configured provider.
func (o *imageOperations) EditImage(ctx context.Context, request contracts.EditImageRequest) (contracts.ImageBinaryResult, error) {

	if o.images == nil {
		return contracts.ImageBinaryResult{}, fmt.Errorf("app images: service not configured")
	}

	result, err := o.images.EditImage(ctx, imagefeature.EditImageRequest{
		ProviderName: request.ProviderName,
		ModelName:    request.ModelName,
		Prompt:       request.Prompt,
		ImagePath:    request.ImagePath,
		MaskPath:     request.MaskPath,
		N:            request.N,
		Size:         request.Size,
	})
	if err != nil {
		return contracts.ImageBinaryResult{}, err
	}
	return contracts.ImageBinaryResult{
		Bytes:         result.Bytes,
		RevisedPrompt: result.RevisedPrompt,
	}, nil
}
