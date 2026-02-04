/**
 * hold provider-related UI state in signals.
 */

import { signal } from '@preact/signals-core';
import type { ProviderInfo } from '../types/wails';

export const providers = signal<ProviderInfo[]>([]);
export const providerBusy = signal<Record<string, boolean>>({});
export const providerError = signal<string | null>(null);

/**
 * replace the provider list state.
 */
export function setProviders(list: ProviderInfo[]): void {
    providers.value = list;
}

/**
 * set the busy state for a provider action.
 */
export function setProviderBusy(name: string, busy: boolean): void {
    providerBusy.value = { ...providerBusy.value, [name]: busy };
}

/**
 * set the provider error state.
 */
export function setProviderError(error: string | null): void {
    providerError.value = error;
}
