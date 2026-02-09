// image_bytes_resolver.go resolves provider image payloads into byte slices.
// internal/features/ai/image/ports/image_bytes_resolver.go
package ports

import (
	"context"

	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/ports/gateway"
)

// ImageBytesResolver resolves provider image payloads into raw bytes.
type ImageBytesResolver interface {
	Resolve(ctx context.Context, imageData providergateway.ImageData) ([]byte, error)
}
