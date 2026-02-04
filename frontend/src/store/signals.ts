/**
 * signals.ts manages chat state with fine-grained reactive signals.
 * frontend/src/store/signals.ts
 */

import { signal, computed } from '@preact/signals-core';
import type { Conversation, Message, StreamChunk, MessageMetadata, ActionExecution } from '../types';
import { log } from '../lib/logger';

// ─────────────────────────────────────────────────────
// Core State Signals
// ─────────────────────────────────────────────────────

export const conversations = signal<Map<string, Conversation>>(new Map());
export const activeId = signal<string | null>(null);
export const isStreaming = signal(false);
export const streamingMessageId = signal<string | null>(null);
export const error = signal<string | null>(null);

// ─────────────────────────────────────────────────────
// Computed Signals
// ─────────────────────────────────────────────────────

export const activeConversation = computed(() => {
    const id = activeId.value;
    return id ? conversations.value.get(id) : undefined;
});

// ─────────────────────────────────────────────────────
// Actions
// ─────────────────────────────────────────────────────

/**
 * clone a Wails model while preserving its prototype.
 */
function cloneModel<T extends object>(source: T, updates: Partial<T>): T {
    return Object.assign(Object.create(Object.getPrototypeOf(source)), source, updates);
}

/**
 * persist a conversation update to the signal store.
 */
function setConversation(updated: Conversation): void {
    const map = new Map(conversations.value);
    map.set(updated.id, updated);
    conversations.value = map;
}

/**
 * set the active conversation id signal.
 */
export function setActiveConversation(id: string | null): void {
    log.info().str('activeId', id ?? 'null').msg('signals: setActiveConversation');
    activeId.value = id;
}

/**
 * insert or update a conversation in the signal store.
 */
export function upsertConversation(conv: Conversation, setActive = false): void {
    log.info()
        .str('convId', conv.id)
        .str('setActive', String(setActive))
        .str('messageCount', String(conv.messages?.length ?? 0))
        .msg('signals: upsertConversation');
    const normalizedMessages = conv.messages ? [...conv.messages] : [];
    const normalizedConversation = cloneModel(conv, { messages: normalizedMessages });
    setConversation(normalizedConversation);
    if (setActive) {
        log.info().str('activeId', conv.id).msg('signals: setting activeId');
        activeId.value = conv.id;
    }
}

/**
 * report whether a conversation exists in the store.
 */
export function hasConversation(id: string): boolean {
    return conversations.value.has(id);
}

/**
 * fetch a conversation by id from the store.
 */
export function getConversation(id: string): Conversation | undefined {
    return conversations.value.get(id);
}

/**
 * set the title for an existing conversation.
 */
export function setConversationTitle(conversationId: string, title: string): void {
    const conv = conversations.value.get(conversationId);
    if (!conv) return;
    const updated = cloneModel(conv, { title, updatedAt: Date.now() });
    setConversation(updated);
}


/**
 * insert or update a message within a conversation.
 */
export function upsertMessage(conversationId: string, message: Message): void {
    const conv = conversations.value.get(conversationId);
    if (!conv) {
        log.warn()
            .str('conversationId', conversationId)
            .str('messageId', message.id)
            .msg('signals: upsertMessage FAILED - conversation not found');
        return;
    }

    log.trace()
        .str('conversationId', conversationId)
        .str('messageId', message.id)
        .str('role', message.role)
        .str('blocks', String(message.blocks.length))
        .msg('signals: upsertMessage');

    const messages = [...conv.messages];
    const msgIndex = messages.findIndex((m) => m.id === message.id);
    if (msgIndex >= 0) {
        log.trace().str('action', 'update').msg('signals: upsertMessage updating existing');
        const updatedMessage = cloneModel(messages[msgIndex], message);
        updatedMessage.blocks = updatedMessage.blocks ?? [];
        messages[msgIndex] = updatedMessage;
    } else {
        log.trace().str('action', 'create').msg('signals: upsertMessage creating new');
        const normalizedMessage = cloneModel(message, { blocks: message.blocks ?? [] });
        messages.push(normalizedMessage);
    }
    const updatedConv = cloneModel(conv, { messages, updatedAt: Date.now() });
    setConversation(updatedConv);
}


/**
 * mark streaming state as active for the given message.
 */
export function startStreaming(conversationId: string, messageId: string): void {
    log.trace().str('conversationId', conversationId).str('messageId', messageId).msg('signals: startStreaming');
    isStreaming.value = true;
    streamingMessageId.value = messageId;
}

/**
 * append streaming content or errors to a message block.
 */
