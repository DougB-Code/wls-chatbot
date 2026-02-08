// orchestration.go orchestrates settings provider workflows and event emission.
// internal/features/settings/app/provider/orchestration.go
package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	coreevents "github.com/MadeByDoug/wls-chatbot/internal/core/events"
)

// Orchestrator orchestrates provider workflows and event emission.
type Orchestrator struct {
	providers *Service
	emitter   coreevents.Bus
	activeMu  sync.Mutex
	activeRun bool
	ensureMu  sync.Mutex
}

// NewOrchestrator creates a provider orchestrator with required dependencies.
func NewOrchestrator(service *Service, _ SecretStore, emitter coreevents.Bus) *Orchestrator {

	return &Orchestrator{providers: service, emitter: emitter}
}

// GetProviders returns all available providers with their status.
func (o *Orchestrator) GetProviders() []Info {

	o.ensureActiveProviderAsync()
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

// GenerateImage generates an image through a configured provider.
func (o *Orchestrator) GenerateImage(ctx context.Context, name string, options ImageGenerationOptions) (*ImageResult, error) {

	prov, err := o.providerByName(name)
	if err != nil {
		return nil, err
	}
	if options.N <= 0 {
		options.N = 1
	}
	return prov.GenerateImage(ctx, options)
}

// EditImage edits an image through a configured provider.
func (o *Orchestrator) EditImage(ctx context.Context, name string, options ImageEditOptions) (*ImageResult, error) {

	prov, err := o.providerByName(name)
	if err != nil {
		return nil, err
	}
	if options.N <= 0 {
		options.N = 1
	}
	return prov.EditImage(ctx, options)
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

	coreevents.Emit(o.emitter, SignalProvidersUpdated, coreevents.EmptyPayload{})
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

	coreevents.Emit(o.emitter, coreevents.SignalToast, coreevents.ToastPayload{
		Type:    "info",
		Title:   "Provider switched",
		Message: message,
	})
}

// ensureActiveProvider selects an active provider with valid credentials.
func (o *Orchestrator) ensureActiveProvider() {

	o.ensureMu.Lock()
	defer o.ensureMu.Unlock()

	infos := o.providers.List()
	active := o.providers.GetActiveProvider()
	if active != nil {
		for _, info := range infos {
			if info.Name != active.Name() || !info.IsConnected {
				continue
			}
			if err := o.providers.EnsureProviderConfigured(info.Name); err == nil {
				return
			}
			break
		}
	}
	for _, info := range infos {
		if !info.IsConnected {
			continue
		}
		if err := o.providers.EnsureProviderConfigured(info.Name); err != nil {
			continue
		}
		if o.providers.SetActive(info.Name) {
			return
		}
	}
}

// ensureActiveProviderAsync de-duplicates background active-provider checks.
func (o *Orchestrator) ensureActiveProviderAsync() {

	o.activeMu.Lock()
	if o.activeRun {
		o.activeMu.Unlock()
		return
	}
	o.activeRun = true
	o.activeMu.Unlock()

	go func() {
		defer func() {
			o.activeMu.Lock()
			o.activeRun = false
			o.activeMu.Unlock()
		}()
		o.ensureActiveProvider()
	}()
}

// providerByName resolves and configures a provider before runtime operations.
func (o *Orchestrator) providerByName(name string) (Provider, error) {

	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, fmt.Errorf("provider name required")
	}

	resolvedName := trimmed
	infos := o.providers.List()
	for _, info := range infos {
		if strings.EqualFold(info.Name, trimmed) {
			resolvedName = info.Name
			break
		}
	}

	if err := o.providers.EnsureProviderConfigured(resolvedName); err != nil {
		return nil, err
	}

	prov := o.providers.GetProvider(resolvedName)
	if prov == nil {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return prov, nil
}
