// backend_helpers.go loads the application facade for AI CLI command execution.
// internal/ui/adapters/cli/ai/backend_helpers.go
package ai

import (
	"fmt"

	application "github.com/MadeByDoug/wls-chatbot/internal/app"
)

// loadApp builds the application facade using root-resolved dependencies.
func loadApp(deps Dependencies) (*application.App, error) {

	if deps.Dependencies == nil {
		return nil, fmt.Errorf("cli ai adapter: common dependencies required")
	}

	if err := deps.Dependencies.ValidateResolved(); err != nil {
		return nil, fmt.Errorf("cli ai adapter: %w", err)
	}

	applicationFacade, err := buildApp(
		deps.BaseLogger,
		deps.Config,
		deps.DB,
		deps.AppName,
		deps.KeyringServiceName,
	)
	if err != nil {
		return nil, err
	}

	return applicationFacade, nil
}
