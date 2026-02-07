// provider.go defines the base gateway provider network contract.
// internal/features/providers/interfaces/gateway/provider.go
package gateway

import "context"

// Provider defines baseline network calls for model providers.
type Provider interface {
	Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error)
	GenerateImage(ctx context.Context, opts ImageGenerationOptions) (*ImageResult, error)
	EditImage(ctx context.Context, opts ImageEditOptions) (*ImageResult, error)
	TestConnection(ctx context.Context) error
}
