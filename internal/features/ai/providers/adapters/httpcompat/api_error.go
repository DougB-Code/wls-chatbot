// api_error.go defines HTTP API error types for HTTP-compatible provider adapters.
// internal/features/providers/adapters/httpcompat/api_error.go
package providerhttp

import "fmt"

// APIError represents an HTTP API error with a status code.
type APIError struct {
	Code    int
	Message string
}

// Error formats the API error for display.
func (e *APIError) Error() string {

	if e == nil {
		return ""
	}
	if e.Message == "" {
		return fmt.Sprintf("API error: %d", e.Code)
	}
	return fmt.Sprintf("API error: %d - %s", e.Code, e.Message)
}

// StatusCode returns the HTTP status code for this error.
func (e *APIError) StatusCode() int {

	if e == nil {
		return 0
	}
	return e.Code
}
