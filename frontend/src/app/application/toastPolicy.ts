/**
 * translate toast events into toast state updates.
 * frontend/src/app/application/toastPolicy.ts
 */

import { onEvent, type EventUnsubscribe } from '../../core/infrastructure/wails/events';
import { pushToast } from '../state/toastSignals';
import { createNotification } from '../../features/notifications/application/notificationPolicy';
import { log } from '../../lib/logger';

export interface ToastPayload {
    type?: 'info' | 'error';
    title?: string;
    message?: string;
    durationMs?: number;
}

let toastEventsInitialized = false;
let unsubscribeToastEvents: EventUnsubscribe | null = null;

/**
 * attach toast event listeners and push toast state updates.
 */
export function initToastEvents(): void {
    if (toastEventsInitialized) return;
    toastEventsInitialized = true;
    unsubscribeToastEvents = onEvent<ToastPayload>('toast', (payload) => {
        const message = payload?.message ?? 'Something went wrong.';
        const type = payload?.type === 'error' ? 'error' : 'info';
        const title = payload?.title ?? (type === 'error' ? 'Connection issue' : '');
        pushToast({
            type,
            title,
            message,
            durationMs: payload?.durationMs,
        });
        void createNotification({
            type,
            title,
            message,
        }).catch((err) => {
            const errorMessage = err instanceof Error ? err.message : 'Unknown notification error';
            log.warn().str('error', errorMessage).msg('Failed to persist toast notification');
        });
    });
}

/**
 * detach toast event listeners.
 */
export function teardownToastEvents(): void {
    if (!toastEventsInitialized) return;
    toastEventsInitialized = false;
    if (unsubscribeToastEvents) {
        unsubscribeToastEvents();
        unsubscribeToastEvents = null;
    }
}
