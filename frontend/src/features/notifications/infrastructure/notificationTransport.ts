/**
 * notificationTransport.ts persists notifications in frontend local storage.
 * frontend/src/features/notifications/infrastructure/notificationTransport.ts
 */

import type { Notification, NotificationPayload } from '../domain/types';

const storageKey = 'wls.notifications';

/**
 * read notifications from local storage.
 */
function readNotifications(): Notification[] {
    if (typeof globalThis.localStorage === 'undefined') {
        return [];
    }
    try {
        const raw = globalThis.localStorage.getItem(storageKey);
        if (!raw) {
            return [];
        }
        const parsed = JSON.parse(raw) as Notification[];
        if (!Array.isArray(parsed)) {
            return [];
        }
        return parsed
            .filter((item) => typeof item?.id === 'number' && typeof item?.message === 'string' && typeof item?.createdAt === 'number')
            .sort((a, b) => b.createdAt - a.createdAt);
    } catch {
        return [];
    }
}

/**
 * write notifications to local storage.
 */
function writeNotifications(list: Notification[]): void {
    if (typeof globalThis.localStorage === 'undefined') {
        return;
    }
    globalThis.localStorage.setItem(storageKey, JSON.stringify(list));
}

/**
 * fetch the current list of notifications from local storage.
 */
export async function listNotifications(): Promise<Notification[]> {
    return readNotifications();
}

/**
 * persist a notification in local storage.
 */
export async function createNotification(payload: NotificationPayload): Promise<Notification | null> {
    const message = (payload.message ?? '').trim();
    if (message === '') {
        return null;
    }
    const existing = readNotifications();
    const nextID = existing.reduce((maxID, item) => Math.max(maxID, item.id), 0) + 1;
    const item: Notification = {
        id: nextID,
        type: payload.type === 'error' ? 'error' : 'info',
        title: payload.title?.trim() || undefined,
        message,
        createdAt: Date.now(),
    };
    writeNotifications([item, ...existing]);
    return item;
}

/**
 * delete a notification by id.
 */
export async function deleteNotification(id: number): Promise<boolean> {
    const existing = readNotifications();
    const next = existing.filter((item) => item.id !== id);
    if (next.length === existing.length) {
        return false;
    }
    writeNotifications(next);
    return true;
}

/**
 * clear all notifications.
 */
export async function clearNotifications(): Promise<boolean> {
    writeNotifications([]);
    return true;
}
