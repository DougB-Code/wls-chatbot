/**
 * NotificationsHeader.ts renders the notifications header content for the app shell.
 * frontend/src/features/notifications/ui/NotificationsHeader.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';

import { notifications } from '../state/notificationSignals';

/**
 * render the notifications header with counts.
 */
@customElement('wls-notifications-header')
export class NotificationsHeader extends SignalWatcher(LitElement) {
    static styles = css`
        :host {
            display: flex;
            flex-direction: column;
            gap: 2px;
        }

        .title {
            margin: 0;
            font-size: 18px;
            font-weight: 600;
        }

        .subtitle {
            margin: 0;
            font-size: 13px;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
        }
    `;

    /**
     * render the notifications header content.
     */
    render() {
        const count = notifications.value.length;
        const label = count === 1 ? '1 notification' : `${count} notifications`;

        return html`
            <h1 class="title">Notifications</h1>
            <p class="subtitle">${label} saved in this workspace.</p>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-notifications-header': NotificationsHeader;
    }
}
