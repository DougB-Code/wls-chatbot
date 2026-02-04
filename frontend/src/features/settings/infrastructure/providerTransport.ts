/**
 * perform provider-related backend calls.
 * frontend/src/features/settings/infrastructure/providerTransport.ts
 */

import type { ProviderInfo } from '../../../types/wails';
import {
    GetProviders,
    ConnectProvider,
    DisconnectProvider,
    ConfigureProvider,
    SetActiveProvider,
    TestProvider,
    RefreshProviderResources,
    GetActiveProvider,
} from '../../../../wailsjs/go/wails/Bridge';

/**
 * fetch providers from the backend.
 */
export async function getProviders(): Promise<ProviderInfo[]> {
    return GetProviders();
}

/**
 * connect a provider with the given API key.
 */
export async function connectProvider(name: string, apiKey: string): Promise<ProviderInfo> {
    return ConnectProvider(name, apiKey);
}

/**
 * disconnect a provider and remove its credentials.
 */
export async function disconnectProvider(name: string): Promise<void> {
    await DisconnectProvider(name);
}

/**
 * update provider credentials without changing active state.
 */
export async function configureProvider(name: string, apiKey: string): Promise<void> {
    await ConfigureProvider(name, apiKey);
}

/**
 * set the active provider in the backend.
 */
export async function setActiveProvider(name: string): Promise<boolean> {
    return SetActiveProvider(name);
}

/**
 * trigger a provider connectivity test.
 */
export async function testProvider(name: string): Promise<void> {
    await TestProvider(name);
}

/**
 * refresh provider resources/models from the backend.
 */
export async function refreshProviderResources(name: string): Promise<void> {
    await RefreshProviderResources(name);
}

/**
 * fetch the active provider from the backend.
 */
export async function getActiveProvider(): Promise<ProviderInfo | null> {
    return GetActiveProvider();
}
