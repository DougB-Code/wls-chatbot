// events_test.go verifies typed signal registration and emission helpers.
// internal/core/events/events_test.go
package events

import "testing"

// TestRegisterInRejectsMismatchedPayloadTypes verifies one signal name cannot bind to multiple payload types.
func TestRegisterInRejectsMismatchedPayloadTypes(t *testing.T) {

	registry := NewRegistry()
	if _, err := RegisterIn[string](registry, "sample"); err != nil {
		t.Fatalf("register string signal: %v", err)
	}
	if _, err := RegisterIn[int](registry, "sample"); err == nil {
		t.Fatalf("expected payload type mismatch error")
	}
}

// TestRegisterInAllowsRepeatedType verifies repeated registration with the same type is idempotent.
func TestRegisterInAllowsRepeatedType(t *testing.T) {

	registry := NewRegistry()
	first, err := RegisterIn[string](registry, "sample")
	if err != nil {
		t.Fatalf("register first signal: %v", err)
	}
	second, err := RegisterIn[string](registry, "sample")
	if err != nil {
		t.Fatalf("register repeated signal: %v", err)
	}
	if first.Name() != second.Name() {
		t.Fatalf("expected identical signal names")
	}
}

// TestEmitForwardsPayload verifies typed emit forwards signal name and payload to the bus.
func TestEmitForwardsPayload(t *testing.T) {

	bus := &stubBus{}
	signal := MustRegisterIn[string](NewRegistry(), "emit.test")
	Emit(bus, signal, "hello")

	if bus.name != "emit.test" {
		t.Fatalf("expected signal name emit.test, got %s", bus.name)
	}
	payload, ok := bus.payload.(string)
	if !ok {
		t.Fatalf("expected string payload, got %T", bus.payload)
	}
	if payload != "hello" {
		t.Fatalf("expected payload hello, got %s", payload)
	}
}

type stubBus struct {
	name    Name
	payload interface{}
}

func (s *stubBus) Emit(name Name, payload interface{}) {

	s.name = name
	s.payload = payload
}
