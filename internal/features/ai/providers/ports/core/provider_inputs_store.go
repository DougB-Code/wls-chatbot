// provider_inputs_store.go manages provider input persistence for non-secret fields.
// internal/features/ai/providers/ports/core/provider_inputs_store.go
package core

// ProviderInputsStore persists non-secret provider input values.
type ProviderInputsStore interface {
	LoadProviderInputs(providerName string) (map[string]string, error)
	SaveProviderInputs(providerName string, inputs map[string]string) error
}
