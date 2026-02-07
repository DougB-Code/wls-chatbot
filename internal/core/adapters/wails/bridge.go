// provide Wails bindings for frontend-backend communication.
// internal/core/adapters/wails/bridge.go
package wails

import (
	"context"
	"sync"

	coreinterfaces "github.com/MadeByDoug/wls-chatbot/internal/core/interfaces"
	coreports "github.com/MadeByDoug/wls-chatbot/internal/core/ports"
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
	backend       coreinterfaces.Backend

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var lifecycleUnavailableCtx = canceledContext()

// New creates a new Bridge instance.
func New(chatPolicy *chat.Orchestrator, providerPolicy *provider.Orchestrator, catalogPolicy *catalog.Orchestrator, notificationPolicy *notifications.Orchestrator, emitter *Emitter, backend coreinterfaces.Backend) *Bridge {

	return &Bridge{
		chat:          chatPolicy,
		providers:     providerPolicy,
		catalog:       catalogPolicy,
		notifications: notificationPolicy,
		emitter:       emitter,
		backend:       backend,
	}
}

// Startup is called by Wails when the application starts.
func (b *Bridge) Startup(ctx context.Context) {

	if ctx == nil {
		ctx = context.Background()
	}

	b.mu.RLock()
	previousCancel := b.cancel
	b.mu.RUnlock()
	if previousCancel != nil {
		previousCancel()
		b.wg.Wait()
	}

	appCtx, cancel := context.WithCancel(ctx)
	b.mu.Lock()
	b.ctx = appCtx
	b.cancel = cancel
	b.mu.Unlock()

	if b.emitter != nil {
		b.emitter.SetContext(appCtx)
	}
	if b.catalog != nil {
		b.wg.Add(1)
		go func(refreshCtx context.Context) {
			defer b.wg.Done()
			if err := b.catalog.RefreshAll(refreshCtx); err != nil && refreshCtx.Err() == nil && b.emitter != nil {
				b.emitter.EmitToast(coreports.ToastPayload{
					Type:    "error",
					Title:   "Catalog refresh failed",
					Message: err.Error(),
				})
			}
		}(appCtx)
	}
}

// Shutdown is called by Wails when the application shuts down.
func (b *Bridge) Shutdown(context.Context) {

	if b.emitter != nil {
		b.emitter.ClearContext()
	}

	b.mu.Lock()
	cancel := b.cancel
	b.cancel = nil
	b.ctx = nil
	b.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	b.wg.Wait()
}

// ctxOrBackground returns the app context or a canceled context when lifecycle context is unavailable.
func (b *Bridge) ctxOrBackground() context.Context {

	b.mu.RLock()
	ctx := b.ctx
	b.mu.RUnlock()
	if ctx != nil {
		return ctx
	}
	return lifecycleUnavailableCtx
}

// canceledContext returns an already-canceled context for calls made outside app lifecycle.
func canceledContext() context.Context {

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
