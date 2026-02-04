/**
 * chatEvents.ts handles backend chat events and updates chat state.
 * frontend/src/features/chat/infrastructure/chatEvents.ts
 */

import type { ChatEvent, ConversationTitlePayload, StreamChunkPayload } from '../../../types/events';
import type { Message } from '../../../types';
import { onEvent, offEvent } from '../../../core/infrastructure/wails/events';
import {
    appendStreamChunk,
    finalizeMessage,
    setConversationTitle,
    setError,
    setMessageError,
    startStreaming,
    stopStreaming,
    upsertMessage,
} from '../state/chatSignals';
import { log } from '../../../lib/logger';

let chatEventsInitialized = false;

/**
 * attach backend chat event listeners.
 */
export function initChatEvents(): void {
    if (chatEventsInitialized) return;
    chatEventsInitialized = true;
    onEvent<ChatEvent>('chat:event', (event) => {
        handleChatEvent(event);
    });
}

/**
 * detach backend chat event listeners.
 */
export function teardownChatEvents(): void {
    if (!chatEventsInitialized) return;
    chatEventsInitialized = false;
    offEvent('chat:event');
}

/**
 * route a chat event to the correct domain handler.
 */
function handleChatEvent(event: ChatEvent): void {
    log.trace()
        .str('eventType', event.type)
        .str('conversationId', event.conversationId ?? 'none')
        .str('messageId', event.messageId ?? 'none')
        .msg('ChatEvents received chat event');

    if (!event?.conversationId) {
        log.warn().msg('ChatEvents: event missing conversationId, ignoring');
        return;
    }

    switch (event.type) {
        case 'chat.message':
            handleMessage(event);
            break;

        case 'chat.stream.start':
            handleStreamStart(event);
            break;

        case 'chat.stream.chunk':
            handleStreamChunk(event);
            break;

        case 'chat.stream.error':
            handleStreamError(event);
            break;

        case 'chat.stream.complete':
            handleStreamComplete(event);
            break;

        case 'chat.conversation.title':
            handleConversationTitle(event);
            break;

        default:
            break;
    }
}

/**
 * upsert a completed chat message into the store.
 */
function handleMessage(event: ChatEvent): void {
    const message = event.payload as Message | undefined;
    if (!message) return;
    upsertMessage(event.conversationId, message);
}

/**
 * create the streaming message and flip streaming state on.
 */
function handleStreamStart(event: ChatEvent): void {
    const message = event.payload as Message | undefined;
    if (!message) return;
    upsertMessage(event.conversationId, message);
    startStreaming(event.conversationId, message.id);
}

/**
 * append a streaming content chunk to the active message.
 */
function handleStreamChunk(event: ChatEvent): void {
    const payload = event.payload as StreamChunkPayload | undefined;
    if (!payload || !event.messageId) return;

    appendStreamChunk({
        conversationId: event.conversationId,
        messageId: event.messageId,
        blockIndex: payload.blockIndex ?? 0,
        content: payload.content ?? '',
        isDone: false,
        metadata: payload.metadata,
    });
}

/**
 * handle streaming errors and update error state and message blocks.
 */
function handleStreamError(event: ChatEvent): void {
    const payload = event.payload as StreamChunkPayload | undefined;
    const errorMessage = payload?.error ?? 'An unknown error occurred';

    log.info()
        .str('conversationId', event.conversationId)
        .str('messageId', event.messageId ?? 'none')
        .str('error', errorMessage)
        .msg('ChatEvents: handling stream error');

    setError(errorMessage);
    stopStreaming();

    if (event.messageId) {
        setMessageError(event.conversationId, event.messageId, errorMessage);
    }
}

/**
 * finalize streaming state and message metadata on completion.
 */
function handleStreamComplete(event: ChatEvent): void {
    const payload = event.payload as StreamChunkPayload | undefined;

    log.info()
        .str('conversationId', event.conversationId)
        .str('messageId', event.messageId ?? 'none')
        .str('finishReason', payload?.metadata?.finishReason ?? 'unknown')
        .msg('ChatEvents: stream complete');

    stopStreaming();

    if (event.messageId) {
        finalizeMessage(event.conversationId, event.messageId, payload?.metadata);
    }
}

/**
 * update the conversation title from backend updates.
 */
function handleConversationTitle(event: ChatEvent): void {
    const payload = event.payload as ConversationTitlePayload | undefined;
    if (!payload?.title) return;
    setConversationTitle(event.conversationId, payload.title);
}
