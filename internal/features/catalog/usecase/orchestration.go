// orchestration.go exposes catalog workflows for the UI.
// internal/features/catalog/usecase/orchestration.go
package catalog

import (
	"context"
	"fmt"

	coreports "github.com/MadeByDoug/wls-chatbot/internal/core/ports"
)

// Orchestrator coordinates catalog interactions.
type Orchestrator struct {
	service *Service
	emitter coreports.Emitter
}

// NewOrchestrator creates a catalog orchestrator.
func NewOrchestrator(service *Service, emitter coreports.Emitter) *Orchestrator {

	return &Orchestrator{service: service, emitter: emitter}
}

// RefreshAll refreshes models for all endpoints.
func (o *Orchestrator) RefreshAll(ctx context.Context) error {

	if err := o.requireService(); err != nil {
		return err
	}
	return o.service.RefreshAll(ctx)
}

// GetOverview returns the catalog overview.
func (o *Orchestrator) GetOverview(ctx context.Context) (CatalogOverview, error) {

	if err := o.requireService(); err != nil {
		return CatalogOverview{}, err
	}
	return o.service.GetOverview(ctx)
}

// RefreshEndpoint refreshes models for a single endpoint.
func (o *Orchestrator) RefreshEndpoint(ctx context.Context, endpointID string) error {

	if err := o.requireService(); err != nil {
		return err
	}
	err := o.service.RefreshEndpoint(ctx, endpointID)
	if err == nil {
		o.emitCatalogUpdated()
	}
	return err
}

// TestEndpoint tests connectivity for a single endpoint.
func (o *Orchestrator) TestEndpoint(ctx context.Context, endpointID string) error {

	if err := o.requireService(); err != nil {
		return err
	}
	err := o.service.TestEndpoint(ctx, endpointID)
	if err == nil {
		o.emitCatalogUpdated()
	}
	return err
}

// SaveRole creates or updates a role.
func (o *Orchestrator) SaveRole(ctx context.Context, role RoleSummary) (RoleSummary, error) {

	if err := o.requireService(); err != nil {
		return RoleSummary{}, err
	}
	saved, err := o.service.SaveRole(ctx, role)
	if err == nil {
		o.emitCatalogUpdated()
	}
	return saved, err
}

// DeleteRole removes a role.
func (o *Orchestrator) DeleteRole(ctx context.Context, roleID string) error {

	if err := o.requireService(); err != nil {
		return err
	}
	err := o.service.DeleteRole(ctx, roleID)
	if err == nil {
		o.emitCatalogUpdated()
	}
	return err
}

// AssignRole assigns a model to a role.
func (o *Orchestrator) AssignRole(ctx context.Context, roleID, modelEntryID, assignedBy string) (RoleAssignmentResult, error) {

	if err := o.requireService(); err != nil {
		return RoleAssignmentResult{}, err
	}
	result, err := o.service.AssignRole(ctx, roleID, modelEntryID, assignedBy)
	if err == nil {
		o.emitCatalogUpdated()
	}
	return result, err
}

// UnassignRole removes a role assignment.
func (o *Orchestrator) UnassignRole(ctx context.Context, roleID, modelEntryID string) error {

	if err := o.requireService(); err != nil {
		return err
	}
	err := o.service.UnassignRole(ctx, roleID, modelEntryID)
	if err == nil {
		o.emitCatalogUpdated()
	}
	return err
}

func (o *Orchestrator) emitCatalogUpdated() {

	if o.emitter == nil {
		return
	}
	o.emitter.EmitCatalogUpdated()
}

// requireService validates orchestrator dependencies before use.
func (o *Orchestrator) requireService() error {

	if o == nil || o.service == nil {
		return fmt.Errorf("catalog service not configured")
	}
	return nil
}
