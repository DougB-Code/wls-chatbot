/**
 * catalog.ts defines frontend types for model catalog data.
 * frontend/src/types/catalog.ts
 */

export interface ProviderSummary {
    id: string;
    name: string;
    displayName: string;
    adapterType: string;
    trustMode: string;
    baseUrl: string;
    lastTestAt: number;
    lastTestOk: boolean;
    lastError?: string;
    lastDiscoveryAt: number;
}

export interface ModelSummary {
    id: string;
    endpointId: string;
    modelId: string;
    displayName: string;
    availabilityState: string;
    supportsStreaming: boolean;
    supportsToolCalling: boolean;
    supportsStructuredOutput: boolean;
    supportsVision: boolean;
    inputModalities: string[];
    outputModalities: string[];
}

export interface EndpointSummary {
    id: string;
    providerId: string;
    providerName: string;
    providerDisplayName: string;
    displayName: string;
    adapterType: string;
    baseUrl: string;
    routeKind: string;
    originProvider: string;
    originRouteLabel: string;
    lastTestAt: number;
    lastTestOk: boolean;
    lastError?: string;
    models: ModelSummary[];
}

export interface RoleRequirements {
    requiredInputModalities: string[];
    requiredOutputModalities: string[];
    requiresStreaming: boolean;
    requiresToolCalling: boolean;
    requiresStructuredOutput: boolean;
    requiresVision: boolean;
}

export interface RoleConstraints {
    maxCostTier?: string;
    maxLatencyTier?: string;
    minReliabilityTier?: string;
}

export interface RoleAssignmentSummary {
    roleId: string;
    modelCatalogEntryId: string;
    modelLabel: string;
    assignedBy: string;
    createdAt: number;
    enabled: boolean;
}

export interface RoleSummary {
    id: string;
    name: string;
    requirements: RoleRequirements;
    constraints: RoleConstraints;
    assignments: RoleAssignmentSummary[];
}

export interface CatalogOverview {
    providers: ProviderSummary[];
    endpoints: EndpointSummary[];
    roles: RoleSummary[];
}

export interface RoleAssignmentResult {
    missingModalities?: string[];
    missingFeatures?: string[];
}
