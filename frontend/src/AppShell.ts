/**
 * define the app shell that wires top-level layout and navigation.
 */

import { LitElement, html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';

import { SectionId } from './types/shell';
import './shell/MainLayout';
import './features/chat';
import './features/settings/SettingsView';
import './components/ToastStack';

import logoUrl from './assets/images/wls-chatbot-logo.png';

/**
 * host the main layout and switch between workspace views.
 */
@customElement('wls-app-shell')
export class AppShell extends LitElement {
    static styles = css`
        :host {
            display: block;
            height: 100%;
        }
    `;

    @state()
    private _activeSection: SectionId = 'chat';

    /**
     * update the active workspace section on navigation changes.
     */
    private _handleSectionChange(e: CustomEvent<{ section: SectionId }>) {
        const { section } = e.detail;
        if (this._activeSection !== section) {
            this._activeSection = section;
        }
    }

    /**
     * render the shell layout and active workspace content.
     */
    render() {
        return html`
            <wls-main-layout
                .activeSection=${this._activeSection}
                .logoUrl=${logoUrl}
                @section-change=${this._handleSectionChange}
            >
                ${this._renderContent()}
            </wls-main-layout>
            <wls-toast-stack></wls-toast-stack>
        `;
    }

    /**
     * choose the workspace view based on the active section.
     */
    private _renderContent() {
        switch (this._activeSection) {
            case 'chat':
                return html`<wls-chat-view></wls-chat-view>`;

            case 'settings':
                return html`<wls-settings-view></wls-settings-view>`;
            default:
                return nothing;
        }
    }
}
