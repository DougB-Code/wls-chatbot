// bootstrap the Wails application and wire dependencies.
// main.go
package main

import (
	"database/sql"
	"embed"
	"path/filepath"
	"strings"

	"github.com/MadeByDoug/wls-chatbot/internal/adapters/chatrepo"
	"github.com/MadeByDoug/wls-chatbot/internal/adapters/configstore"
	"github.com/MadeByDoug/wls-chatbot/internal/adapters/datastore"
	"github.com/MadeByDoug/wls-chatbot/internal/adapters/logger"
	provideradapter "github.com/MadeByDoug/wls-chatbot/internal/adapters/provider"
	"github.com/MadeByDoug/wls-chatbot/internal/adapters/securestore"
	wailsadapter "github.com/MadeByDoug/wls-chatbot/internal/adapters/wails"
	"github.com/MadeByDoug/wls-chatbot/internal/app/config"
	"github.com/MadeByDoug/wls-chatbot/internal/app/wiring"
	chatusecase "github.com/MadeByDoug/wls-chatbot/internal/core/usecase/chat"
	providerusecase "github.com/MadeByDoug/wls-chatbot/internal/core/usecase/provider"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

const AppName = "wls-chatbot"
const KeyringServiceName = "github.com/MadeByDoug/wls-chatbot"

// main is the application entry point.
func main() {

	// 0. Initialize Logger
	log := logger.New("info")
	cmd := newRootCommand(log)
	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Application failure")
	}

}

// newRootCommand builds the CLI entrypoint.
func newRootCommand(log zerolog.Logger) *cobra.Command {

	var dbPath string
	var logLevel string
	cmd := &cobra.Command{
		Use:   AppName,
		Short: "Wails Lit Starter ChatBot",
		RunE: func(_ *cobra.Command, _ []string) error {
			if logLevel != "" {
				log = logger.New(logLevel)
			} else {
				log = logger.New("info")
			}
			log.Info().Msg("Starting Wails Lit Starter ChatBot...")
			databasePath, err := resolveDatabasePath(dbPath)
			if err != nil {
				return err
			}
			db, err := datastore.OpenSQLite(databasePath)
			if err != nil {
				return err
			}
			defer func() {
				_ = db.Close()
			}()
			store, err := configstore.NewSQLiteStore(db)
			if err != nil {
				return err
			}
			cfg, err := config.LoadConfig(store)
			if err != nil {
				return err
			}
			return runUI(log, cfg, db)
		},
	}

	cmd.Flags().StringVar(&dbPath, "db-path", "", "Path to the SQLite database file")
	cmd.Flags().StringVar(&logLevel, "log-level", "debug", "Log level (debug, info, warn, error)")
	return cmd
}

// runUI wires dependencies and launches the app.
func runUI(log zerolog.Logger, cfg config.AppConfig, db *sql.DB) error {

	bridgeService, logBridge, err := setupApp(log, cfg, db)
	if err != nil {
		if bridgeService == nil || logBridge == nil {
			return err
		}
		log.Warn().Err(err).Msg("Failed to initialize app services; continuing with defaults")
	}

	// Create application with options
	if err := wails.Run(&options.App{
		Title:  "Wails Lit Starter ChatBot",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 255},
		OnStartup:        bridgeService.Startup,
		OnShutdown:       bridgeService.Shutdown,
		Bind: []interface{}{
			bridgeService,
			logBridge,
		},
	}); err != nil {
		return err
	}

	return nil
}

// setupApp wires application services and adapters.
func setupApp(log zerolog.Logger, cfg config.AppConfig, db *sql.DB) (*wailsadapter.Bridge, *logger.Logger, error) {

	coreLogger := logger.NewAdapter(log)
	secrets := securestore.NewKeyringStore(KeyringServiceName)
	cache, cacheErr := provideradapter.NewCache(db)
	if cacheErr != nil {
		return nil, nil, cacheErr
	}
	providerService, registry, err := wiring.BuildProviderService(cfg, cache, secrets, coreLogger)

	chatRepo, repoErr := chatrepo.NewRepository(db)
	if repoErr != nil {
		return nil, nil, repoErr
	}
	chatService := chatusecase.NewService(chatRepo)

	emitter := &wailsadapter.Emitter{}
	bridgeService := wailsadapter.New(
		chatusecase.NewOrchestrator(chatService, registry, secrets, emitter),
		providerusecase.NewOrchestrator(providerService, secrets, emitter),
		emitter,
	)
	logBridge := logger.NewLogBridge(log)

	return bridgeService, logBridge, err
}

// resolveDatabasePath resolves the SQLite database path to use.
func resolveDatabasePath(override string) (string, error) {

	if strings.TrimSpace(override) != "" {
		return override, nil
	}

	appDataDir, err := config.ResolveAppDataDir(AppName)
	if err != nil {
		return "", err
	}

	return filepath.Join(appDataDir, "appdata.db"), nil
}
