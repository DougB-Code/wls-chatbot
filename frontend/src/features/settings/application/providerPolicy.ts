/**
 * orchestrate provider actions and update provider state.
 * frontend/src/features/settings/application/providerPolicy.ts
 */

import * as providerTransport from '../infrastructure/providerTransport';
import { onEvent, offEvent } from '../../../core/infrastructure/wails/events';
import { setProviders, setProviderBusy, setProviderError } from '../state/providerSignals';

let providerEventsInitialized = false;

/**
 * attach provider-related backend event listeners.
 */
export function initProviderEvents(): void {
    if (providerEventsInitialized) return;
    providerEventsInitialized = true;
    onEvent('providers:updated', () => {
        void refreshProviders();
    });
}

/**
 * detach provider-related backend event listeners.
 */
export function teardownProviderEvents(): void {
    if (!providerEventsInitialized) return;
    providerEventsInitialized = false;
    offEvent('providers:updated');
}

/**
 * refresh provider state from the backend.
 */
export async function refreshProviders(): Promise<void> {
    setProviderError(null);
    try {
        const list = await providerTransport.getProviders();
        setProviders(list);
    } catch (err) {
        setProviderError((err as Error).message || 'Failed to load providers');
    }
}

/**
 * connect a provider and refresh provider state.
 */
export async function connectProvider(name: string, apiKey: string): Promise<void> {
    setProviderBusy(name, true);
    setProviderError(null);
    try {
        await providerTransport.connectProvider(name, apiKey);
        await refreshProviders();
    } catch (err) {
        setProviderError((err as Error).message || `Failed to connect ${name}`);
        throw err;
    } finally {
        setProviderBusy(name, false);
    }
}

/**
 * update provider credentials and refresh provider state.
 */
export async function configureProvider(name: string, apiKey: string): Promise<void> {
    setProviderBusy(name, true);
    setProviderError(null);
    try {
        await providerTransport.configureProvider(name, apiKey);
        await refreshProviders();
    } catch (err) {
        setProviderError((err as Error).message || `Failed to update ${name}`);
        throw err;
    } finally {
        setProviderBusy(name, false);
    }
}

/**
 * switch the active provider and refresh provider state.
 */
export async function setActiveProvider(name: string): Promise<void> {
    setProviderBusy(name, true);
    setProviderError(null);
    try {
        const changed = await providerTransport.setActiveProvider(name);
        if (!changed) {
            throw new Error(`Failed to set active provider: ${name}`);
        }
        await refreshProviders();
    } catch (err) {
        setProviderError((err as Error).message || `Failed to set active provider: ${name}`);
        throw err;
    } finally {
        setProviderBusy(name, false);
    }
}

/**
 * disconnect a provider and refresh provider state.
 */
export async function disconnectProvider(name: string): Promise<void> {
    setProviderBusy(name, true);
    setProviderError(null);
    try {
        await providerTransport.disconnectProvider(name);
        await refreshProviders();
    } catch (err) {
        setProviderError((err as Error).message || `Failed to disconnect ${name}`);
        throw err;
    } finally {
        setProviderBusy(name, false);
    }
}

/**
 * refresh provider resources and provider state.
 */
export async function refreshProviderResources(name: string): Promise<void> {
    setProviderBusy(name, true);
    setProviderError(null);
    try {
        await providerTransport.refreshProviderResources(name);
        await refreshProviders();
    } catch (err) {
        setProviderError((err as Error).message || `Failed to refresh ${name}`);
        throw err;
    } finally {
        setProviderBusy(name, false);
    }
}
