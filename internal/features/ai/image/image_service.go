// image_service.go provides image generation and editing backend operations.
// internal/core/backend/ai/image_service.go
package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	aiinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/ports"
	provider "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/provider"
)

// ImageProviderOperations defines image operations required by the image backend service.
type ImageProviderOperations interface {
	GenerateImage(ctx context.Context, name string, options provider.ImageGenerationOptions) (*provider.ImageResult, error)
	EditImage(ctx context.Context, name string, options provider.ImageEditOptions) (*provider.ImageResult, error)
}

// ImageService handles image generation and editing operations for transport adapters.
type ImageService struct {
	providers ImageProviderOperations
}

var _ aiinterfaces.ImageInterface = (*ImageService)(nil)

// NewImageService creates an image backend service from provider dependencies.
func NewImageService(providers ImageProviderOperations) *ImageService {

	return &ImageService{providers: providers}
}

// GenerateImage produces an image using a configured provider.
func (s *ImageService) GenerateImage(ctx context.Context, request aiinterfaces.GenerateImageRequest) (aiinterfaces.ImageBinaryResult, error) {

	if s.providers == nil {
		return aiinterfaces.ImageBinaryResult{}, fmt.Errorf("backend service: providers not configured")
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
		return aiinterfaces.ImageBinaryResult{}, err
	}

	return firstImageBinaryResult(ctx, result)
}

// EditImage edits an image using a configured provider.
func (s *ImageService) EditImage(ctx context.Context, request aiinterfaces.EditImageRequest) (aiinterfaces.ImageBinaryResult, error) {

	if s.providers == nil {
		return aiinterfaces.ImageBinaryResult{}, fmt.Errorf("backend service: providers not configured")
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
		return aiinterfaces.ImageBinaryResult{}, err
	}

	return firstImageBinaryResult(ctx, result)
}

// maxCount normalizes optional image count values.
func maxCount(count int) int {

	if count <= 0 {
		return 1
	}
	return count
}

// firstImageBinaryResult extracts and resolves the first image payload.
func firstImageBinaryResult(ctx context.Context, result *provider.ImageResult) (aiinterfaces.ImageBinaryResult, error) {

	if result == nil || len(result.Data) == 0 {
		return aiinterfaces.ImageBinaryResult{}, fmt.Errorf("no image data returned")
	}

	imageData := result.Data[0]
	bytes, err := resolveImageBytes(ctx, imageData)
	if err != nil {
		return aiinterfaces.ImageBinaryResult{}, err
	}

	return aiinterfaces.ImageBinaryResult{
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
