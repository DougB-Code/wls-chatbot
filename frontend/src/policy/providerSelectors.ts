/**
 * derive provider-related view models from provider state.
 */

import { computed } from '@preact/signals-core';
import type { ProviderInfo } from '../types/wails';
import { providers } from '../store/providerSignals';
import { pickDefaultModel, resolveProviderModels } from './modelSelection';

export const activeProvider = computed(() => providers.value.find((p) => p.isActive) ?? null);

export const isConnected = computed(() => activeProvider.value?.isConnected ?? false);

export const providerConfig = computed(() => mapProviderToConfig(activeProvider.value));

/**
 * map backend provider info into UI-facing config.
 */
export function mapProviderToConfig(provider: ProviderInfo | null) {
    if (!provider) return null;

    const providerModels = resolveProviderModels(provider);

    return {
        name: provider.name,
        displayName: provider.displayName,
        models: providerModels.map((model) => ({
            id: model.id,
            name: model.name || model.id,
            contextWindow: model.contextWindow,
            supportsStreaming: model.supportsStreaming,
            supportsTools: model.supportsTools,
            supportsVision: model.supportsVision,
        })),
        defaultModel: pickDefaultModel(providerModels),
        isConnected: provider.isConnected,
    };
}
