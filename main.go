// main.go composes shared dependencies and delegates to UI adapters.
// main.go
package main

import (
	"database/sql"
	"embed"
	"os"
	"path/filepath"
	"strings"

	config "github.com/MadeByDoug/wls-chatbot/internal/core/config"
	"github.com/MadeByDoug/wls-chatbot/internal/core/datastore"
	"github.com/MadeByDoug/wls-chatbot/internal/core/logger"
	"github.com/MadeByDoug/wls-chatbot/internal/platform"
	cliadapter "github.com/MadeByDoug/wls-chatbot/internal/ui/adapters/cli"
	commonadapter "github.com/MadeByDoug/wls-chatbot/internal/ui/adapters/common"
	wailsadapter "github.com/MadeByDoug/wls-chatbot/internal/ui/adapters/wails"
	"github.com/spf13/cobra"
)

//go:embed all:frontend/dist
var assets embed.FS

const AppName = "wls-chatbot"
const KeyringServiceName = "github.com/MadeByDoug/wls-chatbot"
const DefaultLogLevel = "info"

// main initializes shared dependencies and explicitly mounts UI adapters.
func main() {

	commonDependencies := &commonadapter.Dependencies{
		AppName:            AppName,
		KeyringServiceName: KeyringServiceName,
		DefaultLogLevel:    DefaultLogLevel,
	}
	var dbPath string
	var logLevel string

	root := &cobra.Command{
		Use:          AppName,
		Short:        "Wails Lit Starter ChatBot",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Name() == AppName || cmd.Name() == "help" {
				return nil
			}
			resolvedLevel := strings.TrimSpace(logLevel)
			if resolvedLevel == "" {
				resolvedLevel = commonDependencies.DefaultLogLevel
			}
			commonDependencies.BaseLogger = logger.New(resolvedLevel)
			db, cfg, err := loadCommandEnvironment(dbPath)
			if err != nil {
				return err
			}
			commonDependencies.DB = db
			commonDependencies.Config = cfg
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.PersistentFlags().StringVar(&dbPath, "db-path", "", "Path to the SQLite database file")
	root.PersistentFlags().StringVar(&logLevel, "log-level", commonDependencies.DefaultLogLevel, "Log level (debug, info, warn, error)")

	root.AddCommand(cliadapter.NewCommand(cliadapter.Dependencies{
		Dependencies: commonDependencies,
	}))

	root.AddCommand(wailsadapter.NewCommand(wailsadapter.Dependencies{
		Dependencies: commonDependencies,
		Assets:       assets,
	}))

	err := root.Execute()
	if commonDependencies.DB != nil {
		_ = commonDependencies.DB.Close()
	}
	if err != nil {
		os.Exit(1)
	}
}

// loadCommandEnvironment opens shared command dependencies from CLI flags.
func loadCommandEnvironment(dbPath string) (*sql.DB, config.AppConfig, error) {

	databasePath, err := resolveDatabasePath(dbPath)
	if err != nil {
		return nil, config.AppConfig{}, err
	}

	db, err := datastore.OpenSQLite(databasePath)
	if err != nil {
		return nil, config.AppConfig{}, err
	}

	store, err := config.NewSQLiteStore(db)
	if err != nil {
		_ = db.Close()
		return nil, config.AppConfig{}, err
	}

	cfg, err := config.LoadConfig(store)
	if err != nil {
		_ = db.Close()
		return nil, config.AppConfig{}, err
	}

	return db, cfg, nil
}

// resolveDatabasePath resolves the SQLite database path to use.
func resolveDatabasePath(override string) (string, error) {

	if strings.TrimSpace(override) != "" {
		return override, nil
	}

	appDataDir, err := platform.ResolveAppDataDir(AppName)
	if err != nil {
		return "", err
	}

	return filepath.Join(appDataDir, "appdata.db"), nil
}
