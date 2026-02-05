/**
 * manage notification state for the notifications feature.
 * frontend/src/features/notifications/state/notificationSignals.ts
 */

import { signal } from '@preact/signals-core';
import type { Notification } from '../../../types/wails';

export const notifications = signal<Notification[]>([]);

/**
 * replace notification state with the provided list.
 */
export function setNotifications(list: Notification[]): void {
    notifications.value = [...list];
}

/**
 * remove a notification by id.
 */
export function removeNotification(id: number): void {
    notifications.value = notifications.value.filter((notification) => notification.id !== id);
}

/**
 * clear all notifications.
 */
export function clearNotifications(): void {
    notifications.value = [];
}

/**
 * prepend a newly created notification to state.
 */
export function prependNotification(notification: Notification): void {
    notifications.value = [notification, ...notifications.value];
}
