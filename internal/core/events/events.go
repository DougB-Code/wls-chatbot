// events.go defines generic typed event signal contracts and registration.
// internal/core/events/events.go
package events

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Name identifies a frontend/backend event signal.
type Name string

// Signal is a strongly typed event signal identifier.
type Signal[T any] struct {
	name Name
}

// EmptyPayload is used for signals that intentionally carry no payload.
type EmptyPayload struct{}

// ToastPayload describes a generic toast notification signal payload.
type ToastPayload struct {
	Type    string `json:"type,omitempty"`
	Title   string `json:"title,omitempty"`
	Message string `json:"message,omitempty"`
}

// SignalToast is a shared UI signal for toast notifications.
var SignalToast = MustRegister[ToastPayload]("toast")

// Bus emits signals for transport adapters.
type Bus interface {
	Emit(signal Name, payload interface{})
}

// Emit publishes a strongly typed payload to a registered signal.
func Emit[T any](bus Bus, signal Signal[T], payload T) {

	if bus == nil {
		return
	}
	bus.Emit(signal.name, payload)
}

// Registry stores typed signal registrations.
type Registry struct {
	mu    sync.RWMutex
	types map[Name]reflect.Type
}

var defaultRegistry = NewRegistry()

// NewRegistry creates a typed signal registry.
func NewRegistry() *Registry {

	return &Registry{
		types: make(map[Name]reflect.Type),
	}
}

// register reserves a signal name for a payload type.
func (r *Registry) register(name Name, payloadType reflect.Type) (Name, error) {

	normalized, err := normalizeName(name)
	if err != nil {
		return "", err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.types[normalized]
	if exists {
		if existing == payloadType {
			return normalized, nil
		}
		return "", fmt.Errorf(
			"event registry: signal %q already registered with payload %s (requested %s)",
			normalized,
			existing.String(),
			payloadType.String(),
		)
	}

	r.types[normalized] = payloadType
	return normalized, nil
}

// RegisterIn adds a typed signal to a specific registry.
func RegisterIn[T any](registry *Registry, name Name) (Signal[T], error) {

	if registry == nil {
		return Signal[T]{}, fmt.Errorf("event registry: registry required")
	}

	registeredName, err := registry.register(name, payloadTypeOf[T]())
	if err != nil {
		return Signal[T]{}, err
	}
	return Signal[T]{name: registeredName}, nil
}

// MustRegisterIn adds a typed signal and panics when registration fails.
func MustRegisterIn[T any](registry *Registry, name Name) Signal[T] {

	signal, err := RegisterIn[T](registry, name)
	if err != nil {
		panic(err)
	}
	return signal
}

// MustRegister adds a typed signal to the default registry and panics on failure.
func MustRegister[T any](name Name) Signal[T] {

	return MustRegisterIn[T](defaultRegistry, name)
}

func normalizeName(name Name) (Name, error) {

	normalized := Name(strings.TrimSpace(string(name)))
	if normalized == "" {
		return "", fmt.Errorf("event registry: signal name required")
	}
	return normalized, nil
}

func payloadTypeOf[T any]() reflect.Type {

	var pointer *T
	return reflect.TypeOf(pointer).Elem()
}
