/**
 * define UI-specific streaming types.
 */

import type { StreamChunkPayload } from './events';

/**
 * describe a streaming update to a message block.
 */
export type StreamChunk = StreamChunkPayload & {
    conversationId?: string;
    messageId: string;
};
