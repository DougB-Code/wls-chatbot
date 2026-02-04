/**
 * translate toast events into toast state updates.
 * frontend/src/app/application/toastPolicy.ts
 */

import { onEvent, offEvent } from '../../core/infrastructure/wails/events';
import { pushToast, dismissToast } from '../state/toastSignals';

export type ToastPayload = {
    type?: 'info' | 'error';
    title?: string;
    message?: string;
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
        const id = pushToast({
            type: payload?.type === 'error' ? 'error' : 'info',
            title: payload?.title ?? (payload?.type === 'error' ? 'Connection issue' : ''),
            message,
        });
        window.setTimeout(() => dismissToast(id), 8000);
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
