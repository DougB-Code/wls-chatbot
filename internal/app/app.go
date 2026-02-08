// app.go defines the application facade used by all transport adapters.
// internal/app/app.go
package app

import (
	"context"

	"github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
)

// ProviderManagement exposes provider management capabilities to transports.
type ProviderManagement interface {
	GetProviders(ctx context.Context) ([]contracts.ProviderInfo, error)
	TestProvider(ctx context.Context, name string) error
	AddProvider(ctx context.Context, request contracts.AddProviderRequest) (contracts.ProviderInfo, error)
	UpdateProvider(context.Context, contracts.UpdateProviderRequest) (contracts.ProviderInfo, error)
	RemoveProvider(ctx context.Context, name string) error
	UpdateProviderCredentials(ctx context.Context, request contracts.UpdateProviderCredentialsRequest) error
	SetActiveProvider(name string) bool
	RefreshProviderResources(ctx context.Context, name string) error
	GetActiveProvider() *contracts.ProviderInfo
}

// ModelCatalog exposes model catalog capabilities to transports.
type ModelCatalog interface {
	ListModels(ctx context.Context, filter contracts.ModelListFilter) ([]contracts.ModelSummary, error)
	ImportModels(ctx context.Context, request contracts.ImportModelsRequest) error
	SyncModels(ctx context.Context) (contracts.SyncModelsResult, error)
}

// ImageOperations exposes image generation capabilities to transports.
type ImageOperations interface {
	GenerateImage(ctx context.Context, request contracts.GenerateImageRequest) (contracts.ImageBinaryResult, error)
	EditImage(ctx context.Context, request contracts.EditImageRequest) (contracts.ImageBinaryResult, error)
}

// ChatCompletion exposes stateless chat completion capabilities to transports.
type ChatCompletion interface {
	Chat(ctx context.Context, request contracts.ChatRequest) (<-chan contracts.ChatChunk, error)
}

// ConversationManagement exposes conversation lifecycle and streaming capabilities to transports.
type ConversationManagement interface {
	CreateConversation(providerName, model string) (*contracts.Conversation, error)
	SetActiveConversation(id string)
	GetActiveConversation() *contracts.Conversation
	GetConversation(id string) *contracts.Conversation
	ListConversations() []contracts.ConversationSummary
	ListDeletedConversations() []contracts.ConversationSummary
	UpdateConversationModel(conversationID, model string) bool
	UpdateConversationProvider(conversationID, provider string) bool
	DeleteConversation(id string) bool
	RestoreConversation(id string) bool
	PurgeConversation(id string) bool
	SendMessage(ctx context.Context, conversationID, content string) (*contracts.Message, error)
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
