/**
 * types.ts defines notification domain types.
 * frontend/src/features/notifications/domain/types.ts
 */

export type NotificationType = 'info' | 'error';

export interface NotificationPayload {
    type?: NotificationType;
    title?: string;
    message: string;
}

export interface Notification {
    id: number;
    type: NotificationType;
    title?: string;
    message: string;
    createdAt: number;
}
