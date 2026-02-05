// orchestrate provider workflows and event emission.
// internal/features/settings/usecase/orchestration.go
package provider

import (
	"context"
	"fmt"
	"strings"

	coreports "github.com/MadeByDoug/wls-chatbot/internal/core/ports"
)

// Orchestrator orchestrates provider workflows and event emission.
type Orchestrator struct {
	providers *Service
	secrets   SecretStore
	emitter   coreports.Emitter
}

// NewOrchestrator creates a provider orchestrator with required dependencies.
func NewOrchestrator(service *Service, secrets SecretStore, emitter coreports.Emitter) *Orchestrator {

	return &Orchestrator{providers: service, secrets: secrets, emitter: emitter}
}

// GetProviders returns all available providers with their status.
func (o *Orchestrator) GetProviders() []Info {

	o.ensureActiveProvider()
	return o.providers.List()
}

// ConnectProvider connects and configures a provider with the given credentials.
func (o *Orchestrator) ConnectProvider(ctx context.Context, name string, credentials ProviderCredentials) (Info, error) {

	info, err := o.providers.Connect(ctx, name, credentials)
	if err == nil {
		o.emitProvidersUpdated()
	}
	return info, err
}

// ConfigureProvider updates a provider's credentials without full connection flow.
func (o *Orchestrator) ConfigureProvider(name string, credentials ProviderCredentials) error {

	return o.providers.Configure(name, credentials)
}

// DisconnectProvider removes a provider's credentials and resets its state.
func (o *Orchestrator) DisconnectProvider(name string) error {

	previousActive := o.providers.GetActiveProvider()
	previousActiveName := ""
	if previousActive != nil {
		previousActiveName = previousActive.Name()
	}

	err := o.providers.Disconnect(name)
	if err != nil {
		return err
	}

	currentActive := o.providers.GetActiveProvider()
	if previousActiveName == name && currentActive != nil && currentActive.Name() != name {
		o.emitProviderSwitchToast(previousActive, currentActive)
	}

	o.emitProvidersUpdated()
	return nil
}

// SetActiveProvider sets the active provider by name.
func (o *Orchestrator) SetActiveProvider(name string) bool {

	ok := o.providers.SetActive(name)
	if ok {
		o.emitProvidersUpdated()
	}
	return ok
}

// TestProvider tests the connection to a provider.
func (o *Orchestrator) TestProvider(ctx context.Context, name string) error {

	return o.providers.TestConnection(ctx, name)
}

// RefreshProviderResources fetches the latest resources from a provider.
func (o *Orchestrator) RefreshProviderResources(ctx context.Context, name string) error {

	err := o.providers.RefreshResources(ctx, name)
	if err == nil {
		o.emitProvidersUpdated()
	}
	return err
}

// GetActiveProvider returns the currently active provider, if any.
func (o *Orchestrator) GetActiveProvider() *Info {

	o.ensureActiveProvider()
	infos := o.providers.List()
	for i := range infos {
		if infos[i].IsActive {
			info := infos[i]
			return &info
		}
	}
	return nil
}

// emitProvidersUpdated publishes a provider update event.
func (o *Orchestrator) emitProvidersUpdated() {

	if o.emitter == nil {
		return
	}
	o.emitter.EmitProvidersUpdated()
}

// emitProviderSwitchToast notifies the frontend about an automatic provider switch.
func (o *Orchestrator) emitProviderSwitchToast(previousActive, currentActive Provider) {

	if o.emitter == nil || currentActive == nil {
		return
	}

	previousName := ""
	if previousActive != nil {
		previousName = previousActive.DisplayName()
	}

	message := fmt.Sprintf("Active provider switched to %s.", currentActive.DisplayName())
	if previousName != "" {
		message = fmt.Sprintf("Active provider switched from %s to %s.", previousName, currentActive.DisplayName())
	}

	o.emitter.EmitToast(coreports.ToastPayload{
		Type:    "info",
		Title:   "Provider switched",
		Message: message,
	})
}

// ensureActiveProvider selects an active provider with valid credentials.
func (o *Orchestrator) ensureActiveProvider() {

	infos := o.providers.List()
	active := o.providers.GetActiveProvider()
	if active != nil {
		for _, info := range infos {
			if info.Name != active.Name() || !info.IsConnected {
				continue
			}
			if prov := o.providers.GetProvider(info.Name); prov != nil {
				if err := o.applyProviderSecrets(info.Name, prov); err == nil {
					return
				}
			}
			break
		}
	}
	for _, info := range infos {
		if !info.IsConnected {
			continue
		}
		prov := o.providers.GetProvider(info.Name)
		if prov == nil {
			continue
		}
		if err := o.applyProviderSecrets(info.Name, prov); err != nil {
			continue
		}
		if o.providers.SetActive(info.Name) {
			return
		}
	}
}

// applyProviderSecrets loads stored secrets into the provider instance.
func (o *Orchestrator) applyProviderSecrets(name string, prov Provider) error {

	if o.secrets == nil {
		return fmt.Errorf("secret store not configured")
	}
	fields := prov.CredentialFields()
	credentials := make(ProviderCredentials)
	for _, field := range fields {
		if !field.Secret {
			continue
		}
		value, err := o.secrets.GetProviderSecret(name, field.Name)
		if err != nil || strings.TrimSpace(value) == "" {
			if field.Required {
				return fmt.Errorf("missing required credential: %s", field.Name)
			}
			continue
		}
		credentials[field.Name] = value
	}
	if len(credentials) > 0 {
		_ = prov.Configure(Config{Credentials: credentials})
	}
	return nil
}
