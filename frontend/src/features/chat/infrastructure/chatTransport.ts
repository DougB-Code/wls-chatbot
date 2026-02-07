/**
 * chatTransport.ts performs chat-related backend calls.
 * frontend/src/features/chat/infrastructure/chatTransport.ts
 */

import type { Conversation, ConversationSummary, Message } from '../../../types';
import {
    CreateConversation,
    GetConversation,
    GetActiveConversation,
    ListConversations,
    ListDeletedConversations,
    DeleteConversation,
    RestoreConversation,
    PurgeConversation,
    SetActiveConversation,
    UpdateConversationModel,
    UpdateConversationProvider,
    SendMessage,
    StopStream,
} from '../../../../wailsjs/go/wails/Bridge';

/**
 * create a new conversation in the backend.
 */
export async function createConversation(provider: string, model: string): Promise<Conversation> {
    const conversation = await CreateConversation(provider, model);
    if (!conversation || !conversation.id) {
        throw new Error('Failed to create conversation');
    }
    return conversation;
}

/**
 * fetch a conversation by id from the backend.
 */
export async function getConversation(id: string): Promise<Conversation | null> {
    return GetConversation(id);
}

/**
 * fetch the active conversation from the backend.
 */
export async function getActiveConversation(): Promise<Conversation | null> {
    return GetActiveConversation();
}

/**
 * fetch conversation summaries from the backend.
 */
export async function listConversations(): Promise<ConversationSummary[]> {
    return ListConversations();
}

/**
 * fetch archived conversation summaries from the backend.
 */
export async function listDeletedConversations(): Promise<ConversationSummary[]> {
    return ListDeletedConversations();
}

/**
 * move a conversation to the recycle bin in the backend.
 */
export async function deleteConversation(id: string): Promise<boolean> {
    return DeleteConversation(id);
}

/**
 * restore a recycled conversation in the backend.
 */
export async function restoreConversation(id: string): Promise<boolean> {
    return RestoreConversation(id);
}

/**
 * permanently delete a recycled conversation in the backend.
 */
export async function purgeConversation(id: string): Promise<boolean> {
    return PurgeConversation(id);
}

/**
 * set the active conversation in the backend.
 */
export async function setActiveConversation(id: string): Promise<void> {
    await SetActiveConversation(id);
}

/**
 * update the model setting for a conversation.
 */
export async function updateConversationModel(conversationId: string, model: string): Promise<boolean> {
    return UpdateConversationModel(conversationId, model);
}

/**
 * update the provider setting for a conversation.
 */
export async function updateConversationProvider(conversationId: string, provider: string): Promise<boolean> {
    return UpdateConversationProvider(conversationId, provider);
}

/**
 * send a message to the backend for a conversation.
 */
export async function sendMessage(conversationId: string, content: string): Promise<Message> {
    return SendMessage(conversationId, content);
}

/**
 * request the backend to stop streaming output.
 */
export async function stopStream(): Promise<void> {
    await StopStream();
}
