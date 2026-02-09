// http.go resolves provider image payloads using base64 decoding and HTTP download.
// internal/features/ai/image/adapters/imageresolver/http.go
package imageresolver

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	imageports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/ports"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/ports/gateway"
)

// HTTPClient executes HTTP requests for image payload retrieval.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPResolver resolves image bytes from provider payloads.
type HTTPResolver struct {
	client HTTPClient
}

var _ imageports.ImageBytesResolver = (*HTTPResolver)(nil)

// NewHTTPResolver creates an image payload resolver backed by an HTTP client.
func NewHTTPResolver(client HTTPClient) *HTTPResolver {

	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPResolver{client: client}
}

// Resolve resolves either base64 or URL image payloads into bytes.
func (r *HTTPResolver) Resolve(ctx context.Context, imageData providergateway.ImageData) ([]byte, error) {

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
		response, err := r.client.Do(request)
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
