// emit backend events through the Wails runtime.
// internal/core/adapters/wails/emitter.go
package wails

import (
	"context"
	"sync"

	"github.com/MadeByDoug/wls-chatbot/internal/core/ports"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Emitter sends backend events to the Wails frontend.
type Emitter struct {
	mu  sync.RWMutex
	ctx context.Context
}

var _ ports.Emitter = (*Emitter)(nil)

// SetContext sets the Wails context used for event emission.
func (e *Emitter) SetContext(ctx context.Context) {

	e.mu.Lock()
	e.ctx = ctx
	e.mu.Unlock()
}

// ClearContext clears the Wails context.
func (e *Emitter) ClearContext() {

	e.mu.Lock()
	e.ctx = nil
	e.mu.Unlock()
}

// EmitChatEvent sends a chat event to the frontend.
func (e *Emitter) EmitChatEvent(event ports.ChatEvent) {

	ctx := e.context()
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, "chat:event", event)
}

// EmitProvidersUpdated notifies the frontend that providers have changed.
func (e *Emitter) EmitProvidersUpdated() {

	ctx := e.context()
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, "providers:updated")
}

// EmitCatalogUpdated notifies the frontend that catalog data has changed.
func (e *Emitter) EmitCatalogUpdated() {

	ctx := e.context()
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, "catalog:updated")
}

// EmitToast sends a toast notification to the frontend.
func (e *Emitter) EmitToast(payload ports.ToastPayload) {

	ctx := e.context()
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, "toast", payload)
}

// context returns the stored Wails context.
func (e *Emitter) context() context.Context {

	e.mu.RLock()
	ctx := e.ctx
	e.mu.RUnlock()
	return ctx
}
