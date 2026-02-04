/**
 * derive chat view models from chat and provider state.
 */

import { computed } from '@preact/signals-core';
import { activeConversation } from '../store/signals';
import { preferredModelId } from '../store/chatPreferences';
import { providerConfig } from './providerSelectors';
import { resolveEffectiveModelId } from './modelSelection';

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
