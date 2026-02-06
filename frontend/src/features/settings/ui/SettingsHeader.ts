/**
 * SettingsHeader.ts renders the settings header content for the app shell.
 * frontend/src/features/settings/ui/SettingsHeader.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';

/**
 * render the settings header title and description.
 */
@customElement('wls-settings-header')
export class SettingsHeader extends LitElement {
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
     * render the settings header content.
     */
    render() {
        return html`
            <h1 class="title">AI Providers and Connections</h1>
            <p class="subtitle">Manage provider connections and workspace defaults.</p>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-settings-header': SettingsHeader;
    }
}
