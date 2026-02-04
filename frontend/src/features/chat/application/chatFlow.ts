/**
 * orchestrate the send-message flow outside the service layer.
 * frontend/src/features/chat/application/chatFlow.ts
 */

import type { Conversation, Message } from '../../../types';
import type { ProviderInfo } from '../../../types/wails';
import * as store from '../state/chatSignals';
import { log } from '../../../lib/logger';
import { isModelAvailable, pickDefaultModel, resolveProviderModels } from '../../settings/domain/modelSelection';

/**
 * define the client contract used by the chat flow orchestrator.
 */
export interface ChatClient {
    init(): Promise<void>;
    createConversation(provider: string, model: string): Promise<Conversation>;
    setActiveConversation(conversationId: string): Promise<void>;
    getConversation(id: string): Promise<Conversation | null>;
    updateConversationModel(conversationId: string, model: string): Promise<boolean>;
    sendMessage(conversationId: string, content: string): Promise<Message>;
}

/**
 * orchestrate conversation setup, model selection, and message send.
 */
export async function orchestrateSendMessage(
    client: ChatClient,
    content: string,
    activeProvider: ProviderInfo | null,
    currentConversation: Conversation | null,
    selectedModel?: string,
    attachments: File[] = []
): Promise<string> {
    await client.init();
    let conversationId = currentConversation?.id;

    if (!conversationId) {
        if (!activeProvider) {
            log.error().msg('No active provider when creating conversation');
            throw new Error('No active provider');
        }
        const availableModels = resolveProviderModels(activeProvider);
        if (availableModels.length === 0) {
            throw new Error(`No models available for provider: ${activeProvider.displayName}`);
        }
        let model = selectedModel?.trim() || '';
        if (!isModelAvailable(availableModels, model)) {
            model = pickDefaultModel(availableModels);
        }
        log.info()
            .str('provider', activeProvider.name)
            .str('model', model)
            .msg('Creating new conversation');

        const newConv = await client.createConversation(activeProvider.name, model);
        conversationId = newConv.id;
        log.info().str('conversationId', conversationId).msg('Conversation created');
        store.upsertConversation(newConv, true);
    } else {
        store.setActiveConversation(conversationId);
        await client.setActiveConversation(conversationId);
    }

    if (!conversationId) throw new Error('Failed to resolve conversation ID');

    if (!store.hasConversation(conversationId)) {
        const conv = await client.getConversation(conversationId);
        if (conv) {
            store.upsertConversation(conv, true);
        }
    }

    await ensureConversationModel(
        client,
        store.getConversation(conversationId) ?? currentConversation,
        activeProvider
    );

    const convForModel = store.getConversation(conversationId) ?? currentConversation;
    if (convForModel) {
        const desiredModel = resolveDesiredModel(convForModel, activeProvider);
        const updated = await client.updateConversationModel(conversationId, desiredModel);
        if (!updated) {
            throw new Error(`Failed to update model to ${desiredModel}`);
        }
        if (convForModel.settings?.model !== desiredModel) {
            store.upsertConversation({
                ...convForModel,
                settings: {
                    ...convForModel.settings,
                    model: desiredModel,
                },
                updatedAt: Date.now(),
            }, true);
        }
    }

    const formattedContent = appendAttachments(content, attachments);
    await client.sendMessage(conversationId, formattedContent);

    return conversationId;
}

/**
 * choose a valid model for the conversation and provider.
 */
function resolveDesiredModel(conversation: Conversation, provider: ProviderInfo | null): string {
    const currentModel = conversation.settings?.model?.trim() ?? '';
    const availableModels = resolveProviderModels(provider);

    if (currentModel === '') {
        if (availableModels.length === 0) {
            throw new Error(`No models available for provider: ${provider?.displayName ?? 'unknown'}`);
        }
        return pickDefaultModel(availableModels);
    }

    if (availableModels.length > 0 && !isModelAvailable(availableModels, currentModel)) {
        throw new Error(`Selected model is not available for provider: ${currentModel}`);
    }

    return currentModel;
}

/**
 * ensure the conversation has a model set in the backend/store.
 */
async function ensureConversationModel(
    client: ChatClient,
    conversation: Conversation | null,
    provider: ProviderInfo | null
): Promise<void> {
    if (!conversation || !provider) return;
    const availableModels = resolveProviderModels(provider);
    if (availableModels.length === 0) return;

    const currentModel = conversation.settings?.model;
    if (currentModel) return;

    const desiredModel = pickDefaultModel(availableModels);
    const updated = await client.updateConversationModel(conversation.id, desiredModel);
    if (updated) {
        store.upsertConversation({
            ...conversation,
            settings: {
                ...conversation.settings,
                model: desiredModel,
            },
            updatedAt: Date.now(),
        }, true);
    }
}

/**
 * append attachment metadata into the outgoing message body.
 */
function appendAttachments(content: string, attachments: File[]): string {
    if (!attachments.length) return content;
    const lines = attachments.map((file) => {
        const size = formatBytes(file.size);
        const type = file.type || 'unknown';
        return `- ${file.name} (${type}, ${size})`;
    });
    const header = `Attachments:\n${lines.join('\n')}\n\n`;
    return `${header}${content}`.trim();
}

/**
 * format byte sizes into human-readable strings.
 */
function formatBytes(size: number): string {
    if (!size) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    let unitIndex = 0;
    let value = size;
    while (value >= 1024 && unitIndex < units.length - 1) {
        value /= 1024;
        unitIndex += 1;
    }
    return `${value.toFixed(value >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}
