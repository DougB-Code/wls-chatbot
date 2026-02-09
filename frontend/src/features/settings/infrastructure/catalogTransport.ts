/**
 * catalogTransport.ts persists catalog role state in frontend local storage.
 * frontend/src/features/settings/infrastructure/catalogTransport.ts
 */

import type { CatalogOverview, RoleAssignmentResult, RoleSummary } from '../../../types/catalog';

const storageKey = 'wls.catalog.roles';

/**
 * build an empty catalog overview shape.
 */
function emptyOverview(): CatalogOverview {
    return {
        providers: [],
        endpoints: [],
        roles: [],
    };
}

/**
 * read stored catalog roles from local storage.
 */
function readRoles(): RoleSummary[] {
    if (typeof globalThis.localStorage === 'undefined') {
        return [];
    }

    try {
        const raw = globalThis.localStorage.getItem(storageKey);
        if (!raw) {
            return [];
        }
        const parsed = JSON.parse(raw) as RoleSummary[];
        if (!Array.isArray(parsed)) {
            return [];
        }
        return parsed;
    } catch {
        return [];
    }
}

/**
 * write catalog roles to local storage.
 */
function writeRoles(roles: RoleSummary[]): void {
    if (typeof globalThis.localStorage === 'undefined') {
        return;
    }

    globalThis.localStorage.setItem(storageKey, JSON.stringify(roles));
}

/**
 * normalize role payload shape for persistence.
 */
function normalizeRole(role: RoleSummary): RoleSummary {
    return {
        id: role.id,
        name: role.name,
        requirements: {
            requiredInputModalities: [...(role.requirements?.requiredInputModalities ?? [])],
            requiredOutputModalities: [...(role.requirements?.requiredOutputModalities ?? [])],
            requiresStreaming: !!role.requirements?.requiresStreaming,
            requiresToolCalling: !!role.requirements?.requiresToolCalling,
            requiresStructuredOutput: !!role.requirements?.requiresStructuredOutput,
            requiresVision: !!role.requirements?.requiresVision,
        },
        constraints: {
            maxCostTier: role.constraints?.maxCostTier,
            maxLatencyTier: role.constraints?.maxLatencyTier,
            minReliabilityTier: role.constraints?.minReliabilityTier,
        },
        assignments: [...(role.assignments ?? [])],
    };
}

/**
 * fetch the catalog overview from local storage.
 */
export async function getCatalogOverview(): Promise<CatalogOverview> {
    return {
        ...emptyOverview(),
        roles: readRoles(),
    };
}

/**
 * refresh models for a catalog endpoint.
 */
export async function refreshCatalogEndpoint(_endpointId: string): Promise<void> {
    return;
}

/**
 * test connectivity for a catalog endpoint.
 */
export async function testCatalogEndpoint(_endpointId: string): Promise<void> {
    return;
}

/**
 * create or update a catalog role.
 */
export async function saveCatalogRole(role: RoleSummary): Promise<RoleSummary> {
    const roles = readRoles();
    const trimmedName = role.name?.trim() ?? '';
    if (!trimmedName) {
        throw new Error('Role name is required');
    }

    const normalized: RoleSummary = normalizeRole({
        ...role,
        id: role.id?.trim() || `role-${Date.now()}`,
        name: trimmedName,
    });

    const next = roles.filter((item) => item.id !== normalized.id);
    next.push(normalized);
    writeRoles(next);
    return normalized;
}

/**
 * delete a catalog role.
 */
export async function deleteCatalogRole(roleId: string): Promise<void> {
    const roles = readRoles();
    writeRoles(roles.filter((item) => item.id !== roleId));
}

/**
 * assign a model to a role.
 */
export async function assignCatalogRole(roleId: string, modelEntryId: string, assignedBy: string): Promise<RoleAssignmentResult> {
    const roles = readRoles();
    const index = roles.findIndex((item) => item.id === roleId);
    if (index < 0) {
        throw new Error(`Role not found: ${roleId}`);
    }

    const role = normalizeRole(roles[index]);
    const exists = role.assignments.some((assignment) => assignment.modelCatalogEntryId === modelEntryId);
    if (!exists) {
        role.assignments.push({
            roleId,
            modelCatalogEntryId: modelEntryId,
            modelLabel: modelEntryId,
            assignedBy: assignedBy.trim() || 'system',
            createdAt: Date.now(),
            enabled: true,
        });
    }

    roles[index] = role;
    writeRoles(roles);

    return {};
}

/**
 * remove a role assignment.
 */
export async function unassignCatalogRole(roleId: string, modelEntryId: string): Promise<void> {
    const roles = readRoles();
    const index = roles.findIndex((item) => item.id === roleId);
    if (index < 0) {
        return;
    }

    const role = normalizeRole(roles[index]);
    role.assignments = role.assignments.filter((assignment) => assignment.modelCatalogEntryId !== modelEntryId);
    roles[index] = role;
    writeRoles(roles);
}
