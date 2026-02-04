/**
 * verify signal-store behavior during streaming error flows.
 * frontend/src/features/chat/state/chatSignals.test.ts
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest';

import * as store from './chatSignals';
import type { StreamChunk } from '../../../types';

/**
 * ensure the test environment has a window object.
 */
const ensureWindow = () => {
    const globalAny = globalThis as unknown as { window?: Record<string, unknown> };
    if (!globalAny.window) {
        globalAny.window = {};
    }
};

describe('Signals store stream error handling', () => {
    beforeEach(() => {
        ensureWindow();
        store.clearAll();
    });

    afterEach(() => {
        store.clearAll();
    });

    it('stops streaming and renders error block when error arrives', () => {
        // Setup conversation
        store.upsertConversation({
            id: 'conv-1',
            title: 'Test',
            messages: [],
            settings: { provider: 'openai', model: 'gpt-4o' },
            createdAt: 0,
            updatedAt: 0,
            isArchived: false,
        }, true);

        // Simulate stream start
        store.upsertMessage('conv-1', {
            id: 'msg-1',
            conversationId: 'conv-1',
            role: 'assistant',
            blocks: [],
            timestamp: 1,
            isStreaming: true,
        });
        store.startStreaming('conv-1', 'msg-1');

        expect(store.isStreaming.value).toBe(true);

        // Simulate stream error
        const errorChunk: StreamChunk = {
            conversationId: 'conv-1',
            messageId: 'msg-1',
            blockIndex: 0,
            content: '',
            isDone: true,
            error: 'API error: 403',
            statusCode: 403,
        };

        store.appendStreamChunk(errorChunk);

        expect(store.isStreaming.value).toBe(false);
        expect(store.streamingMessageId.value).toBe(null);
        expect(store.error.value).toBe('API error: 403');

        const conv = store.getConversation('conv-1');
        const message = conv?.messages.find((m) => m.id === 'msg-1');
        expect(message?.isStreaming).toBe(false);
        expect(message?.blocks[0]?.type).toBe('error');
        expect(message?.blocks[0]?.content).toBe('API error: 403');
    });

    it('creates message when appendStreamChunk called without prior message', () => {
        store.upsertConversation({
            id: 'conv-1',
            title: 'Test',
            messages: [],
            settings: { provider: 'openai', model: 'gpt-4o' },
            createdAt: 0,
            updatedAt: 0,
            isArchived: false,
        }, true);

        const errorChunk: StreamChunk = {
            conversationId: 'conv-1',
            messageId: 'msg-2',
            blockIndex: 0,
            content: '',
            isDone: true,
            error: 'API error: 403',
            statusCode: 403,
        };

        store.appendStreamChunk(errorChunk);

        expect(store.isStreaming.value).toBe(false);

        const conv = store.getConversation('conv-1');
        const message = conv?.messages.find((m) => m.id === 'msg-2');
        expect(message).toBeTruthy();
        expect(message?.blocks[0]?.type).toBe('error');
        expect(message?.blocks[0]?.content).toBe('API error: 403');
    });
});
