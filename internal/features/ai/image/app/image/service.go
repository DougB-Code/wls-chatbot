// service.go provides image generation and editing backend operations.
// internal/features/ai/image/app/image/service.go
package image

import (
	"context"
	"fmt"

	imageports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/ports"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/ports/gateway"
)

// ImageProviderOperations defines image operations required by the image backend service.
type ImageProviderOperations interface {
	GenerateImage(ctx context.Context, name string, options providergateway.ImageGenerationOptions) (*providergateway.ImageResult, error)
	EditImage(ctx context.Context, name string, options providergateway.ImageEditOptions) (*providergateway.ImageResult, error)
}

// Service handles image generation and editing operations for transport adapters.
type Service struct {
	providers     ImageProviderOperations
	imageResolver imageports.ImageBytesResolver
}

var _ imageports.ImageInterface = (*Service)(nil)

// NewService creates an image backend service from provider dependencies.
func NewService(providers ImageProviderOperations, imageResolver imageports.ImageBytesResolver) *Service {

	return &Service{
		providers:     providers,
		imageResolver: imageResolver,
	}
}

// GenerateImage produces an image using a configured provider.
func (s *Service) GenerateImage(ctx context.Context, request imageports.GenerateImageRequest) (imageports.ImageBinaryResult, error) {

	if s.providers == nil {
		return imageports.ImageBinaryResult{}, fmt.Errorf("backend service: providers not configured")
	}
	if s.imageResolver == nil {
		return imageports.ImageBinaryResult{}, fmt.Errorf("backend service: image bytes resolver not configured")
	}

	result, err := s.providers.GenerateImage(ctx, request.ProviderName, providergateway.ImageGenerationOptions{
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
		return imageports.ImageBinaryResult{}, err
	}

	return firstImageBinaryResult(ctx, s.imageResolver, result)
}

// EditImage edits an image using a configured provider.
func (s *Service) EditImage(ctx context.Context, request imageports.EditImageRequest) (imageports.ImageBinaryResult, error) {

	if s.providers == nil {
		return imageports.ImageBinaryResult{}, fmt.Errorf("backend service: providers not configured")
	}
	if s.imageResolver == nil {
		return imageports.ImageBinaryResult{}, fmt.Errorf("backend service: image bytes resolver not configured")
	}

	result, err := s.providers.EditImage(ctx, request.ProviderName, providergateway.ImageEditOptions{
		Model:  request.ModelName,
		Image:  request.ImagePath,
		Mask:   request.MaskPath,
		Prompt: request.Prompt,
		N:      maxCount(request.N),
		Size:   request.Size,
	})
	if err != nil {
		return imageports.ImageBinaryResult{}, err
	}

	return firstImageBinaryResult(ctx, s.imageResolver, result)
}

// maxCount normalizes optional image count values.
func maxCount(count int) int {

	if count <= 0 {
		return 1
	}
	return count
}

// firstImageBinaryResult extracts and resolves the first image payload.
func firstImageBinaryResult(ctx context.Context, resolver imageports.ImageBytesResolver, result *providergateway.ImageResult) (imageports.ImageBinaryResult, error) {

	if result == nil || len(result.Data) == 0 {
		return imageports.ImageBinaryResult{}, fmt.Errorf("no image data returned")
	}

	imageData := result.Data[0]
	bytes, err := resolver.Resolve(ctx, imageData)
	if err != nil {
		return imageports.ImageBinaryResult{}, err
	}

	return imageports.ImageBinaryResult{
		Bytes:         bytes,
		RevisedPrompt: imageData.RevisedPrompt,
	}, nil
}
