// service.go manages provider connections, resources, and status for settings.
// internal/features/settings/app/provider/service.go
package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/interfaces/core"
)

// Info represents provider information for the frontend.
type Info struct {
	Name             string            `json:"name"`
	DisplayName      string            `json:"displayName"`
	CredentialFields []CredentialField `json:"credentialFields,omitempty"`
	CredentialValues map[string]string `json:"credentialValues,omitempty"`
	Models           []Model           `json:"models"`
	Resources        []Model           `json:"resources"`
	IsConnected      bool              `json:"isConnected"`
	IsActive         bool              `json:"isActive"`
	Status           *Status           `json:"status,omitempty"`
}

// Status represents the last known health check for a provider.
type Status struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message,omitempty"`
	CheckedAt int64  `json:"checkedAt"`
}

// resourceLister is implemented by providers that can list resources.
type resourceLister interface {
	ListResources(ctx context.Context) ([]Model, error)
}

var errNoResources = errors.New("no resources returned from provider")

// Service handles provider management logic.
type Service struct {
	registry          Registry
	resources         map[string][]Model
	resourceUpdatedAt map[string]int64
	cache             Cache
	enabledModelIDs   map[string][]string
	updateFrequency   map[string]time.Duration
	status            map[string]Status
	refreshing        map[string]bool
	mu                sync.RWMutex
	inputsStore       InputsStore
	secrets           SecretStore
	logger            Logger
	providerOpsMu     sync.Mutex
}

// validateProvider tests connectivity and updates resources when supported.
func (s *Service) validateProvider(ctx context.Context, name string, p Provider) (bool, error) {

	if lister, ok := p.(resourceLister); ok {
		resources, err := lister.ListResources(ctx)
		if err != nil {
			return true, err
		}
		if len(resources) == 0 {
			return true, errNoResources
		}
		s.SetResources(name, resources)
		return true, nil
	}
	if err := p.TestConnection(ctx); err != nil {
		return false, err
	}
	return false, nil
}

// NewService creates a new provider service.
func NewService(registry Registry, cache Cache, secrets SecretStore, inputs InputsStore, updateFrequency map[string]time.Duration, logger Logger) *Service {

	s := &Service{
		registry:          registry,
		resources:         make(map[string][]Model),
		resourceUpdatedAt: make(map[string]int64),
		cache:             cache,
		enabledModelIDs:   captureEnabledModelIDs(registry),
		updateFrequency:   copyUpdateFrequency(updateFrequency),
		status:            make(map[string]Status),
		refreshing:        make(map[string]bool),
		inputsStore:       inputs,
		secrets:           secrets,
		logger:            logger,
	}
	s.loadCache()
	s.applyEnabledModelsFromCache()
	return s
}

// loadCache loads cached provider resources from disk.
func (s *Service) loadCache() {

	if s.cache == nil {
		return
	}
	snapshot, err := s.cache.Load()
	if err != nil {
		return
	}

	for name, entry := range snapshot {
		if entry.Models != nil {
			s.resources[name] = entry.Models
		}
		if entry.UpdatedAt > 0 {
			s.resourceUpdatedAt[name] = entry.UpdatedAt
		}
	}
}

// applyEnabledModelsFromCache applies cached resources to enabled model lists.
func (s *Service) applyEnabledModelsFromCache() {

	if s.registry == nil {
		return
	}
	for _, prov := range s.registry.List() {
		name := prov.Name()
		resources := s.GetResources(name)
		s.applyEnabledModels(name, resources)
	}
}

// applyEnabledModels updates provider models using enabled IDs and available resources.
func (s *Service) applyEnabledModels(name string, resources []Model) {

	enabledIDs := s.enabledModelIDs[name]
	if p := s.registry.Get(name); p != nil {
		if len(resources) == 0 {
			_ = p.Configure(Config{Models: buildFallbackModels(enabledIDs)})
			return
		}
		if len(enabledIDs) == 0 {
			_ = p.Configure(Config{Models: resources})
			return
		}
		available := indexModelsByID(resources)
		_ = p.Configure(Config{Models: selectModelsByID(enabledIDs, available)})
	}
}

