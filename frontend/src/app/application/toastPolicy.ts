/**
 * translate toast events into toast state updates.
 * frontend/src/app/application/toastPolicy.ts
 */

import { onEvent, offEvent } from '../../core/infrastructure/wails/events';
import { pushToast } from '../state/toastSignals';
import { createNotification } from '../../features/notifications/application/notificationPolicy';

export type ToastPayload = {
    type?: 'info' | 'error';
    title?: string;
    message?: string;
    durationMs?: number;
};

let toastEventsInitialized = false;

/**
 * attach toast event listeners and push toast state updates.
 */
export function initToastEvents(): void {
    if (toastEventsInitialized) return;
    toastEventsInitialized = true;
    onEvent<ToastPayload>('toast', (payload) => {
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
        }).catch(() => undefined);
    });
}

/**
 * detach toast event listeners.
 */
export function teardownToastEvents(): void {
    if (!toastEventsInitialized) return;
    toastEventsInitialized = false;
    offEvent('toast');
}