export function appendStreamChunk(chunk: StreamChunk): void {
    const convId = chunk.conversationId ?? activeId.value;
    const blockIndex = chunk.blockIndex ?? 0;

    log.trace()
        .str('conversationId', convId ?? 'null')
        .str('messageId', chunk.messageId)
        .str('isDone', String(chunk.isDone))
        .str('error', chunk.error ?? '')
        .msg('signals: appendStreamChunk');

    if (!convId) return; // Simplified for brevity

    const conv = conversations.value.get(convId);
    if (!conv) {
        log.warn().str('conversationId', convId).msg('signals: appendStreamChunk - conv not found');
        return;
    }

    const messages = [...conv.messages];
    let msgIndex = messages.findIndex((m) => m.id === chunk.messageId);
    let msg: Message;

    if (msgIndex === -1) {
        log.trace().msg('signals: appendStreamChunk - creating placeholder');
        // Create placeholder message if not found
        msg = {
            id: chunk.messageId,
            conversationId: convId,
            role: 'assistant',
            blocks: [],
            timestamp: Date.now(),
            isStreaming: true,
        };
        messages.push(msg);
        msgIndex = messages.length - 1;
    } else {
        // Clone existing message
        msg = cloneModel(messages[msgIndex], {});
    }

    // Handle error vs normal content
    // Note: blocks array should also be cloned if we were strictly immutable, 
    // but message clone is enough for Lit property change on <wls-message-bubble>
    const blocks = [...(msg.blocks ?? [])];

    if (chunk.error) {
        log.trace().str('error', chunk.error).msg('signals: appendStreamChunk - handling error chunk');
        blocks[blockIndex] = { type: 'error', content: chunk.error };
        error.value = chunk.error;
    } else if (blocks[blockIndex]) {
        const existing = blocks[blockIndex];
        const existingContent = existing?.content ?? '';
        blocks[blockIndex] = { ...existing, content: existingContent + chunk.content };
    } else {
        blocks[blockIndex] = { type: 'text', content: chunk.content };
    }

    // Update streaming state
    let metadata = msg.metadata;
    if (chunk.isDone) {
        log.trace().msg('signals: appendStreamChunk - done');
        isStreaming.value = false;
        streamingMessageId.value = null;
        metadata = chunk.metadata ?? metadata;
    } else {
        isStreaming.value = true;
        streamingMessageId.value = chunk.messageId;
    }

    // Update message in conversation array
    const updatedMessage = cloneModel(msg, {
        blocks,
        isStreaming: !chunk.isDone,
        metadata,
    });
    messages[msgIndex] = updatedMessage;

    const updatedConv = cloneModel(conv, { messages, updatedAt: Date.now() });
    setConversation(updatedConv);
}


/**
 * set the global error signal.
 */
export function setError(errorMessage: string): void {
    log.trace().str('error', errorMessage).msg('signals: setError');
    error.value = errorMessage;
}

/**
 * clear streaming state signals.
 */
export function stopStreaming(): void {
    log.trace().msg('signals: stopStreaming');
    isStreaming.value = false;
    streamingMessageId.value = null;
}

/**
 * attach an error block to a specific message.
 */
export function setMessageError(conversationId: string, messageId: string, errorMessage: string): void {
    log.trace()
        .str('conversationId', conversationId)
        .str('messageId', messageId)
        .str('error', errorMessage)
        .msg('signals: setMessageError');

    const conv = conversations.value.get(conversationId);
    if (!conv) {
        log.warn().str('conversationId', conversationId).msg('signals: setMessageError - conv not found');
        return;
    }

    const messages = [...conv.messages];
    let msgIndex = messages.findIndex((m) => m.id === messageId);
    let msg: Message;

    if (msgIndex === -1) {
        // Create the message if it doesn't exist
        msg = {
            id: messageId,
            conversationId,
            role: 'assistant',
            blocks: [],
            timestamp: Date.now(),
            isStreaming: false,
        };
        messages.push(msg);
        msgIndex = messages.length - 1;
    } else {
        msg = cloneModel(messages[msgIndex], {});
    }

    // Set the first block to an error block
    const blocks = [...(msg.blocks ?? [])];
    blocks[0] = { type: 'error', content: errorMessage };

    // Update message in conversation
    const updatedMessage = cloneModel(msg, { blocks, isStreaming: false });
    messages[msgIndex] = updatedMessage;

    const updatedConv = cloneModel(conv, { messages, updatedAt: Date.now() });
    setConversation(updatedConv);
}

/**
 * finalize a message by stopping streaming and applying metadata.
 */
export function finalizeMessage(conversationId: string, messageId: string, metadata?: MessageMetadata): void {
    log.trace().str('conversationId', conversationId).str('messageId', messageId).msg('signals: finalizeMessage');
    const conv = conversations.value.get(conversationId);
    if (!conv) return;

    const messages = [...conv.messages];
    const msgIndex = messages.findIndex((m) => m.id === messageId);
    if (msgIndex === -1) return;

    // Clone message
    const msg = cloneModel(messages[msgIndex], {});
    const updatedMessage = cloneModel(msg, {
        isStreaming: false,
        metadata: metadata ?? msg.metadata,
    });

    // Update message in conversation
    messages[msgIndex] = updatedMessage;

    const updatedConv = cloneModel(conv, { messages, updatedAt: Date.now() });
    setConversation(updatedConv);
}

/**
 * update the status/result of an action embedded in message blocks.
 */
export function updateActionStatus(
    conversationId: string,
    actionId: string,
    status: ActionExecution['status'],
    result?: string
): void {
    const conv = conversations.value.get(conversationId);
    if (!conv) return;

    let updated = false;
    const updatedMessages = conv.messages.map((msg) => {
        let msgUpdated = false;
        const updatedBlocks = msg.blocks.map((block) => {
            if (!block.action || block.action.id !== actionId) {
                return block;
            }

            msgUpdated = true;
            updated = true;
            return {
                ...block,
                action: {
                    ...block.action,
                    status,
                    result: result ?? block.action.result,
                },
            };
        });

        return msgUpdated ? cloneModel(msg, { blocks: updatedBlocks }) : msg;
    });

    if (!updated) return;

    const updatedConv = cloneModel(conv, { messages: updatedMessages, updatedAt: Date.now() });
    setConversation(updatedConv);
}

/**
 * mark an action as approved with a default result message.
 */
export function approveAction(conversationId: string, actionId: string): void {
    updateActionStatus(conversationId, actionId, 'approved', 'Approved by user');
}

/**
 * mark an action as rejected with a default result message.
 */
export function rejectAction(conversationId: string, actionId: string): void {
    updateActionStatus(conversationId, actionId, 'rejected', 'Rejected by user');
}


/**
 * reset all conversation-related signals to their defaults.
 */
export function clearAll(): void {
    conversations.value = new Map();
    activeId.value = null;
    isStreaming.value = false;
    streamingMessageId.value = null;
    error.value = null;
}
