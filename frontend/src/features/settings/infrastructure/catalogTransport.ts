/**
 * catalogTransport.ts performs catalog-related backend calls.
 * frontend/src/features/settings/infrastructure/catalogTransport.ts
 */

import type { CatalogOverview, RoleAssignmentResult, RoleSummary } from '../../../types/catalog';
import {
    GetCatalogOverview,
    RefreshCatalogEndpoint,
    TestCatalogEndpoint,
    SaveCatalogRole,
    DeleteCatalogRole,
    AssignCatalogRole,
    UnassignCatalogRole,
} from '../../../../wailsjs/go/wails/Bridge';

/**
 * fetch the catalog overview from the backend.
 */
export async function getCatalogOverview(): Promise<CatalogOverview> {
    return GetCatalogOverview();
}

/**
 * refresh models for a catalog endpoint.
 */
export async function refreshCatalogEndpoint(endpointId: string): Promise<void> {
    await RefreshCatalogEndpoint(endpointId);
}

/**
 * test connectivity for a catalog endpoint.
 */
export async function testCatalogEndpoint(endpointId: string): Promise<void> {
    await TestCatalogEndpoint(endpointId);
}

/**
 * create or update a catalog role.
 */
export async function saveCatalogRole(role: RoleSummary): Promise<RoleSummary> {
    return SaveCatalogRole(role);
}

/**
 * delete a catalog role.
 */
export async function deleteCatalogRole(roleId: string): Promise<void> {
    await DeleteCatalogRole(roleId);
}

/**
 * assign a model to a role.
 */
export async function assignCatalogRole(roleId: string, modelEntryId: string, assignedBy: string): Promise<RoleAssignmentResult> {
    return AssignCatalogRole(roleId, modelEntryId, assignedBy);
}

/**
 * remove a role assignment.
 */
export async function unassignCatalogRole(roleId: string, modelEntryId: string): Promise<void> {
    await UnassignCatalogRole(roleId, modelEntryId);
}
