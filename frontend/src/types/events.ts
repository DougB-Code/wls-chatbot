/**
 * define backend-to-frontend event payload shapes.
 * frontend/src/types/events.ts
 */

import type { Message, MessageMetadata } from './wails';

/**
 * enumerate supported chat event types.
 */
export type ChatEventType =
    | 'chat.message'
    | 'chat.stream.start'
    | 'chat.stream.chunk'
    | 'chat.stream.error'
    | 'chat.stream.complete'
    | 'chat.conversation.title';

/**
 * describe the payload for streaming content chunks.
 */
export interface StreamChunkPayload {
    blockIndex: number;
    content: string;
    isDone: boolean;
    metadata?: MessageMetadata;
    error?: string;
    statusCode?: number;
}

/**
 * describe a conversation title update payload.
 */
export interface ConversationTitlePayload {
    title: string;
}

/**
 * wrap chat events with metadata and typed payloads.
 */
export interface ChatEvent<TPayload = Message | StreamChunkPayload | ConversationTitlePayload | Record<string, unknown>> {
    type: ChatEventType;
    conversationId: string;
    messageId?: string;
    ts: number;
    payload?: TPayload;
}
