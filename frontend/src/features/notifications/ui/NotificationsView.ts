/**
 * NotificationsView.ts renders the notifications list workspace.
 * frontend/src/features/notifications/ui/NotificationsView.ts
 */

import { LitElement, html, css, nothing } from 'lit';
import { customElement } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';

import { notifications } from '../state/notificationSignals';
import { clearAllNotifications, deleteNotification, refreshNotifications } from '../application/notificationPolicy';
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
            background: #161A22;
        }

        .notifications__header {
            display: flex;
            align-items: center;
            justify-content: flex-end;
            gap: 16px;
        }

        .clear-button {
            border-radius: 999px;
            border: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.12));
            background: transparent;
            color: var(--color-text-secondary, hsla(0, 0%, 100%, 0.68));
            font-size: 12px;
            padding: 6px 12px;
            cursor: pointer;
            transition: all 120ms ease-out;
        }

        .clear-button:hover {
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
            border-color: var(--color-border-default, hsla(0, 0%, 100%, 0.2));
            background: var(--color-interactive-hover, hsla(0, 0%, 100%, 0.08));
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

        .notification__actions {
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }

        .notification__delete {
            border: 0;
            background: transparent;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
            cursor: pointer;
            padding: 0;
            line-height: 1;
            width: 24px;
            height: 24px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            border-radius: 6px;
            transition: all 120ms ease-out;
        }

        .notification__delete:hover {
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
            background: var(--color-interactive-hover, hsla(0, 0%, 100%, 0.08));
        }

        .notification__delete svg {
            width: 14px;
            height: 14px;
            stroke: currentColor;
            fill: none;
            stroke-width: 2;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        .notification__title {
            margin: 0;
            font-size: 14px;
            font-weight: 600;
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }

        .notification__count {
            font-size: 11px;
            font-weight: 600;
            letter-spacing: 0.04em;
            text-transform: uppercase;
            padding: 2px 8px;
            border-radius: 999px;
            background: rgba(255, 255, 255, 0.08);
            color: var(--color-text-secondary, hsla(0, 0%, 100%, 0.68));
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
                    <div class="notifications__header">
                        <button class="clear-button" type="button" disabled>
                            Clear All
                        </button>
                    </div>
                    <div class="empty-state">No notifications yet.</div>
                </div>
            `;
        }

        const grouped = this._groupNotifications(list);
        return html`
            <div class="notifications">
                <div class="notifications__header">
                    <button class="clear-button" type="button" @click=${this._handleClearAll}>
                        Clear All
                    </button>
                </div>
                ${grouped.map((group) => this._renderNotification(group))}
            </div>
        `;
    }

    /**
     * group duplicate notifications by title/message/type.
     */
    private _groupNotifications(list: Notification[]) {
        const groups = new Map<string, { notification: Notification; count: number; ids: number[] }>();
        for (const notification of list) {
            const title = notification.title || (notification.type === 'error' ? 'Error' : 'Notification');
            const key = `${notification.type}::${title}::${notification.message}`;
            const group = groups.get(key);
            if (group) {
                group.count += 1;
                group.ids.push(notification.id);
                if (notification.createdAt > group.notification.createdAt) {
                    group.notification = notification;
                }
            } else {
                groups.set(key, { notification, count: 1, ids: [notification.id] });
            }
        }
        return Array.from(groups.values());
    }

    /**
     * render a single notification card.
     */
    private _renderNotification(group: { notification: Notification; count: number; ids: number[] }) {
        const { notification, count, ids } = group;
        const title = notification.title || (notification.type === 'error' ? 'Error' : 'Notification');
        return html`
            <article class="notification ${notification.type === 'error' ? 'notification--error' : ''}">
                <div class="notification__header">
                    <h2 class="notification__title">
                        ${title}
                        ${count > 1 ? html`<span class="notification__count">x${count}</span>` : nothing}
                    </h2>
                    <div class="notification__actions">
                        <span class="notification__timestamp">${this._formatTimestamp(notification.createdAt)}</span>
                        <button
                            class="notification__delete"
                            type="button"
                            aria-label="Delete notification"
                            title="Delete notification"
                            @click=${() => this._handleDeleteGroup(ids)}
                        >
                            <svg viewBox="0 0 24 24" aria-hidden="true">
                                <path d="M18 6L6 18" />
                                <path d="M6 6l12 12" />
                            </svg>
                        </button>
                    </div>
                </div>
                <p class="notification__message">${notification.message}</p>
            </article>
        `;
    }

    /**
     * delete a notification by id.
     */
    private _handleDeleteGroup(ids: number[]) {
        for (const id of ids) {
            void deleteNotification(id).catch(() => undefined);
        }
    }

    /**
     * clear all notifications.
     */
    private _handleClearAll() {
        void clearAllNotifications().catch(() => undefined);
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