// captureEnabledModelIDs extracts enabled model IDs from configured providers.
func captureEnabledModelIDs(registry Registry) map[string][]string {

	result := make(map[string][]string)
	if registry == nil {
		return result
	}
	for _, prov := range registry.List() {
		result[prov.Name()] = extractModelIDs(prov.Models())
	}
	return result
}

// extractModelIDs normalizes model IDs from configured models.
func extractModelIDs(models []Model) []string {

	seen := make(map[string]struct{}, len(models))
	ids := make([]string, 0, len(models))
	for _, model := range models {
		trimmed := strings.TrimSpace(model.ID)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		ids = append(ids, trimmed)
	}
	return ids
}

// providerCredentialFields returns the credential schema for a provider.
func (s *Service) providerCredentialFields(p Provider) []CredentialField {

	if p == nil {
		return nil
	}
	return p.CredentialFields()
}

// loadProviderInputs returns stored non-secret inputs for a provider.
func (s *Service) loadProviderInputs(name string) ProviderCredentials {

	if s.inputsStore == nil {
		return nil
	}
	inputs, err := s.inputsStore.LoadProviderInputs(name)
	if err != nil {
		s.logWarn("Failed to load provider inputs", err, LogField{Key: "provider", Value: name})
		return nil
	}
	return inputs
}

// loadProviderSecrets returns stored secret values for a provider.
func (s *Service) loadProviderSecrets(name string, fields []CredentialField) ProviderCredentials {

	if s.secrets == nil {
		return nil
	}

	credentials := make(ProviderCredentials)
	for _, field := range fields {
		if !field.Secret {
			continue
		}
		value, err := s.secrets.GetProviderSecret(name, field.Name)
		if err != nil || strings.TrimSpace(value) == "" {
			continue
		}
		credentials[field.Name] = value
	}

	if len(credentials) == 0 {
		return nil
	}
	return credentials
}

// mergeCredentialValues combines stored and incoming credential values.
func mergeCredentialValues(base, override ProviderCredentials) ProviderCredentials {

	if len(base) == 0 && len(override) == 0 {
		return nil
	}

	merged := make(ProviderCredentials)
	for key, value := range base {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		merged[key] = trimmed
	}
	for key, value := range override {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		merged[key] = trimmed
	}

	if len(merged) == 0 {
		return nil
	}
	return merged
}

// validateRequiredCredentials verifies all required fields are present.
func validateRequiredCredentials(fields []CredentialField, credentials ProviderCredentials) error {

	for _, field := range fields {
		if !field.Required {
			continue
		}
		if strings.TrimSpace(credentials[field.Name]) == "" {
			label := field.Label
			if label == "" {
				label = field.Name
			}
			return fmt.Errorf("missing required credential: %s", label)
		}
	}
	return nil
}

