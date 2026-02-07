// bootstrap the Wails application and wire dependencies.
// main.go
package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corebackend "github.com/MadeByDoug/wls-chatbot/internal/core/backend"
	"github.com/MadeByDoug/wls-chatbot/internal/core/adapters/datastore"
	"github.com/MadeByDoug/wls-chatbot/internal/core/adapters/logger"
	wailsadapter "github.com/MadeByDoug/wls-chatbot/internal/core/adapters/wails"
	coreinterfaces "github.com/MadeByDoug/wls-chatbot/internal/core/interfaces"
	"github.com/MadeByDoug/wls-chatbot/internal/features/catalog/adapters/catalogrepo"
	catalogusecase "github.com/MadeByDoug/wls-chatbot/internal/features/catalog/usecase"
	"github.com/MadeByDoug/wls-chatbot/internal/features/chat/adapters/chatrepo"
	chatusecase "github.com/MadeByDoug/wls-chatbot/internal/features/chat/usecase"
	"github.com/MadeByDoug/wls-chatbot/internal/features/notifications/adapters/notificationrepo"
	notificationusecase "github.com/MadeByDoug/wls-chatbot/internal/features/notifications/usecase"
	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/config"
	providercache "github.com/MadeByDoug/wls-chatbot/internal/features/providers/core/cache"
	"github.com/MadeByDoug/wls-chatbot/internal/features/providers/core/configstore"
	"github.com/MadeByDoug/wls-chatbot/internal/features/providers/core/securestore"
	providerusecase "github.com/MadeByDoug/wls-chatbot/internal/features/settings/usecase"
	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/wiring"

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

	// Add commands
	cmd.AddCommand(newGenerateCommand(log))
	cmd.AddCommand(newModelCommand(log))
	cmd.AddCommand(newProviderCommand(log))

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

}

