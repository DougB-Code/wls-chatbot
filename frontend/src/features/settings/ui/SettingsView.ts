/**
 * provide the outer settings page shell and padding.
 * frontend/src/features/settings/ui/SettingsView.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import './AiGateways';
import { settingsSharedStyles } from './settingsStyles';

/**
 * host the settings panel and its child content.
 */
@customElement('wls-settings-view')
export class SettingsView extends LitElement {
    static styles = [
        settingsSharedStyles,
        css`
        :host {
            display: block;
            height: 100%;
            box-sizing: border-box;
        }

        .settings-panel {
            padding: 24px 32px;
            display: flex;
            flex-direction: column;
            gap: 24px;
            width: 100%;
            height: 100%;
            overflow-y: auto;
            box-sizing: border-box;
        }

    `,
    ];

    /**
     * render the settings panel wrapper.
     */
    render() {
        return html`
            <div class="settings-panel">
                <wls-connections-view></wls-connections-view>
            </div>
        `;
    }
}
