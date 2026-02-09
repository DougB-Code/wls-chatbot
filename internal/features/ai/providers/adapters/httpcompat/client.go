// client.go defines shared HTTP client helpers for provider adapters.
// internal/features/ai/providers/adapters/httpcompat/client.go
package providerhttp

import (
	"net/http"
	"time"
)

const defaultHTTPTimeout = 15 * time.Second

// Client defines the minimal HTTP client contract for providers.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewDefaultClient constructs the default HTTP client with timeouts.
func NewDefaultClient() *http.Client {

	return &http.Client{
		Timeout: defaultHTTPTimeout,
	}
}
