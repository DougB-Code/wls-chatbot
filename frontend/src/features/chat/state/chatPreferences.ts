/**
 * store chat UI preferences in signals.
 * frontend/src/features/chat/state/chatPreferences.ts
 */

import { signal } from '@preact/signals-core';

export const preferredModelId = signal<string>('');

/**
 * update the preferred model id for new conversations.
 */
export function setPreferredModelId(modelId: string): void {
    preferredModelId.value = modelId;
}
