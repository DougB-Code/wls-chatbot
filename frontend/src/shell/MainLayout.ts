/**
 * define the top-level layout grid and navigation slot.
 */

import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import './NavRail';
import { SectionId } from '../types/shell';

/**
 * render the nav rail and workspace slot layout.
 */
@customElement('wls-main-layout')
export class MainLayout extends LitElement {
    static styles = css`
        :host {
            display: block;
            height: 100vh;
            width: 100vw;
            overflow: hidden;
        }

        .app {
            height: 100%;
            display: grid;
            grid-template-columns: 80px minmax(0, 1fr);
            background:
                radial-gradient(1200px 600px at 18% -10%, rgba(207, 217, 223, 0.12), transparent 60%),
                radial-gradient(900px 420px at 92% 12%, rgba(226, 235, 240, 0.08), transparent 65%),
                var(--color-bg-base, hsl(220, 25%, 8%));
        }

        .workspace {
            display: flex;
            flex-direction: column;
            height: 100%;
            min-height: 0;
            overflow: hidden;
        }
    `;

    @property()
    activeSection: SectionId = 'chat';

    @property()
    logoUrl = '';

    /**
     * emit a section-change event on nav clicks.
     */
    private _handleNavClick(e: CustomEvent<{ section: SectionId }>) {
        this.dispatchEvent(new CustomEvent('section-change', { detail: { section: e.detail.section } }));
    }

    /**
     * emit a section-dblclick event on nav double-clicks.
     */
    private _handleNavDblClick(e: CustomEvent<{ section: SectionId }>) {
        this.dispatchEvent(new CustomEvent('section-dblclick', { detail: { section: e.detail.section } }));
    }

    /**
     * render the layout skeleton with navigation and slot.
     */
    render() {
        return html`
            <div class="app">
                <wls-nav-rail
                    .activeSection=${this.activeSection}
                    .logoUrl=${this.logoUrl}
                    @nav-click=${this._handleNavClick}
                    @nav-dblclick=${this._handleNavDblClick}
                ></wls-nav-rail>
                
                <main class="workspace">
                    <slot></slot>
                </main>
            </div>
        `;
    }
}
