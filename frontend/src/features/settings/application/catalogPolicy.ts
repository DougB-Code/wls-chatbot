/**
 * catalogPolicy.ts orchestrates catalog actions and state updates.
 * frontend/src/features/settings/application/catalogPolicy.ts
 */

import * as catalogTransport from '../infrastructure/catalogTransport';
import { onEvent, offEvent } from '../../../core/infrastructure/wails/events';
import { setCatalogOverview, setCatalogBusy, setCatalogError } from '../state/catalogSignals';
import type { RoleSummary } from '../../../types/catalog';

let catalogEventsInitialized = false;

/**
 * attach catalog-related backend event listeners.
 */
export function initCatalogEvents(): void {
    if (catalogEventsInitialized) return;
    catalogEventsInitialized = true;
    onEvent('catalog:updated', () => {
        void refreshCatalogOverview();
    });
}

/**
 * detach catalog-related backend event listeners.
 */
export function teardownCatalogEvents(): void {
    if (!catalogEventsInitialized) return;
    catalogEventsInitialized = false;
    offEvent('catalog:updated');
}

/**
 * refresh catalog overview from the backend.
 */
export async function refreshCatalogOverview(): Promise<void> {
    setCatalogError(null);
    try {
        const overview = await catalogTransport.getCatalogOverview();
        setCatalogOverview(overview);
    } catch (err) {
        setCatalogError((err as Error).message || 'Failed to load catalog');
    }
}

/**
 * refresh models for a catalog endpoint.
 */
export async function refreshCatalogEndpoint(endpointId: string): Promise<void> {
    setCatalogBusy(endpointId, true);
    setCatalogError(null);
    try {
        await catalogTransport.refreshCatalogEndpoint(endpointId);
        await refreshCatalogOverview();
    } catch (err) {
        setCatalogError((err as Error).message || 'Failed to refresh endpoint');
        throw err;
    } finally {
        setCatalogBusy(endpointId, false);
    }
}

/**
 * test connectivity for a catalog endpoint.
 */
export async function testCatalogEndpoint(endpointId: string): Promise<void> {
    setCatalogBusy(endpointId, true);
    setCatalogError(null);
    try {
        await catalogTransport.testCatalogEndpoint(endpointId);
        await refreshCatalogOverview();
    } catch (err) {
        setCatalogError((err as Error).message || 'Failed to test endpoint');
        throw err;
    } finally {
        setCatalogBusy(endpointId, false);
    }
}

/**
 * save a role definition.
 */
export async function saveCatalogRole(role: RoleSummary): Promise<RoleSummary> {
    setCatalogError(null);
    const key = role.id ? `role-${role.id}` : 'role-new';
    setCatalogBusy(key, true);
    try {
        const saved = await catalogTransport.saveCatalogRole(role);
        await refreshCatalogOverview();
        return saved;
    } catch (err) {
        setCatalogError((err as Error).message || 'Failed to save role');
        throw err;
    } finally {
        setCatalogBusy(key, false);
    }
}

/**
 * delete a role.
 */
export async function deleteCatalogRole(roleId: string): Promise<void> {
    setCatalogBusy(`role-${roleId}`, true);
    setCatalogError(null);
    try {
        await catalogTransport.deleteCatalogRole(roleId);
        await refreshCatalogOverview();
    } catch (err) {
        setCatalogError((err as Error).message || 'Failed to delete role');
        throw err;
    } finally {
        setCatalogBusy(`role-${roleId}`, false);
    }
}

/**
 * assign a model to a role.
 */
export async function assignCatalogRole(roleId: string, modelEntryId: string, assignedBy: string): Promise<void> {
    setCatalogBusy(`role-${roleId}`, true);
    setCatalogError(null);
    try {
        await catalogTransport.assignCatalogRole(roleId, modelEntryId, assignedBy);
        await refreshCatalogOverview();
    } catch (err) {
        setCatalogError((err as Error).message || 'Failed to assign role');
        throw err;
    } finally {
        setCatalogBusy(`role-${roleId}`, false);
    }
}

/**
 * remove a role assignment.
 */
export async function unassignCatalogRole(roleId: string, modelEntryId: string): Promise<void> {
    setCatalogBusy(`role-${roleId}`, true);
    setCatalogError(null);
    try {
        await catalogTransport.unassignCatalogRole(roleId, modelEntryId);
        await refreshCatalogOverview();
    } catch (err) {
        setCatalogError((err as Error).message || 'Failed to remove assignment');
        throw err;
    } finally {
        setCatalogBusy(`role-${roleId}`, false);
    }
}
