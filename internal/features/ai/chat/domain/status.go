// status.go normalizes status codes from errors.
// internal/features/chat/domain/status.go
package domain

import "errors"

// StatusCoder is an interface for errors that have a status code.
type StatusCoder interface {
	StatusCode() int
}

// StatusCodeFromErr extracts a status code from an error if available.
func StatusCodeFromErr(err error) int {

	var coder StatusCoder
	if errors.As(err, &coder) {
		return coder.StatusCode()
	}
	return 0
}
