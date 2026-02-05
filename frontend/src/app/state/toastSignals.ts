/**
 * manage toast notification state in signals.
 * frontend/src/app/state/toastSignals.ts
 */

import { signal } from '@preact/signals-core';

export type ToastType = 'info' | 'error';

export type ToastMessage = {
    id: number;
    type: ToastType;
    title?: string;
    message: string;
    createdAt: number;
    durationMs: number;
};

export const toasts = signal<ToastMessage[]>([]);

let toastId = 0;

/**
 * add a toast to state and return its id.
 */
export function pushToast(payload: Omit<ToastMessage, 'id' | 'createdAt' | 'durationMs'> & { durationMs?: number }): number {
    const id = ++toastId;
    const createdAt = Date.now();
    const durationMs = payload.durationMs ?? 8000;
    toasts.value = [...toasts.value, { ...payload, id, createdAt, durationMs }];
    return id;
}

/**
 * remove a toast by id.
 */
export function dismissToast(id: number): void {
    toasts.value = toasts.value.filter((toast) => toast.id !== id);
}

/**
 * clear all toast notifications.
 */
export function clearToasts(): void {
    toasts.value = [];
}