// filterCredentialValues returns credential values matching the secret flag.
func filterCredentialValues(fields []CredentialField, credentials ProviderCredentials, secret bool) ProviderCredentials {

	if len(credentials) == 0 {
		return nil
	}

	filtered := make(ProviderCredentials)
	for _, field := range fields {
		if field.Secret != secret {
			continue
		}
		value := strings.TrimSpace(credentials[field.Name])
		if value == "" {
			continue
		}
		filtered[field.Name] = value
	}

	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

// resolveCredentials merges stored and incoming credentials with validation.
func (s *Service) resolveCredentials(name string, fields []CredentialField, input ProviderCredentials) (ProviderCredentials, error) {

	stored := mergeCredentialValues(s.loadProviderInputs(name), s.loadProviderSecrets(name, fields))
	merged := mergeCredentialValues(stored, input)
	if err := validateRequiredCredentials(fields, merged); err != nil {
		return nil, err
	}
	return merged, nil
}

// persistCredentials saves provided credential values to storage.
func (s *Service) persistCredentials(name string, fields []CredentialField, input ProviderCredentials) error {

	if len(input) == 0 {
		return nil
	}

	var secretUpdates ProviderCredentials
	var inputUpdates ProviderCredentials

	for _, field := range fields {
		value, ok := input[field.Name]
		if !ok {
			continue
		}
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if field.Secret {
			if secretUpdates == nil {
				secretUpdates = make(ProviderCredentials)
			}
			secretUpdates[field.Name] = trimmed
		} else {
			if providercore.IsSensitiveCredentialName(field.Name) {
				return fmt.Errorf("credential field %q must be stored as secret", field.Name)
			}
			if inputUpdates == nil {
				inputUpdates = make(ProviderCredentials)
			}
			inputUpdates[field.Name] = trimmed
		}
	}

	if len(secretUpdates) > 0 {
		if s.secrets == nil {
			return fmt.Errorf("secret store not configured")
		}
		for fieldName, value := range secretUpdates {
			if err := s.secrets.SaveProviderSecret(name, fieldName, value); err != nil {
				return err
			}
		}
	}

	if len(inputUpdates) > 0 {
		if s.inputsStore == nil {
			return fmt.Errorf("config store not configured")
		}
		mergedInputs := mergeCredentialValues(s.loadProviderInputs(name), inputUpdates)
		if err := s.inputsStore.SaveProviderInputs(name, mergedInputs); err != nil {
			return err
		}
	}

	return nil
}

// clearStoredCredentials removes stored inputs and secrets for a provider.
func (s *Service) clearStoredCredentials(name string, fields []CredentialField) error {

	if s.inputsStore != nil {
		if err := s.inputsStore.SaveProviderInputs(name, nil); err != nil {
			return err
		}
	}
	if s.secrets == nil {
		return nil
	}
	for _, field := range fields {
		if !field.Secret {
			continue
		}
		_ = s.secrets.DeleteProviderSecret(name, field.Name)
	}
	return nil
}

// isProviderConfigured returns true when required credential fields are present.
func (s *Service) isProviderConfigured(name string, fields []CredentialField) bool {

	credentials := mergeCredentialValues(s.loadProviderInputs(name), s.loadProviderSecrets(name, fields))
	return validateRequiredCredentials(fields, credentials) == nil
}

// refreshResourcesIfStale launches a background refresh when cache is outdated.
func (s *Service) refreshResourcesIfStale(name string) {

	if !s.shouldRefreshResources(name) {
		return
	}
	if s.registry == nil {
		return
	}
	if !s.isProviderConfigured(name, s.providerCredentialFields(s.registry.Get(name))) {
		return
	}
	if !s.markRefreshing(name) {
		return
	}

	go func() {
		defer s.clearRefreshing(name)
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		if err := s.RefreshResources(ctx, name); err != nil {
			s.logWarn("Failed to refresh stale resources", err, LogField{Key: "provider", Value: name})
		}
	}()
}

// shouldRefreshResources determines if cached resources are stale.
func (s *Service) shouldRefreshResources(name string) bool {

	frequency := s.getUpdateFrequency(name)
	if frequency <= 0 {
		return false
	}
	lastUpdated := s.getResourceUpdatedAt(name)
	if lastUpdated == 0 {
		return true
	}
	lastUpdateTime := time.UnixMilli(lastUpdated)
	return time.Since(lastUpdateTime) >= frequency
}

// getUpdateFrequency returns the configured update cadence for a provider.
func (s *Service) getUpdateFrequency(name string) time.Duration {

	if s.updateFrequency == nil {
		return 0
	}
	return s.updateFrequency[name]
}

// getResourceUpdatedAt returns the last update timestamp for cached resources.
func (s *Service) getResourceUpdatedAt(name string) int64 {

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.resourceUpdatedAt[name]
}

// markRefreshing marks a provider as having an in-flight refresh.
func (s *Service) markRefreshing(name string) bool {

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.refreshing[name] {
		return false
	}
	s.refreshing[name] = true
	return true
}

// clearRefreshing clears the in-flight refresh marker.
func (s *Service) clearRefreshing(name string) {

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.refreshing, name)
}

