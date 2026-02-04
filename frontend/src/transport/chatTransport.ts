/**
 * perform chat-related backend calls.
 */

import type { Conversation, ConversationSummary, Message } from '../types';
import {
    CreateConversation,
    GetConversation,
    GetActiveConversation,
    ListConversations,
    SetActiveConversation,
    UpdateConversationModel,
    SendMessage,
    StopStream,
} from '../../wailsjs/go/wails/Bridge';

/**
 * create a new conversation in the backend.
 */
export async function createConversation(provider: string, model: string): Promise<Conversation> {
    return CreateConversation(provider, model);
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
