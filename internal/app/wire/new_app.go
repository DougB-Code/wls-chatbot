// new_app.go composes all feature dependencies into one application facade.
// internal/app/wire/new_app.go
package wire

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/MadeByDoug/wls-chatbot/internal/app"
	config "github.com/MadeByDoug/wls-chatbot/internal/core/config"
	coreevents "github.com/MadeByDoug/wls-chatbot/internal/core/events"
	corelogger "github.com/MadeByDoug/wls-chatbot/internal/core/logger"
	"github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/adapters/chatrepo"
	chatfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/app/chat"
	imageresolver "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/adapters/imageresolver"
	imagefeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/app/image"
	modelio "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/adapters/io"
	modelseeder "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/adapters/seeder"
	modelfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/app/model"
	providersmodule "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers"
	providercache "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/cache"
	securestore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/secretstore"
	providerfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/app/provider"
	"github.com/rs/zerolog"
)

// Dependencies defines infrastructure required to build the application facade.
type Dependencies struct {
	Log                zerolog.Logger
	Config             config.AppConfig
	DB                 *sql.DB
	AppName            string
	KeyringServiceName string
	Events             coreevents.Bus
}

// NewApp builds the single composition root for all application capabilities.
func NewApp(deps Dependencies) (*app.App, error) {

	if deps.DB == nil {
		return nil, fmt.Errorf("app wire: database required")
	}
	if strings.TrimSpace(deps.AppName) == "" {
		return nil, fmt.Errorf("app wire: app name required")
	}
	if strings.TrimSpace(deps.KeyringServiceName) == "" {
		return nil, fmt.Errorf("app wire: keyring service name required")
	}

	coreLog := corelogger.NewAdapter(deps.Log)
	secrets := securestore.NewKeyringStore(deps.KeyringServiceName)

	cache, err := providercache.NewSQLiteStore(deps.DB)
	if err != nil {
		return nil, err
	}
	configStore, err := config.NewSQLiteStore(deps.DB)
	if err != nil {
		return nil, err
	}

	providerService, registry, err := providersmodule.BuildProviderService(deps.Config, cache, secrets, configStore, coreLog)
	if err != nil {
		return nil, err
	}

	chatRepo, err := chatrepo.NewRepository(deps.DB)
	if err != nil {
		return nil, err
	}
	chatService := chatfeature.NewService(chatRepo)
	chatCompletionService := chatfeature.NewChatService(registry, secrets)

	providerOrchestrator := providerfeature.NewOrchestrator(providerService, secrets, deps.Events)
	conversationOrchestrator := chatfeature.NewOrchestrator(chatService, chatCompletionService, deps.Events)
	modelService := modelfeature.NewModelService(
		nil,
		deps.DB,
		deps.AppName,
		modelio.NewLocalFileSystem(),
		modelio.NewPlatformAppDataDirResolver(),
		modelseeder.NewDatastoreSeeder(),
	)
	imageService := imagefeature.NewService(providerOrchestrator, imageresolver.NewHTTPResolver(nil))

	return app.New(app.Dependencies{
		Providers:     app.NewProviderManagement(providerOrchestrator),
		Models:        app.NewModelCatalog(modelService),
		Images:        app.NewImageOperations(imageService),
		Chat:          app.NewChatCompletion(chatCompletionService),
		Conversations: app.NewConversationManagement(conversationOrchestrator),
	}), nil
}