// ensureProviderConfiguredLocked applies stored credentials while providerOpsMu is held.
func (s *Service) ensureProviderConfiguredLocked(name string) error {

	if s.registry == nil {
		return nil
	}
	p := s.registry.Get(name)
	if p == nil {
		return fmt.Errorf("provider not found: %s", name)
	}
	fields := s.providerCredentialFields(p)
	resolved, err := s.resolveCredentials(name, fields, nil)
	if err != nil {
		return err
	}
	if len(resolved) == 0 {
		return nil
	}
	if err := p.Configure(Config{Credentials: resolved}); err != nil {
		return err
	}
	return nil
}

// EnsureProviderConfigured applies persisted credentials for a provider under serialization lock.
func (s *Service) EnsureProviderConfigured(name string) error {

	s.providerOpsMu.Lock()
	defer s.providerOpsMu.Unlock()
	return s.ensureProviderConfiguredLocked(name)
}

// List returns all available providers with their status.
func (s *Service) List() []Info {

	if s.registry == nil {
		return nil
	}
	providers := s.registry.List()
	active := s.registry.GetActive()

	info := make([]Info, len(providers))
	for i, p := range providers {
		// Trigger stale-check; method schedules background refresh only when needed.
		s.refreshResourcesIfStale(p.Name())
		fields := s.providerCredentialFields(p)
		isConfigured := s.isProviderConfigured(p.Name(), fields)
		hasHealthyStatus := s.hasSuccessfulStatus(p.Name())
		// Skip loading inputs during list to avoid blocking - credentials are only needed on connect/configure
		info[i] = Info{
			Name:             p.Name(),
			DisplayName:      p.DisplayName(),
			CredentialFields: fields,
			CredentialValues: nil, // Load on demand, not during list
			Models:           p.Models(),
			Resources:        s.GetResources(p.Name()),
			IsConnected:      isConfigured || hasHealthyStatus,
			IsActive:         active != nil && active.Name() == p.Name(),
			Status:           s.GetStatus(p.Name()),
		}
	}
	return info
}

