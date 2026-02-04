/**
 * coordinate chat policy actions and chat state updates.
 */

import type { Conversation } from '../types';
import * as chatTransport from '../transport/chatTransport';
import * as store from '../store/signals';
import { orchestrateSendMessage } from './chatFlow';
import { initChatEvents } from './chatEvents';
import { activeProvider } from './providerSelectors';
import { preferredModelId } from '../store/chatPreferences';

let chatInitialized = false;

/**
 * initialize chat event handling and hydrate active conversation state.
 */
export async function initChatPolicy(): Promise<void> {
    if (chatInitialized) return;
    chatInitialized = true;
    initChatEvents();

    try {
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
