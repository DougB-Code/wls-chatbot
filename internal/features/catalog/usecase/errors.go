// errors.go defines structured error codes for catalog operations.
// internal/features/catalog/usecase/errors.go
package catalog

import "fmt"

// ErrorCode identifies catalog failure types.
type ErrorCode string

const (
	ErrorCodeProviderAuthFailure ErrorCode = "provider_auth_failure"
	ErrorCodeDiscoveryFailure    ErrorCode = "discovery_failure"
	ErrorCodeRoleValidation      ErrorCode = "role_validation_failure"
)

// CatalogError wraps a failure with a structured error code.
type CatalogError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error formats the catalog error for display.
func (e *CatalogError) Error() string {

	if e == nil {
		return ""
	}
	if e.Message == "" && e.Cause != nil {
		return e.Cause.Error()
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("catalog error: %s", e.Code)
}

// Unwrap returns the underlying cause.
func (e *CatalogError) Unwrap() error {

	if e == nil {
		return nil
	}
	return e.Cause
}
