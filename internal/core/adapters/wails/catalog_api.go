// expose model catalog and role endpoints to the frontend.
// internal/core/adapters/wails/catalog_api.go
package wails

import "github.com/MadeByDoug/wls-chatbot/internal/features/catalog/usecase"

// GetCatalogOverview returns the model catalog overview.
func (b *Bridge) GetCatalogOverview() (catalog.CatalogOverview, error) {

	return b.catalog.GetOverview(b.ctxOrBackground())
}

// RefreshCatalogEndpoint refreshes models for a specific endpoint.
func (b *Bridge) RefreshCatalogEndpoint(endpointID string) error {

	return b.catalog.RefreshEndpoint(b.ctxOrBackground(), endpointID)
}

// TestCatalogEndpoint tests connectivity for an endpoint.
func (b *Bridge) TestCatalogEndpoint(endpointID string) error {

	return b.catalog.TestEndpoint(b.ctxOrBackground(), endpointID)
}

// SaveCatalogRole creates or updates a role.
func (b *Bridge) SaveCatalogRole(role catalog.RoleSummary) (catalog.RoleSummary, error) {

	return b.catalog.SaveRole(b.ctxOrBackground(), role)
}

// DeleteCatalogRole removes a role.
func (b *Bridge) DeleteCatalogRole(roleID string) error {

	return b.catalog.DeleteRole(b.ctxOrBackground(), roleID)
}

// AssignCatalogRole assigns a model to a role.
func (b *Bridge) AssignCatalogRole(roleID, modelEntryID, assignedBy string) (catalog.RoleAssignmentResult, error) {

	return b.catalog.AssignRole(b.ctxOrBackground(), roleID, modelEntryID, assignedBy)
}

// UnassignCatalogRole removes a role assignment.
func (b *Bridge) UnassignCatalogRole(roleID, modelEntryID string) error {

	return b.catalog.UnassignRole(b.ctxOrBackground(), roleID, modelEntryID)
}
