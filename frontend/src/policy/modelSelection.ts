/**
 * provide shared model selection helpers.
 */

import type { ProviderInfo, ProviderModel } from '../types/wails';

/**
 * describe a minimal model identifier shape.
 */
export type ModelLike = Pick<ProviderModel, 'id'>;

/**
 * resolve the effective model list for a provider.
 */
export function resolveProviderModels(provider: ProviderInfo | null): ProviderModel[] {
    if (!provider) return [];
    return provider.models ?? [];
}

/**
 * check whether a model id exists in the available list.
 */
export function isModelAvailable(models: ModelLike[], modelId: string): boolean {
    const trimmed = modelId.trim();
    if (!trimmed) return false;
    return models.some((model) => model.id === trimmed);
}

/**
 * pick a default model id from a list.
 */
export function pickDefaultModel(models: ModelLike[]): string {
    return models[0]?.id ?? '';
}

/**
 * resolve the most appropriate model id from known candidates.
 */
export function resolveEffectiveModelId(options: {
    models: ModelLike[];
    defaultModel: string;
    conversationModel?: string;
    preferredModel?: string;
    selectedModel?: string;
}): string {
    const modelIds = options.models.map((model) => model.id);

    const pickIfAvailable = (candidate?: string): string | null => {
        const trimmed = candidate?.trim() ?? '';
        if (!trimmed) return null;
        if (modelIds.length === 0) return null;
        return modelIds.includes(trimmed) ? trimmed : null;
    };

    return (
        pickIfAvailable(options.selectedModel) ||
        pickIfAvailable(options.conversationModel) ||
        pickIfAvailable(options.preferredModel) ||
        (modelIds.length > 0 ? options.defaultModel || modelIds[0] : '')
    );
}
