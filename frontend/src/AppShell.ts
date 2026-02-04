/**
 * AppShell.ts orchestrates section routing and side-panel behavior.
 * frontend/src/AppShell.ts
 */

import { LitElement, html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';

import { SectionId } from './types/shell';
import './shell/MainLayout';
import './features/chat';
import './features/settings/SettingsView';
import './features/chat/ChatSidePanel';
import './features/settings/SettingsSidePanel';
import './components/ToastStack';

import logoUrl from './assets/images/wls-chatbot-logo.png';

type SectionSidePanelSpec = {
    hasSidePanel: boolean;
    defaultOpen: boolean;
};

const sectionSidePanelSpecs: Record<SectionId, SectionSidePanelSpec> = {
    chat: {
        hasSidePanel: true,
        defaultOpen: true,
    },
    settings: {
        hasSidePanel: true,
        defaultOpen: true,
    },
};

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

    @state()
    private _sidePanelOpen = sectionSidePanelSpecs.chat.defaultOpen;

    @state()
    private _sidePanelOverrides: Partial<Record<SectionId, boolean>> = {};

    /**
     * report whether a section has side panel content.
     */
    private _hasSidePanel(section: SectionId) {
        return sectionSidePanelSpecs[section].hasSidePanel;
    }

    /**
     * resolve the side panel open state for a section based on defaults and overrides.
     */
    private _resolveSidePanelOpen(section: SectionId) {
        if (!this._hasSidePanel(section)) {
            return false;
        }
        const override = this._sidePanelOverrides[section];
        if (override !== undefined) {
            return override;
        }
        return sectionSidePanelSpecs[section].defaultOpen;
    }

    /**
     * toggle side panel visibility for the active section.
     */
    private _toggleActiveSectionSidePanel() {
        if (!this._hasSidePanel(this._activeSection)) {
            this._sidePanelOpen = false;
            return;
        }
        const nextOpen = !this._sidePanelOpen;
        this._sidePanelOpen = nextOpen;
        if (this._sidePanelOverrides[this._activeSection] !== undefined) {
            this._sidePanelOverrides = {
                ...this._sidePanelOverrides,
                [this._activeSection]: nextOpen,
            };
        }
    }

    /**
     * switch the active section and apply side panel defaults.
     */
    private _handleSectionChange(e: CustomEvent<{ section: SectionId }>) {
        const { section } = e.detail;
        if (this._activeSection === section) {
            this._toggleActiveSectionSidePanel();
            return;
        }
        this._activeSection = section;
        this._sidePanelOpen = this._resolveSidePanelOpen(section);
    }

    /**
     * toggle side panel visibility and persist that state for section switches.
     */
    private _handleSectionDoubleClick(e: CustomEvent<{ section: SectionId }>) {
        const { section } = e.detail;
        if (!this._hasSidePanel(section)) {
            this._activeSection = section;
            this._sidePanelOpen = false;
            return;
        }

        const currentOpen = section === this._activeSection
            ? this._sidePanelOpen
            : this._resolveSidePanelOpen(section);
        const nextOpen = !currentOpen;

        this._activeSection = section;
        this._sidePanelOpen = nextOpen;
        this._sidePanelOverrides = {
            ...this._sidePanelOverrides,
            [section]: nextOpen,
        };
    }

    /**
     * render the shell layout and active workspace content.
     */
    render() {
        return html`
            <wls-main-layout
                .activeSection=${this._activeSection}
                .logoUrl=${logoUrl}
                .hasSidePanel=${this._hasSidePanel(this._activeSection)}
                .sidePanelOpen=${this._sidePanelOpen}
                @section-change=${this._handleSectionChange}
                @section-dblclick=${this._handleSectionDoubleClick}
            >
                ${this._renderSidePanel()}
                ${this._renderContent()}
            </wls-main-layout>
            <wls-toast-stack></wls-toast-stack>
        `;
    }

    /**
     * choose side panel content for the active section.
     */
    private _renderSidePanel() {
        if (!this._hasSidePanel(this._activeSection)) {
            return nothing;
        }

        switch (this._activeSection) {
            case 'chat':
                return html`<wls-chat-side-panel slot="side-panel"></wls-chat-side-panel>`;
            case 'settings':
                return html`<wls-settings-side-panel slot="side-panel"></wls-settings-side-panel>`;
            default:
                return nothing;
        }
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
