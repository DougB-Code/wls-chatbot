// ui_runner.go runs the Wails UI flow from app-facade setup.
// internal/ui/adapters/wails/ui_runner.go
package wails

import (
	"database/sql"
	"io/fs"

	config "github.com/MadeByDoug/wls-chatbot/internal/core/config"
	"github.com/rs/zerolog"
)

// runUI launches Wails after resolving bridge dependencies from shared setup.
func runUI(log zerolog.Logger, cfg config.AppConfig, db *sql.DB, assets fs.FS, appName string, keyringServiceName string) error {

	bridgeService, logBridge, err := setupApp(log, cfg, db, appName, keyringServiceName)
	if err != nil {
		return err
	}

	return Run(log, assets, bridgeService, logBridge)
}
