/**
 * notificationTransport.ts bridges Wails notification APIs.
 * frontend/src/features/notifications/infrastructure/notificationTransport.ts
 */

import { CreateNotification, ListNotifications } from '../../../../wailsjs/go/wails/Bridge';
import type { notifications } from '../../../../wailsjs/go/models';

export type NotificationPayload = notifications.NotificationPayload;
export type Notification = notifications.Notification;

/**
 * fetch the current list of notifications from the backend.
 */
export async function listNotifications(): Promise<Notification[]> {
    return ListNotifications();
}

/**
 * persist a notification via the backend.
 */
export async function createNotification(payload: NotificationPayload): Promise<Notification | null> {
    return CreateNotification(payload);
}
