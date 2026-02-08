// app_setup.go builds Wails bridge dependencies from the shared app composition root.
// internal/ui/adapters/wails/app_setup.go
package wails

import (
	"database/sql"

	appwire "github.com/MadeByDoug/wls-chatbot/internal/app/wire"
	config "github.com/MadeByDoug/wls-chatbot/internal/core/config"
	wailslogger "github.com/MadeByDoug/wls-chatbot/internal/ui/adapters/wails/logger"
	"github.com/rs/zerolog"
)

// setupApp wires shared app services and Wails bridge adapters.
func setupApp(log zerolog.Logger, cfg config.AppConfig, db *sql.DB, appName string, keyringServiceName string) (*Bridge, *wailslogger.Logger, error) {

	emitter := &Emitter{}
	applicationFacade, err := appwire.NewApp(appwire.Dependencies{
		Log:                log,
		Config:             cfg,
		DB:                 db,
		AppName:            appName,
		KeyringServiceName: keyringServiceName,
		Events:             emitter,
	})
	if err != nil {
		return nil, nil, err
	}

	bridgeService := New(applicationFacade, emitter)
	logBridge := wailslogger.NewLogBridge(log)
	return bridgeService, logBridge, nil
}
