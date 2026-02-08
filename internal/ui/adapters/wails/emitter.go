// emit backend events through the Wails runtime.
// internal/ui/adapters/wails/emitter.go
package wails

import (
	"context"
	"sync"

	coreevents "github.com/MadeByDoug/wls-chatbot/internal/core/events"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Emitter sends backend events to the Wails frontend.
type Emitter struct {
	mu  sync.RWMutex
	ctx context.Context
}

var _ coreevents.Bus = (*Emitter)(nil)

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

// Emit sends a signal payload to the frontend.
func (e *Emitter) Emit(signal coreevents.Name, payload interface{}) {

	ctx := e.context()
	if ctx == nil {
		return
	}
	if payload == nil {
		runtime.EventsEmit(ctx, string(signal))
		return
	}
	runtime.EventsEmit(ctx, string(signal), payload)
}

// context returns the stored Wails context.
func (e *Emitter) context() context.Context {

	e.mu.RLock()
	ctx := e.ctx
	e.mu.RUnlock()
	return ctx
}
