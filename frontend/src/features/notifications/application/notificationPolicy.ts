/**
 * orchestrate notification persistence and state updates.
 * frontend/src/features/notifications/application/notificationPolicy.ts
 */

import * as notificationTransport from '../infrastructure/notificationTransport';
import { clearNotifications, prependNotification, removeNotification, setNotifications } from '../state/notificationSignals';

export type NotificationPayload = notificationTransport.NotificationPayload;

/**
 * refresh notifications from the backend store.
 */
export async function refreshNotifications(): Promise<void> {
    const list = await notificationTransport.listNotifications();
    setNotifications(list);
}

/**
 * create a new notification in the backend and update state.
 */
export async function createNotification(payload: NotificationPayload): Promise<void> {
    const notification = await notificationTransport.createNotification(payload);
    if (notification) {
        prependNotification(notification);
    }
}

/**
 * delete a notification and update local state.
 */
export async function deleteNotification(id: number): Promise<void> {
    const ok = await notificationTransport.deleteNotification(id);
    if (ok) {
        removeNotification(id);
    }
}

/**
 * clear all notifications and update local state.
 */
export async function clearAllNotifications(): Promise<void> {
    const ok = await notificationTransport.clearNotifications();
    if (ok) {
        clearNotifications();
    }
}
