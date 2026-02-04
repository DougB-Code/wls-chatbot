/**
 * derive chat view models from chat and provider state.
 * frontend/src/features/chat/application/chatSelectors.ts
 */

import { computed } from '@preact/signals-core';
import { activeConversation } from '../state/chatSignals';
import { preferredModelId } from '../state/chatPreferences';
import { providerConfig } from '../../settings/application/providerSelectors';
import { resolveEffectiveModelId } from '../../settings/domain/modelSelection';

export const availableModels = computed(() => providerConfig.value?.models ?? []);

export const effectiveModelId = computed(() => {
    const provider = providerConfig.value;
    if (!provider) return '';
    const conversationModel = activeConversation.value?.settings?.model ?? '';
    const preferred = preferredModelId.value;
    return resolveEffectiveModelId({
        models: provider.models ?? [],
        defaultModel: provider.defaultModel ?? '',
        conversationModel,
        preferredModel: preferred,
    });
});
