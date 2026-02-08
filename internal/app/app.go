// app.go defines the application facade used by all transport adapters.
// internal/app/app.go
package app

import (
	"context"

	chatfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/app/chat"
	chatports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/ports"
	imageports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/ports"
	modelinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/interfaces"
	providerfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/provider"
)

// ProviderManagement exposes provider management capabilities to transports.
type ProviderManagement interface {
	GetProviders(ctx context.Context) ([]providerfeature.Info, error)
	TestProvider(ctx context.Context, name string) error
	AddProvider(ctx context.Context, name string, credentials providerfeature.ProviderCredentials) (providerfeature.Info, error)
	RemoveProvider(ctx context.Context, name string) error
	UpdateProviderCredentials(ctx context.Context, name string, credentials providerfeature.ProviderCredentials) error
	SetActiveProvider(name string) bool
	RefreshProviderResources(ctx context.Context, name string) error
	GetActiveProvider() *providerfeature.Info
}

// ModelCatalog exposes model catalog capabilities to transports.
type ModelCatalog interface {
	ListModels(ctx context.Context, filter modelinterfaces.ModelListFilter) ([]modelinterfaces.ModelSummary, error)
	ImportModels(ctx context.Context, request modelinterfaces.ImportModelsRequest) error
	SyncModels(ctx context.Context) (modelinterfaces.SyncModelsResult, error)
}

// ImageOperations exposes image generation capabilities to transports.
type ImageOperations interface {
	GenerateImage(ctx context.Context, request imageports.GenerateImageRequest) (imageports.ImageBinaryResult, error)
	EditImage(ctx context.Context, request imageports.EditImageRequest) (imageports.ImageBinaryResult, error)
}

// ChatCompletion exposes stateless chat completion capabilities to transports.
type ChatCompletion interface {
	Chat(ctx context.Context, request chatports.ChatRequest) (<-chan chatports.ChatChunk, error)
}

// ConversationManagement exposes conversation lifecycle and streaming capabilities to transports.
type ConversationManagement interface {
	CreateConversation(providerName, model string) (*chatfeature.Conversation, error)
	SetActiveConversation(id string)
	GetActiveConversation() *chatfeature.Conversation
	GetConversation(id string) *chatfeature.Conversation
	ListConversations() []chatfeature.ConversationSummary
	ListDeletedConversations() []chatfeature.ConversationSummary
	UpdateConversationModel(conversationID, model string) bool
	UpdateConversationProvider(conversationID, provider string) bool
	DeleteConversation(id string) bool
	RestoreConversation(id string) bool
	PurgeConversation(id string) bool
	SendMessage(ctx context.Context, conversationID, content string) (*chatfeature.Message, error)
	StopStream()
}

// App groups feature capabilities behind one application facade.
type App struct {
	Providers     ProviderManagement
	Models        ModelCatalog
	Images        ImageOperations
	Chat          ChatCompletion
	Conversations ConversationManagement
}

// Dependencies contains capability implementations used to build the facade.
type Dependencies struct {
	Providers     ProviderManagement
	Models        ModelCatalog
	Images        ImageOperations
	Chat          ChatCompletion
	Conversations ConversationManagement
}

// New constructs an application facade from capability implementations.
func New(deps Dependencies) *App {

	return &App{
		Providers:     deps.Providers,
		Models:        deps.Models,
		Images:        deps.Images,
		Chat:          deps.Chat,
		Conversations: deps.Conversations,
	}
}
