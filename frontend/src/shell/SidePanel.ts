/**
 * SidePanel.ts renders shared side panel chrome for shell content.
 * frontend/src/shell/SidePanel.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';

/**
 * render a shared side panel wrapper for shell content.
 */
@customElement('wls-shell-side-panel')
export class SidePanel extends LitElement {
    static styles = css`
        :host {
            display: block;
            height: 100%;
        }

        .panel {
            height: 100%;
            min-height: 0;
            padding: 20px 16px;
            display: flex;
            flex-direction: column;
            gap: 20px;
            box-sizing: border-box;
            overflow-y: auto;
        }
    `;

    /**
     * render the shared side panel layout wrapper.
     */
    render() {
        return html`
            <section class="panel">
                <slot></slot>
            </section>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-shell-side-panel': SidePanel;
    }
}
