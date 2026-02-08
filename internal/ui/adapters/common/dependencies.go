// dependencies.go defines shared dependency contracts for UI adapters.
// internal/ui/adapters/common/dependencies.go
package common

import (
	"database/sql"
	"fmt"

	config "github.com/MadeByDoug/wls-chatbot/internal/core/config"
	"github.com/rs/zerolog"
)

// Dependencies groups construction functions shared by all UI adapters.
type Dependencies struct {
	AppName            string
	KeyringServiceName string
	DefaultLogLevel    string
	BaseLogger         zerolog.Logger
	DB                 *sql.DB
	Config             config.AppConfig
}

// ValidateCore validates shared dependency fields required by all adapters.
func (d Dependencies) ValidateCore() error {

	if d.AppName == "" {
		return fmt.Errorf("app name required")
	}
	if d.KeyringServiceName == "" {
		return fmt.Errorf("keyring service name required")
	}
	if d.DefaultLogLevel == "" {
		return fmt.Errorf("default log level required")
	}
	return nil
}

// ValidateResolved validates runtime dependencies resolved before command execution.
func (d Dependencies) ValidateResolved() error {

	if d.DB == nil {
		return fmt.Errorf("database required")
	}
	return nil
}
