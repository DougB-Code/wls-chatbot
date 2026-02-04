// http_client.go defines HTTP client helpers for provider integrations.
// internal/adapters/provider/http_client.go
package provider

import (
	"net/http"
	"time"
)

const defaultHTTPTimeout = 15 * time.Second

// HTTPClient defines the minimal HTTP client contract for providers.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// defaultHTTPClient constructs the default HTTP client with timeouts.
func defaultHTTPClient() *http.Client {

	return &http.Client{
		Timeout: defaultHTTPTimeout,
	}
}
