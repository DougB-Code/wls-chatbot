/**
 * NavRail.ts renders the left navigation rail and emits normalized nav events.
 * frontend/src/shell/NavRail.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { SectionId } from '../types/shell';

/**
 * display navigation controls for switching workspaces.
 */
@customElement('wls-nav-rail')
export class NavRail extends LitElement {
    private static readonly _doubleClickDelayMs = 220;

    private _clickTimeoutId: number | null = null;

    static styles = css`
        :host {
            display: block;
            height: 100%;
        }

        .rail {
            height: 100%;
            min-height: 0;
            padding: 20px 10px 16px;
            box-sizing: border-box;
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 20px;
            border-right: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.06));
            background: rgba(9, 16, 26, 0.92);
            overflow: visible;
        }

        .rail__brand {
            width: 100%;
            display: flex;
            justify-content: center;
        }

        .rail__logo {
            width: 48px;
            height: 48px;
            object-fit: contain;
        }

        .rail__nav {
            flex: 1;
            min-height: 0;
            width: 100%;
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .rail__button {
            width: 100%;
            border-radius: 12px;
            padding: 10px 6px;
            border: 1px solid transparent;
            background: transparent;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 4px;
            cursor: pointer;
            font-size: 10px;
            letter-spacing: 0.04em;
            text-transform: uppercase;
            transition: all 120ms ease-out;
        }

        .rail__button:hover {
            background: var(--color-interactive-hover, hsla(0, 0%, 100%, 0.08));
            color: var(--color-text-secondary);
        }

        .rail__button--active {
            background: hsla(215, 65%, 62%, 0.15);
            border-color: hsla(215, 65%, 62%, 0.35);
            color: var(--color-text-primary);
        }

        .rail__icon {
            width: 22px;
            height: 22px;
        }

        .rail__icon svg {
            width: 100%;
            height: 100%;
            fill: currentColor;
        }

        .rail__icon--stroke svg {
            fill: none;
            stroke: currentColor;
            stroke-width: 2;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        .rail__footer {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 2px;
            font-size: 9px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: var(--color-text-muted);
            width: 100%;
            padding: 0 4px;
        }

        .rail__footer > span {
            max-width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .rail__footer-value {
            font-size: 10px;
            color: var(--color-text-secondary);
        }
    `;

    @property()
    activeSection: SectionId = 'chat';

    @property()
    logoUrl = '';

    /**
     * emit a single-click nav event for the selected section.
     */
    private _emitNavClick(section: SectionId) {
        this.dispatchEvent(new CustomEvent('nav-click', { detail: { section } }));
    }

    /**
     * emit a double-click nav event for the selected section.
     */
    private _emitNavDoubleClick(section: SectionId) {
        this.dispatchEvent(new CustomEvent('nav-dblclick', { detail: { section } }));
    }

    /**
     * clear any pending single-click dispatch timer.
     */
    private _clearClickTimeout() {
        if (this._clickTimeoutId === null) {
            return;
        }
        window.clearTimeout(this._clickTimeoutId);
        this._clickTimeoutId = null;
    }

    /**
     * normalize click interactions so double-clicks do not trigger single-click logic.
     */
    private _handleRailClick(event: MouseEvent, section: SectionId) {
        if (event.detail >= 2) {
            this._clearClickTimeout();
            this._emitNavDoubleClick(section);
            return;
        }

        this._clearClickTimeout();
        this._clickTimeoutId = window.setTimeout(() => {
            this._clickTimeoutId = null;
            this._emitNavClick(section);
        }, NavRail._doubleClickDelayMs);
    }

    /**
     * clear timers before unmounting.
     */
    disconnectedCallback() {
        this._clearClickTimeout();
        super.disconnectedCallback();
    }

    /**
     * render the rail UI and section controls.
     */
    render() {
        return html`
            <aside class="rail">
                <div class="rail__brand">
                    ${this.logoUrl ? html`<img class="rail__logo" src=${this.logoUrl} alt="WLS ChatBot" />` : null}
                </div>

                <nav class="rail__nav">
                    <button
                        class="rail__button ${this.activeSection === 'chat' ? 'rail__button--active' : ''}"
                        @click=${(event: MouseEvent) => this._handleRailClick(event, 'chat')}
                        title="Chat"
                    >
                        <span class="rail__icon rail__icon--stroke">
                            <svg viewBox="0 0 24 24">
                                <path d="M12 6V2H8" />
                                <path d="m9 11 0 2" />
                                <path d="m15 11 0 2" />
                                <path d="M2 12h2" />
                                <path d="M20 12h2" />
                                <path d="M20 16a2 2 0 0 1-2 2H8.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 4 20.286V8a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2z" />
                            </svg>
                        </span>
                        Chat
                    </button>







                    <button
                        class="rail__button ${this.activeSection === 'settings' ? 'rail__button--active' : ''}"
                        @click=${(event: MouseEvent) => this._handleRailClick(event, 'settings')}
                        title="Workspace Settings"
                    >
                        <span class="rail__icon rail__icon--stroke">
                            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-settings-icon lucide-settings"><path d="M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915"/><circle cx="12" cy="12" r="3"/></svg>
                        </span>
                        Settings
                    </button>
                </nav>

                <div class="rail__footer">
                    <span>Local</span>
                    <span class="rail__footer-value">0.1.0</span>
                </div>
            </aside>
        `;
    }
}
