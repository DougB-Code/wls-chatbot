/**
 * SettingsSidePanel.ts renders settings-specific side panel navigation.
 * frontend/src/features/settings/SettingsSidePanel.ts
 */

import { LitElement, css, html } from 'lit';
import { customElement } from 'lit/decorators.js';

/**
 * display settings context and section shortcuts.
 */
@customElement('wls-settings-side-panel')
export class SettingsSidePanel extends LitElement {
    static styles = css`
        :host {
            display: block;
            height: 100%;
        }

        .panel {
            height: 100%;
            min-height: 0;
            box-sizing: border-box;
            padding: 20px 16px;
            display: flex;
            flex-direction: column;
            gap: 20px;
        }

        .title {
            margin: 0;
            font-size: 15px;
            font-weight: 600;
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
        }

        .subtitle {
            margin: 4px 0 0;
            font-size: 12px;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
        }

        .section-list {
            list-style: none;
            margin: 0;
            padding: 0;
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .section-item {
            border: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.06));
            border-radius: 10px;
            padding: 10px 12px;
            background: var(--color-bg-elevated, hsl(220, 22%, 11%));
        }

        .section-item-title {
            margin: 0;
            font-size: 13px;
            font-weight: 600;
            color: var(--color-text-secondary, hsla(0, 0%, 100%, 0.68));
        }

        .section-item-description {
            margin: 4px 0 0;
            font-size: 12px;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
            line-height: 1.45;
        }
    `;

    /**
     * render settings side panel content.
     */
    render() {
        return html`
            <aside class="panel">
                <div>
                    <h2 class="title">Settings Panel</h2>
                    <p class="subtitle">Configuration sections for this workspace.</p>
                </div>

                <ul class="section-list">
                    <li class="section-item">
                        <h3 class="section-item-title">AI Gateways</h3>
                        <p class="section-item-description">Configure provider credentials and default models.</p>
                    </li>
                    <li class="section-item">
                        <h3 class="section-item-title">Coming Soon</h3>
                        <p class="section-item-description">Additional workspace settings will appear here.</p>
                    </li>
                </ul>
            </aside>
        `;
    }
}
