/**
 * chatPolicy.ts coordinates chat policy actions and chat state updates.
 * frontend/src/features/chat/application/chatPolicy.ts
 */

import type { Conversation } from '../../../types';
import * as chatTransport from '../infrastructure/chatTransport';
import * as store from '../state/chatSignals';
import { orchestrateSendMessage } from './chatFlow';
import { initChatEvents } from '../infrastructure/chatEvents';
import { activeProvider } from '../../settings/application/providerSelectors';
import { preferredModelId } from '../state/chatPreferences';
import { isModelAvailable, pickDefaultModel, resolveProviderModels } from '../../settings/domain/modelSelection';

let chatInitialized = false;

/**
 * initialize chat event handling and hydrate active conversation state.
 */
export async function initChatPolicy(): Promise<void> {
    if (chatInitialized) return;
    chatInitialized = true;
    initChatEvents();

    try {
        const summaries = await chatTransport.listConversations();
        const deletedSummaries = await chatTransport.listDeletedConversations();
        const allSummaries = [...summaries, ...deletedSummaries];

        for (const summary of allSummaries) {
            const conversation = await chatTransport.getConversation(summary.id);
            if (conversation) {
                store.upsertConversation(conversation, false);
            }
        }

        const activeConversation = await chatTransport.getActiveConversation();
        if (activeConversation) {
            store.upsertConversation(activeConversation, true);
        }
    } catch (err) {
        // Backend might not be available yet; keep UI responsive.
        console.warn('Failed to load active conversation:', err);
    }
}

/**
 * orchestrate sending a message through the chat flow.
 */
export async function sendMessage(
    content: string,
    attachments: File[] = [],
    selectedModel?: string,
    currentConversation?: Conversation | null
): Promise<string> {
    const client = {
        init: initChatPolicy,
        createConversation: chatTransport.createConversation,
        setActiveConversation: chatTransport.setActiveConversation,
        getConversation: chatTransport.getConversation,
        updateConversationModel: chatTransport.updateConversationModel,
        sendMessage: chatTransport.sendMessage,
    };

    const modelOverride = selectedModel || preferredModelId.value || undefined;

    return orchestrateSendMessage(
        client,
        content,
        activeProvider.value,
        currentConversation ?? store.activeConversation.value ?? null,
        modelOverride,
        attachments
    );
}

/**
 * create a new empty conversation and mark it active.
 */
export async function createNewConversation(): Promise<Conversation> {
    await initChatPolicy();

    const provider = activeProvider.value;
    if (!provider) {
        throw new Error('No active provider available');
    }

    const models = resolveProviderModels(provider);
    if (models.length === 0) {
        throw new Error(`No models available for provider: ${provider.displayName}`);
    }

    const preferredModel = preferredModelId.value?.trim() ?? '';
    const model = isModelAvailable(models, preferredModel)
        ? preferredModel
        : pickDefaultModel(models);

    const conversation = await chatTransport.createConversation(provider.name, model);
    store.upsertConversation(conversation, true);

    return conversation;
}

/**
 * activate a saved conversation by id.
 */
export async function selectConversation(conversationId: string): Promise<void> {
    if (!conversationId.trim()) return;

    await initChatPolicy();
    await chatTransport.setActiveConversation(conversationId);

    const conversation = await chatTransport.getConversation(conversationId);
    if (conversation) {
        store.upsertConversation(conversation, true);
        return;
    }

    store.setActiveConversation(conversationId);
}

/**
 * delete a saved conversation and refresh active selection.
 */
export async function deleteConversation(conversationId: string): Promise<void> {
    const trimmedId = conversationId.trim();
    if (!trimmedId) return;

    await initChatPolicy();

    const deleted = await chatTransport.deleteConversation(trimmedId);
    if (!deleted) {
        throw new Error('Failed to delete conversation');
    }

    const deletedConversation = await chatTransport.getConversation(trimmedId);
    if (deletedConversation) {
        store.upsertConversation(deletedConversation, false);
    } else {
        const existing = store.getConversation(trimmedId);
        if (existing) {
            store.upsertConversation({ ...existing, isArchived: true, updatedAt: Date.now() }, false);
        }
    }

    const wasActive = store.activeConversation.value?.id === trimmedId;
    if (!wasActive) {
        return;
    }

    const nextConversation = [...store.conversations.value.values()]
        .filter((conversation) => !conversation.isArchived && conversation.id !== trimmedId)
        .sort((left, right) => right.updatedAt - left.updatedAt)[0];
    if (nextConversation) {
        await selectConversation(nextConversation.id);
        return;
    }

    store.setActiveConversation(null);
}

/**
 * restore a recycled conversation.
 */
export async function restoreConversation(conversationId: string): Promise<void> {
    const trimmedId = conversationId.trim();
    if (!trimmedId) return;

    await initChatPolicy();

    const restored = await chatTransport.restoreConversation(trimmedId);
    if (!restored) {
        throw new Error('Failed to restore conversation');
    }

    const restoredConversation = await chatTransport.getConversation(trimmedId);
    if (restoredConversation) {
        store.upsertConversation(restoredConversation, false);
    } else {
        const existing = store.getConversation(trimmedId);
        if (existing) {
            store.upsertConversation({ ...existing, isArchived: false, updatedAt: Date.now() }, false);
        }
    }
}

/**
 * permanently remove a conversation.
 */
export async function purgeConversation(conversationId: string): Promise<void> {
    const trimmedId = conversationId.trim();
    if (!trimmedId) return;

    await initChatPolicy();

    const purged = await chatTransport.purgeConversation(trimmedId);
    if (!purged) {
        throw new Error('Failed to permanently delete conversation');
    }

    const wasActive = store.activeConversation.value?.id === trimmedId;
    store.removeConversation(trimmedId);
    if (!wasActive) {
        return;
    }

    const nextConversation = [...store.conversations.value.values()]
        .filter((conversation) => !conversation.isArchived)
        .sort((left, right) => right.updatedAt - left.updatedAt)[0];
    if (nextConversation) {
        await selectConversation(nextConversation.id);
        return;
    }

    store.setActiveConversation(null);
}

/**
 * request the backend to stop streaming output.
 */
export async function stopStream(): Promise<void> {
    await chatTransport.stopStream();
}

/**
 * update the conversation model in the backend and store.
 */
export async function setConversationModel(conversationId: string, model: string): Promise<void> {
    const updated = await chatTransport.updateConversationModel(conversationId, model);
    if (!updated) {
        throw new Error('Failed to update model');
    }

    const conversation = store.getConversation(conversationId);
    if (!conversation) return;

    if (conversation.settings?.model !== model) {
        store.upsertConversation({
            ...conversation,
            settings: {
                ...conversation.settings,
                model,
            },
            updatedAt: Date.now(),
        }, true);
    }
}

/**
 * update the conversation provider in the backend and store.
 */
export async function setConversationProvider(conversationId: string, provider: string): Promise<void> {
    const updated = await chatTransport.updateConversationProvider(conversationId, provider);
    if (!updated) {
        throw new Error('Failed to update provider');
    }

    const conversation = store.getConversation(conversationId);
    if (!conversation) return;

    if (conversation.settings?.provider !== provider) {
        store.upsertConversation({
            ...conversation,
            settings: {
                ...conversation.settings,
                provider,
            },
            updatedAt: Date.now(),
        }, true);
    }
}
