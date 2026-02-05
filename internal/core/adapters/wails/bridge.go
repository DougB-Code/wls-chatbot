// provide Wails bindings for frontend-backend communication.
// internal/core/adapters/wails/bridge.go
package wails

import (
	"context"

	"github.com/MadeByDoug/wls-chatbot/internal/features/catalog/usecase"
	"github.com/MadeByDoug/wls-chatbot/internal/features/chat/usecase"
	"github.com/MadeByDoug/wls-chatbot/internal/features/notifications/usecase"
	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/usecase"
)

// Bridge is the main Wails binding struct that exposes backend functionality to the frontend.
type Bridge struct {
	chat          *chat.Orchestrator
	providers     *provider.Orchestrator
	catalog       *catalog.Orchestrator
	notifications *notifications.Orchestrator
	emitter       *Emitter

	ctx context.Context
}

// New creates a new Bridge instance.
func New(chatPolicy *chat.Orchestrator, providerPolicy *provider.Orchestrator, catalogPolicy *catalog.Orchestrator, notificationPolicy *notifications.Orchestrator, emitter *Emitter) *Bridge {

	return &Bridge{
		chat:          chatPolicy,
		providers:     providerPolicy,
		catalog:       catalogPolicy,
		notifications: notificationPolicy,
		emitter:       emitter,
	}
}

// Startup is called by Wails when the application starts.
func (b *Bridge) Startup(ctx context.Context) {

	b.ctx = ctx
	if b.emitter != nil {
		b.emitter.SetContext(ctx)
	}
	if b.catalog != nil {
		go func() {
			_ = b.catalog.RefreshAll(b.ctxOrBackground())
		}()
	}
}

// Shutdown is called by Wails when the application shuts down.
func (b *Bridge) Shutdown(context.Context) {

	b.ctx = nil
	if b.emitter != nil {
		b.emitter.ClearContext()
	}
}

// ctxOrBackground returns the app context or a background context if nil.
func (b *Bridge) ctxOrBackground() context.Context {

	if b.ctx != nil {
		return b.ctx
	}
	return context.Background()
}
