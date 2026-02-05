// manage provider input persistence for non-secret fields.
// internal/features/settings/ports/provider_inputs_store.go
package ports

// ProviderInputsStore persists non-secret provider input values.
type ProviderInputsStore interface {
	LoadProviderInputs(providerName string) (map[string]string, error)
	SaveProviderInputs(providerName string, inputs map[string]string) error
}
