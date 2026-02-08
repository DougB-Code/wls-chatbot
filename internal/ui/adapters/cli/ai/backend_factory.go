// backend_factory.go builds application facade dependencies for AI CLI execution.
// internal/ui/adapters/cli/ai/backend_factory.go
package ai

import (
	"database/sql"

	application "github.com/MadeByDoug/wls-chatbot/internal/app"
	appwire "github.com/MadeByDoug/wls-chatbot/internal/app/wire"
	config "github.com/MadeByDoug/wls-chatbot/internal/core/config"
	"github.com/rs/zerolog"
)

// buildApp constructs the shared application facade used by AI CLI commands.
func buildApp(log zerolog.Logger, cfg config.AppConfig, db *sql.DB, appName string, keyringServiceName string) (*application.App, error) {

	return appwire.NewApp(appwire.Dependencies{
		Log:                log,
		Config:             cfg,
		DB:                 db,
		AppName:            appName,
		KeyringServiceName: keyringServiceName,
	})
}