// Connect configures, validates, and persists a provider connection.
func (s *Service) Connect(ctx context.Context, name string, credentials ProviderCredentials) (Info, error) {

	s.providerOpsMu.Lock()
	defer s.providerOpsMu.Unlock()

	s.logInfo("Connecting provider", LogField{Key: "provider", Value: name})
	p := s.registry.Get(name)
	if p == nil {
		err := fmt.Errorf("provider not found: %s", name)
		s.SetStatus(name, false, err.Error())
		s.logWarn("Provider not found during connect", err, LogField{Key: "provider", Value: name})
		return Info{}, err
	}

	fields := s.providerCredentialFields(p)
	resolved, err := s.resolveCredentials(name, fields, credentials)
	if err != nil {
		s.SetStatus(name, false, err.Error())
		s.logWarn("Missing required credentials", err, LogField{Key: "provider", Value: name})
		return Info{}, err
	}

	if err := p.Configure(Config{Credentials: resolved}); err != nil {
		s.SetStatus(name, false, err.Error())
		s.logError("Failed to configure provider", err, LogField{Key: "provider", Value: name})
		return Info{}, err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	// Use a timeout for the connection test/list
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	usedLister, err := s.validateProvider(ctx, name, p)
	if err != nil {
		s.SetStatus(name, false, err.Error())
		if usedLister {
			if errors.Is(err, errNoResources) {
				s.logWarn("No resources returned", err, LogField{Key: "provider", Value: name})
			} else {
				s.logError("Failed to list resources", err, LogField{Key: "provider", Value: name})
			}
		} else {
			s.logError("Connection test failed", err, LogField{Key: "provider", Value: name})
		}
		return Info{}, err
	}

	if err := s.persistCredentials(name, fields, credentials); err != nil {
		s.SetStatus(name, false, err.Error())
		s.logError("Failed to save credentials", err, LogField{Key: "provider", Value: name})
		return Info{}, err
	}

	s.registry.SetActive(name)
	s.logInfo("Provider connected successfully", LogField{Key: "provider", Value: name})

	active := s.registry.GetActive()
	s.SetStatus(name, true, "")
	inputs := s.loadProviderInputs(p.Name())
	return Info{
		Name:             p.Name(),
		DisplayName:      p.DisplayName(),
		CredentialFields: fields,
		CredentialValues: filterCredentialValues(fields, inputs, false),
		Models:           p.Models(),
		Resources:        s.GetResources(p.Name()),
		IsConnected:      s.isProviderConfigured(p.Name(), fields),
		IsActive:         active != nil && active.Name() == p.Name(),
		Status:           s.GetStatus(p.Name()),
	}, nil
}

// Disconnect removes a provider's credentials and resets its state.
func (s *Service) Disconnect(name string) error {

	s.providerOpsMu.Lock()
	defer s.providerOpsMu.Unlock()

	s.logInfo("Disconnecting provider", LogField{Key: "provider", Value: name})
	fields := s.providerCredentialFields(s.registry.Get(name))
	if err := s.clearStoredCredentials(name, fields); err != nil {
		s.logError("Failed to remove credentials", err, LogField{Key: "provider", Value: name})
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	// 2. Clear from registry/memory
	p := s.registry.Get(name)
	if p != nil {
		clear := make(ProviderCredentials)
		for _, field := range fields {
			clear[field.Name] = ""
		}
		_ = p.Configure(Config{Credentials: clear})
	}

	// 3. Clear cached resources
	s.SetResources(name, nil)
	s.ClearStatus(name)

	// 4. Update active provider if needed
	active := s.registry.GetActive()
	if active != nil && active.Name() == name {
		nextActive := s.selectNextActiveProvider(name)
		if nextActive == "" {
			_ = s.registry.SetActive("")
		} else if s.registry.SetActive(nextActive) {
			if err := s.ensureProviderConfiguredLocked(nextActive); err != nil {
				_ = s.registry.SetActive("")
			}
		} else {
			_ = s.registry.SetActive("")
		}
	}

	return nil
}

// selectNextActiveProvider finds the next connected provider after the given name.
func (s *Service) selectNextActiveProvider(disconnected string) string {

	if s.registry == nil {
		return ""
	}

	providers := s.registry.List()
	if len(providers) == 0 {
		return ""
	}

	startIndex := 0
	for i, p := range providers {
		if p != nil && p.Name() == disconnected {
			startIndex = i + 1
			break
		}
	}

	for offset := 0; offset < len(providers); offset++ {
		index := (startIndex + offset) % len(providers)
		candidate := providers[index]
		if candidate == nil {
			continue
		}
		name := candidate.Name()
		if name == "" || name == disconnected {
			continue
		}
		if s.isProviderConfigured(name, s.providerCredentialFields(candidate)) {
			return name
		}
	}

	return ""
}

// Configure updates and persists provider credentials while refreshing status.
func (s *Service) Configure(name string, credentials ProviderCredentials) error {

	s.providerOpsMu.Lock()
	defer s.providerOpsMu.Unlock()

	s.logInfo("Configuring provider", LogField{Key: "provider", Value: name})
	p := s.registry.Get(name)
	if p == nil {
		err := fmt.Errorf("provider not found: %s", name)
		s.SetStatus(name, false, err.Error())
		s.logWarn("Provider not found during configure", err, LogField{Key: "provider", Value: name})
		return err
	}
	trimmedInput := mergeCredentialValues(nil, credentials)
	if len(trimmedInput) == 0 {
		err := fmt.Errorf("credentials required for provider: %s", name)
		s.SetStatus(name, false, err.Error())
		s.logWarn("Empty credentials during configure", err, LogField{Key: "provider", Value: name})
		return err
	}
	fields := s.providerCredentialFields(p)
	resolved, err := s.resolveCredentials(name, fields, credentials)
	if err != nil {
		s.SetStatus(name, false, err.Error())
		s.logWarn("Missing required credentials", err, LogField{Key: "provider", Value: name})
		return err
	}
	if err := p.Configure(Config{Credentials: resolved}); err != nil {
		s.SetStatus(name, false, err.Error())
		s.logError("Failed to configure provider", err, LogField{Key: "provider", Value: name})
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	usedLister, err := s.validateProvider(ctx, name, p)
	if err != nil {
		s.SetStatus(name, false, err.Error())
		if usedLister {
			if errors.Is(err, errNoResources) {
				s.logWarn("No resources returned", err, LogField{Key: "provider", Value: name})
			} else {
				s.logError("Failed to list resources", err, LogField{Key: "provider", Value: name})
			}
		} else {
			s.logError("Connection test failed", err, LogField{Key: "provider", Value: name})
		}
		return err
	}

	if err := s.persistCredentials(name, fields, credentials); err != nil {
		s.SetStatus(name, false, err.Error())
		s.logError("Failed to save credentials", err, LogField{Key: "provider", Value: name})
		return err
	}

	s.SetStatus(name, true, "")
	return nil
}

// SetActive sets the active provider.
func (s *Service) SetActive(name string) bool {

	s.providerOpsMu.Lock()
	defer s.providerOpsMu.Unlock()

	s.logInfo("Setting active provider", LogField{Key: "provider", Value: name})
	return s.registry.SetActive(name)
}

// TestConnection tests the connection to a provider.
func (s *Service) TestConnection(ctx context.Context, name string) error {

	s.providerOpsMu.Lock()
	defer s.providerOpsMu.Unlock()
	return s.testConnectionLocked(ctx, name)
}

// testConnectionLocked tests provider connectivity while providerOpsMu is held.
func (s *Service) testConnectionLocked(ctx context.Context, name string) error {

	p := s.registry.Get(name)
	if p == nil {
		err := fmt.Errorf("provider not found: %s", name)
		s.SetStatus(name, false, err.Error())
		return err
	}
	fields := s.providerCredentialFields(p)
	resolved, err := s.resolveCredentials(name, fields, nil)
	if err != nil {
		s.SetStatus(name, false, err.Error())
		return err
	}
	if err := p.Configure(Config{Credentials: resolved}); err != nil {
		s.SetStatus(name, false, err.Error())
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := p.TestConnection(ctx); err != nil {
		s.SetStatus(name, false, err.Error())
		s.logWarn("Test connection failed", err, LogField{Key: "provider", Value: name})
		return err
	}
	s.SetStatus(name, true, "")
	return nil
}

// RefreshResources fetches the latest resources from the provider.
func (s *Service) RefreshResources(ctx context.Context, name string) error {

	s.providerOpsMu.Lock()
	defer s.providerOpsMu.Unlock()

	s.logDebug("Refreshing resources", LogField{Key: "provider", Value: name})
	if ctx == nil {
		ctx = context.Background()
	}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		timeoutCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
		defer cancel()
		ctx = timeoutCtx
	}
	p := s.registry.Get(name)
	if p == nil {
		err := fmt.Errorf("provider not found: %s", name)
		s.SetStatus(name, false, err.Error())
		return err
	}

	fields := s.providerCredentialFields(p)
	resolved, err := s.resolveCredentials(name, fields, nil)
	if err != nil {
		s.SetStatus(name, false, err.Error())
		return err
	}
	if err := p.Configure(Config{Credentials: resolved}); err != nil {
		s.SetStatus(name, false, err.Error())
		return err
	}
	if lister, ok := p.(resourceLister); ok {
		resources, err := lister.ListResources(ctx)
		if err != nil {
			s.SetStatus(name, false, err.Error())
			s.logError("Failed to refresh resources", err, LogField{Key: "provider", Value: name})
			return err
		}
		if len(resources) == 0 {
			err := fmt.Errorf("no resources returned from provider")
			s.SetStatus(name, false, err.Error())
			return err
		}
		s.SetResources(name, resources)
		s.SetStatus(name, true, "")
		return nil
	}
	if err := s.testConnectionLocked(ctx, name); err != nil {
		return err
	}
	return nil
}

// GetResources returns a copy of cached resources for a provider.
func (s *Service) GetResources(name string) []Model {

	s.mu.RLock()
	resources := s.resources[name]
	s.mu.RUnlock()

	if resources == nil {
		return nil
	}

	cloned := make([]Model, len(resources))
	copy(cloned, resources)
	return cloned
}

// SetResources updates cached resources for a provider.
func (s *Service) SetResources(name string, resources []Model) {

	s.mu.Lock()
	if resources == nil {
		delete(s.resources, name)
		delete(s.resourceUpdatedAt, name)
	} else {
		s.resources[name] = resources
		s.resourceUpdatedAt[name] = time.Now().UnixMilli()
	}
	snapshot := s.buildCacheSnapshotLocked()
	s.mu.Unlock()

	if s.cache != nil {
		_ = s.cache.Save(snapshot)
	}
	s.applyEnabledModels(name, resources)
}

// buildCacheSnapshotLocked generates the cache snapshot from locked state.
func (s *Service) buildCacheSnapshotLocked() CacheSnapshot {

	snapshot := make(CacheSnapshot, len(s.resources))
	for name, resources := range s.resources {
		snapshot[name] = CacheEntry{
			UpdatedAt: s.resourceUpdatedAt[name],
			Models:    resources,
		}
	}
	return snapshot
}

// indexModelsByID builds a lookup map for models.
func indexModelsByID(models []Model) map[string]Model {

	index := make(map[string]Model, len(models))
	for _, model := range models {
		if model.ID == "" {
			continue
		}
		index[model.ID] = model
	}
	return index
}

// selectModelsByID returns models in the order of provided IDs.
func selectModelsByID(ids []string, available map[string]Model) []Model {

	selected := make([]Model, 0, len(ids))
	for _, id := range ids {
		if model, ok := available[id]; ok {
			selected = append(selected, model)
		}
	}
	return selected
}

// buildFallbackModels constructs model structs from IDs.
func buildFallbackModels(ids []string) []Model {

	models := make([]Model, 0, len(ids))
	for _, id := range ids {
		models = append(models, Model{
			ID:   id,
			Name: id,
		})
	}
	return models
}

// copyUpdateFrequency clones the update frequency map.
func copyUpdateFrequency(input map[string]time.Duration) map[string]time.Duration {

	if input == nil {
		return nil
	}
	result := make(map[string]time.Duration, len(input))
	for name, value := range input {
		result[name] = value
	}
	return result
}

// GetStatus returns the last recorded status for a provider.
func (s *Service) GetStatus(name string) *Status {

	s.mu.RLock()
	defer s.mu.RUnlock()
	status, ok := s.status[name]
	if !ok {
		return nil
	}
	result := status
	return &result
}

// hasSuccessfulStatus returns true if the provider has a cached successful status.
func (s *Service) hasSuccessfulStatus(name string) bool {

	s.mu.RLock()
	defer s.mu.RUnlock()
	status, ok := s.status[name]
	return ok && status.OK
}

// SetStatus records a provider status check result.
func (s *Service) SetStatus(name string, ok bool, message string) {

	s.mu.Lock()
	defer s.mu.Unlock()
	s.status[name] = Status{
		OK:        ok,
		Message:   message,
		CheckedAt: time.Now().UnixMilli(),
	}
}

// ClearStatus removes any stored status for a provider.
func (s *Service) ClearStatus(name string) {

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.status, name)
}

// GetActiveProvider returns the currently active provider.
func (s *Service) GetActiveProvider() Provider {

	if s.registry == nil {
		return nil
	}
	return s.registry.GetActive()
}

// GetProvider returns a provider by name.
func (s *Service) GetProvider(name string) Provider {

	if s.registry == nil {
		return nil
	}
	return s.registry.Get(name)
}

func (s *Service) logDebug(message string, fields ...LogField) {

	if s.logger == nil {
		return
	}
	s.logger.Debug(message, fields...)
}

func (s *Service) logInfo(message string, fields ...LogField) {

	if s.logger == nil {
		return
	}
	s.logger.Info(message, fields...)
}

func (s *Service) logWarn(message string, err error, fields ...LogField) {

	if s.logger == nil {
		return
	}
	s.logger.Warn(message, err, fields...)
}

func (s *Service) logError(message string, err error, fields ...LogField) {

	if s.logger == nil {
		return
	}
	s.logger.Error(message, err, fields...)
}