// newRootCommand builds the CLI entrypoint.
func newRootCommand(log zerolog.Logger) *cobra.Command {

	var dbPath string
	var logLevel string
	cmd := &cobra.Command{
		Use:          AppName,
		Short:        "Wails Lit Starter ChatBot",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
		Logger: logger.NewWailsLogger(log),
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
	cache, cacheErr := providercache.NewSQLiteStore(db)
	if cacheErr != nil {
		return nil, nil, cacheErr
	}
	catalogRepository, catalogErr := catalogrepo.NewRepository(db)
	if catalogErr != nil {
		return nil, nil, catalogErr
	}
	for _, providerConfig := range cfg.Providers {
		if _, err := catalogRepository.EnsureProvider(context.Background(), catalogrepo.ProviderRecord{
			Name:        providerConfig.Name,
			DisplayName: providerConfig.DisplayName,
			AdapterType: providerConfig.Type,
			TrustMode:   "user_managed",
			BaseURL:     providerConfig.BaseURL,
		}); err != nil {
			return nil, nil, err
		}
	}

	providerService, registry, err := wiring.BuildProviderService(cfg, cache, secrets, catalogRepository, coreLogger)

	chatRepo, repoErr := chatrepo.NewRepository(db)
	if repoErr != nil {
		return nil, nil, repoErr
	}
	chatService := chatusecase.NewService(chatRepo)

	notificationRepo, notificationErr := notificationrepo.NewRepository(db)
	if notificationErr != nil {
		return nil, nil, notificationErr
	}
	notificationService := notificationusecase.NewService(notificationRepo)

	emitter := &wailsadapter.Emitter{}
	catalogService := catalogusecase.NewService(catalogRepository, providerService, cfg, coreLogger)
	catalogOrchestrator := catalogusecase.NewOrchestrator(catalogService, emitter)
	providerOrchestrator := providerusecase.NewOrchestrator(providerService, secrets, emitter)
	backendService := corebackend.New(providerOrchestrator, catalogRepository, db, AppName)
	bridgeService := wailsadapter.New(
		chatusecase.NewOrchestrator(chatService, registry, secrets, emitter),
		providerOrchestrator,
		catalogOrchestrator,
		notificationusecase.NewOrchestrator(notificationService),
		emitter,
		backendService,
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

// loadCommandEnvironment opens command dependencies from CLI flags.
func loadCommandEnvironment(dbPath string) (*sql.DB, config.AppConfig, error) {

	databasePath, err := resolveDatabasePath(dbPath)
	if err != nil {
		return nil, config.AppConfig{}, err
	}

	db, err := datastore.OpenSQLite(databasePath)
	if err != nil {
		return nil, config.AppConfig{}, err
	}

	store, err := configstore.NewSQLiteStore(db)
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

// newCommandBackend builds the shared backend service for CLI commands.
func newCommandBackend(log zerolog.Logger, cfg config.AppConfig, db *sql.DB) (coreinterfaces.Backend, error) {

	coreLogger := logger.NewAdapter(log)
	secrets := securestore.NewKeyringStore(KeyringServiceName)

	cache, err := providercache.NewSQLiteStore(db)
	if err != nil {
		return nil, err
	}

	catalogRepository, err := catalogrepo.NewRepository(db)
	if err != nil {
		return nil, err
	}
	for _, providerConfig := range cfg.Providers {
		if _, ensureErr := catalogRepository.EnsureProvider(context.Background(), catalogrepo.ProviderRecord{
			Name:        providerConfig.Name,
			DisplayName: providerConfig.DisplayName,
			AdapterType: providerConfig.Type,
			TrustMode:   "user_managed",
			BaseURL:     providerConfig.BaseURL,
		}); ensureErr != nil {
			return nil, ensureErr
		}
	}

	providerService, _, err := wiring.BuildProviderService(cfg, cache, secrets, catalogRepository, coreLogger)
	if err != nil {
		return nil, err
	}

	providerOrchestrator := providerusecase.NewOrchestrator(providerService, secrets, nil)
	return corebackend.New(providerOrchestrator, catalogRepository, db, AppName), nil
}

// newGenerateCommand creates the parent 'generate' command.
func newGenerateCommand(log zerolog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate content using AI providers",
	}
	cmd.AddCommand(newGenerateImageCommand(log))
	cmd.AddCommand(newGenerateImageEditCommand(log))
	return cmd
}

// newGenerateImageCommand creates the 'generate image' command.
func newGenerateImageCommand(log zerolog.Logger) *cobra.Command {
	var providerName string
	var modelName string
	var prompt string
	var outputPath string
	var dbPath string

	cmd := &cobra.Command{
		Use:   "image",
		Short: "Generate an image",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, cfg, err := loadCommandEnvironment(dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			backendService, err := newCommandBackend(log, cfg, db)
			if err != nil {
				return err
			}

			log.Info().Str("provider", providerName).Str("model", modelName).Msg("Generating image...")
			result, err := backendService.GenerateImage(context.Background(), coreinterfaces.GenerateImageRequest{
				ProviderName: providerName,
				ModelName:    modelName,
				Prompt:       prompt,
				N:            1,
			})
			if err != nil {
				return fmt.Errorf("generation failed: %w", err)
			}

			if outputPath != "" {
				if err := os.WriteFile(outputPath, result.Bytes, 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				log.Info().Str("path", outputPath).Msg("Image saved")
			} else {
				log.Info().Int("bytes", len(result.Bytes)).Msg("Image generated (use --output to save)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name (e.g. gemini, openai)")
	_ = cmd.MarkFlagRequired("provider")
	cmd.Flags().StringVar(&modelName, "model", "", "Model name (optional)")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Image prompt")
	_ = cmd.MarkFlagRequired("prompt")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output path for the generated image")
	cmd.Flags().StringVar(&dbPath, "db-path", "", "Path to the SQLite database file")

	return cmd
}

// newGenerateImageEditCommand creates the 'generate image-edit' command.
func newGenerateImageEditCommand(log zerolog.Logger) *cobra.Command {
	var providerName string
	var modelName string
	var prompt string
	var outputPath string
	var imagePath string
	var maskPath string
	var dbPath string

	cmd := &cobra.Command{
		Use:   "image-edit",
		Short: "Edit an image",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, cfg, err := loadCommandEnvironment(dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			backendService, err := newCommandBackend(log, cfg, db)
			if err != nil {
				return err
			}

			log.Info().Str("provider", providerName).Str("model", modelName).Msg("Editing image...")
			result, err := backendService.EditImage(context.Background(), coreinterfaces.EditImageRequest{
				ProviderName: providerName,
				ModelName:    modelName,
				Prompt:       prompt,
				ImagePath:    imagePath,
				MaskPath:     maskPath,
				N:            1,
			})
			if err != nil {
				return fmt.Errorf("editing failed: %w", err)
			}

			if outputPath != "" {
				if err := os.WriteFile(outputPath, result.Bytes, 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				log.Info().Str("path", outputPath).Msg("Image saved")
			} else {
				log.Info().Int("bytes", len(result.Bytes)).Msg("Image generated (use --output to save)")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&providerName, "provider", "", "Provider name (e.g. gemini, openai)")
	_ = cmd.MarkFlagRequired("provider")
	cmd.Flags().StringVar(&modelName, "model", "", "Model name (optional)")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Image prompt")
	_ = cmd.MarkFlagRequired("prompt")
	cmd.Flags().StringVar(&imagePath, "image", "", "Input image path")
	_ = cmd.MarkFlagRequired("image")
	cmd.Flags().StringVar(&maskPath, "mask", "", "Input mask path (optional)")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output path for the generated image")
	cmd.Flags().StringVar(&dbPath, "db-path", "", "Path to the SQLite database file")

	return cmd
}

// newModelCommand creates the 'model' command with subcommands.
func newModelCommand(log zerolog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "model",
		Aliases: []string{"models"},
		Short:   "Manage model catalog",
	}
	cmd.AddCommand(newModelListCommand(log))
	cmd.AddCommand(newModelImportCommand(log))
	cmd.AddCommand(newModelSyncCommand(log))
	return cmd
}

// newModelListCommand lists models in the catalog.
func newModelListCommand(log zerolog.Logger) *cobra.Command {
	var dbPath string
	var source string
	var requiredInputModalities []string
	var requiredOutputModalities []string
	var requiredCapabilityIDs []string
	var requiredSystemTags []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List models in the catalog",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, cfg, err := loadCommandEnvironment(dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			backendService, err := newCommandBackend(log, cfg, db)
			if err != nil {
				return err
			}
			summaries, err := backendService.ListModels(context.Background(), coreinterfaces.ModelListFilter{
				Source:                   source,
				RequiredInputModalities:  requiredInputModalities,
				RequiredOutputModalities: requiredOutputModalities,
				RequiredCapabilityIDs:    requiredCapabilityIDs,
				RequiredSystemTags:       requiredSystemTags,
			})
			if err != nil {
				return err
			}

			fmt.Printf("%-40s %-15s %-12s %-10s\n", "MODEL ID", "PROVIDER", "SOURCE", "APPROVED")
			fmt.Println(strings.Repeat("-", 80))
			for _, s := range summaries {
				approved := "no"
				if s.Approved {
					approved = "yes"
				}
				fmt.Printf("%-40s %-15s %-12s %-10s\n", s.ModelID, s.ProviderName, s.Source, approved)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dbPath, "db-path", "", "Path to the SQLite database file")
	cmd.Flags().StringVar(&source, "source", "", "Filter by source (seed, user, discovered)")
	cmd.Flags().StringSliceVar(&requiredInputModalities, "requires-input-modality", nil, "Require one or more input modalities (repeat flag)")
	cmd.Flags().StringSliceVar(&requiredOutputModalities, "requires-output-modality", nil, "Require one or more output modalities (repeat flag)")
	cmd.Flags().StringSliceVar(&requiredCapabilityIDs, "requires-capability", nil, "Require one or more semantic capability IDs (repeat flag)")
	cmd.Flags().StringSliceVar(&requiredSystemTags, "requires-system-tag", nil, "Require one or more model system tags (repeat flag)")
	return cmd
}

// newModelImportCommand imports custom models from a YAML file.
func newModelImportCommand(log zerolog.Logger) *cobra.Command {
	var dbPath string
	var filePath string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import custom models from a YAML file",
		Long:  "Import custom models from a YAML file. Format matches models.yaml structure.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, cfg, err := loadCommandEnvironment(dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			backendService, err := newCommandBackend(log, cfg, db)
			if err != nil {
				return err
			}
			if err := backendService.ImportModels(context.Background(), coreinterfaces.ImportModelsRequest{
				FilePath: filePath,
			}); err != nil {
				return err
			}

			fmt.Println("Custom models imported successfully.")
			return nil
		},
	}
	cmd.Flags().StringVar(&dbPath, "db-path", "", "Path to the SQLite database file")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to the custom models YAML file")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

// newModelSyncCommand re-syncs custom models from the default location.
func newModelSyncCommand(log zerolog.Logger) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync custom models from ~/.wls-chatbot/custom-models.yaml",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, cfg, err := loadCommandEnvironment(dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			backendService, err := newCommandBackend(log, cfg, db)
			if err != nil {
				return err
			}
			result, err := backendService.SyncModels(context.Background())
			if err != nil {
				return err
			}
			if !result.Imported {
				fmt.Printf("No custom models file found at %s\n", result.Path)
				return nil
			}

			fmt.Printf("Custom models synced from %s\n", result.Path)
			return nil
		},
	}
	cmd.Flags().StringVar(&dbPath, "db-path", "", "Path to the SQLite database file")
	return cmd
}

// newProviderCommand creates the parent 'provider' command.
func newProviderCommand(log zerolog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage AI providers",
	}
	cmd.AddCommand(newProviderListCommand(log))
	cmd.AddCommand(newProviderTestCommand(log))
	return cmd
}

// newProviderListCommand lists configured providers.
func newProviderListCommand(log zerolog.Logger) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured providers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, cfg, err := loadCommandEnvironment(dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			backendService, err := newCommandBackend(log, cfg, db)
			if err != nil {
				return err
			}
			providers, err := backendService.GetProviders(context.Background())
			if err != nil {
				return err
			}

			fmt.Printf("%-15s %-25s %-10s %-10s\n", "NAME", "DISPLAY NAME", "CONNECTED", "ACTIVE")
			fmt.Println(strings.Repeat("-", 70))
			for _, p := range providers {
				connected := "no"
				active := "no"
				if p.IsConnected {
					connected = "yes"
				}
				if p.IsActive {
					active = "yes"
				}
				fmt.Printf("%-15s %-25s %-10s %-10s\n", p.Name, p.DisplayName, connected, active)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dbPath, "db-path", "", "Path to the SQLite database file")
	return cmd
}

// newProviderTestCommand tests a provider connection.
func newProviderTestCommand(log zerolog.Logger) *cobra.Command {
	var dbPath string
	var providerName string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test a provider connection",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, cfg, err := loadCommandEnvironment(dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			backendService, err := newCommandBackend(log, cfg, db)
			if err != nil {
				return err
			}

			fmt.Printf("Testing connection to %s...\n", providerName)
			if err := backendService.TestProvider(context.Background(), providerName); err != nil {
				fmt.Printf("Connection failed: %v\n", err)
				return err
			}
			fmt.Println("Connection successful.")
			return nil
		},
	}
	cmd.Flags().StringVar(&dbPath, "db-path", "", "Path to the SQLite database file")
	cmd.Flags().StringVar(&providerName, "name", "", "Provider name to test")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
