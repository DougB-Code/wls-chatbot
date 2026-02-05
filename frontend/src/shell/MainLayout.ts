/**
 * MainLayout.ts renders shell grid regions for nav, side panel, header, and workspace.
 * frontend/src/shell/MainLayout.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import './NavRail';
import { SectionId } from '../types/shell';

/**
 * render the nav rail, header slot, and workspace layout.
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
            --side-panel-width: 0px;
            --side-panel-open-width: clamp(220px, 30vw, 300px);
            height: 100%;
            display: grid;
            grid-template-columns: 80px auto minmax(0, 1fr);
            background:
                radial-gradient(1200px 600px at 18% -10%, rgba(207, 217, 223, 0.12), transparent 60%),
                radial-gradient(900px 420px at 92% 12%, rgba(226, 235, 240, 0.08), transparent 65%),
                var(--color-bg-base, hsl(220, 25%, 8%));
        }

        .side-panel {
            height: 100%;
            min-height: 0;
            width: var(--side-panel-width);
            overflow: hidden;
            border-right: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.06));
            background: rgba(12, 19, 30, 0.92);
            transition:
                width 160ms var(--ease-out, ease-out),
                border-color 120ms var(--ease-out, ease-out);
        }

        .side-panel--closed {
            border-right-color: transparent;
            pointer-events: none;
        }

        .side-panel__content {
            height: 100%;
            min-height: 0;
            width: var(--side-panel-open-width);
            overflow: hidden;
        }

        .workspace {
            display: flex;
            flex-direction: column;
            height: 100%;
            min-height: 0;
            overflow: hidden;
        }

        .workspace__header {
            flex-shrink: 0;
        }

        .workspace__content {
            flex: 1;
            min-height: 0;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
    `;

    @property()
    activeSection: SectionId = 'chat';

    @property()
    logoUrl = '';

    @property({ type: Boolean })
    hasSidePanel = false;

    @property({ type: Boolean })
    sidePanelOpen = false;

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
        const sidePanelWidth = this.hasSidePanel && this.sidePanelOpen ? 'var(--side-panel-open-width)' : '0px';
        return html`
            <div class="app" style=${`--side-panel-width: ${sidePanelWidth};`}>
                <wls-nav-rail
                    .activeSection=${this.activeSection}
                    .logoUrl=${this.logoUrl}
                    @nav-click=${this._handleNavClick}
                    @nav-dblclick=${this._handleNavDblClick}
                ></wls-nav-rail>

                <aside
                    class="side-panel ${this.hasSidePanel && this.sidePanelOpen ? 'side-panel--open' : 'side-panel--closed'}"
                    aria-hidden=${String(!this.hasSidePanel || !this.sidePanelOpen)}
                >
                    <div class="side-panel__content">
                        <slot name="side-panel"></slot>
                    </div>
                </aside>

                <main class="workspace">
                    <div class="workspace__header">
                        <slot name="header"></slot>
                    </div>
                    <div class="workspace__content">
                        <slot></slot>
                    </div>
                </main>
            </div>
        `;
    }
}
