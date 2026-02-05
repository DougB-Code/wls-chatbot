/**
 * NotificationsView.ts renders the notifications list workspace.
 * frontend/src/features/notifications/ui/NotificationsView.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';

import { notifications } from '../state/notificationSignals';
import { refreshNotifications } from '../application/notificationPolicy';
import type { Notification } from '../../../types/wails';

/**
 * render stored notifications for the workspace.
 */
@customElement('wls-notifications-view')
export class NotificationsView extends SignalWatcher(LitElement) {
    static styles = css`
        :host {
            display: block;
            height: 100%;
            min-height: 0;
        }

        .notifications {
            height: 100%;
            min-height: 0;
            padding: 24px 32px;
            display: flex;
            flex-direction: column;
            gap: 16px;
            overflow-y: auto;
            box-sizing: border-box;
        }

        .notification {
            border-radius: 14px;
            border: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.06));
            background: var(--color-bg-elevated, hsl(220, 22%, 11%));
            padding: 16px;
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .notification--error {
            border-color: var(--color-error-border, hsla(0, 68%, 55%, 0.3));
            background: var(--color-error-surface, hsla(0, 68%, 55%, 0.1));
        }

        .notification__header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
        }

        .notification__title {
            margin: 0;
            font-size: 14px;
            font-weight: 600;
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
        }

        .notification__timestamp {
            font-size: 11px;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
        }

        .notification__message {
            margin: 0;
            font-size: 13px;
            color: var(--color-text-secondary, hsla(0, 0%, 100%, 0.68));
            line-height: 1.5;
            white-space: pre-wrap;
        }

        .empty-state {
            margin: auto;
            text-align: center;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
            font-size: 14px;
        }
    `;

    /**
     * refresh notification data when the view mounts.
     */
    connectedCallback() {
        super.connectedCallback();
        void refreshNotifications().catch(() => undefined);
    }

    /**
     * render the notifications list or an empty state.
     */
    render() {
        const list = notifications.value;
        if (list.length === 0) {
            return html`
                <div class="notifications">
                    <div class="empty-state">No notifications yet.</div>
                </div>
            `;
        }

        return html`
            <div class="notifications">
                ${list.map((notification) => this._renderNotification(notification))}
            </div>
        `;
    }

    /**
     * render a single notification card.
     */
    private _renderNotification(notification: Notification) {
        const title = notification.title || (notification.type === 'error' ? 'Error' : 'Notification');
        return html`
            <article class="notification ${notification.type === 'error' ? 'notification--error' : ''}">
                <div class="notification__header">
                    <h2 class="notification__title">${title}</h2>
                    <span class="notification__timestamp">${this._formatTimestamp(notification.createdAt)}</span>
                </div>
                <p class="notification__message">${notification.message}</p>
            </article>
        `;
    }

    /**
     * format a timestamp into a localized short label.
     */
    private _formatTimestamp(timestamp: number) {
        const date = new Date(timestamp);
        return date.toLocaleString(undefined, {
            month: 'short',
            day: 'numeric',
            hour: 'numeric',
            minute: '2-digit',
        });
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-notifications-view': NotificationsView;
    }
}
