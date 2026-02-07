// logger.go builds zerolog test loggers with contract-friendly defaults.
// pkg/zerologtest/logger.go
package zerologtest

import (
	"io"

	"github.com/rs/zerolog"
)

// NewLogger creates a zerolog logger with timestamp and caller fields.
func NewLogger(writer io.Writer) zerolog.Logger {

	return zerolog.New(writer).Level(zerolog.TraceLevel).With().Timestamp().Caller().Logger()
}
