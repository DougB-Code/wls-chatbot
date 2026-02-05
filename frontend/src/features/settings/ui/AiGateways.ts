/**
 * render provider connection settings and emit provider actions.
 * frontend/src/features/settings/ui/AiGateways.ts
 */

import { LitElement, html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { providers, providerBusy } from '../state/providerSignals';
import type { ProviderInfo } from '../../../types/wails';
import { settingsSharedStyles } from './settingsStyles';

/**
 * display and manage AI provider connections for the settings page.
 */
@customElement('wls-connections-view')
export class ConnectionsView extends SignalWatcher(LitElement) {
    static styles = [
        settingsSharedStyles,
        css`
        :host {
            display: block;
            height: 100%;
            overflow-y: auto;
        }

        .settings {
            display: flex;
            flex-direction: column;
            gap: 24px;
            flex: 1;
            min-height: 0;
        }

        .provider-list {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }

        .provider-item {
            display: flex;
            flex-direction: column;
            align-items: stretch;
            gap: 12px;
            padding: 14px 16px;
            border-radius: 10px;
            border: 1px solid var(--color-border-subtle);
            background: var(--color-bg-elevated);
        }

        .provider-item.active {
            border-color: var(--color-user-border);
            background: var(--color-user-surface);
        }

        .provider-row {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 16px;
            flex-wrap: wrap;
        }

        .provider-info {
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            gap: 4px;
        }

        .provider-name {
            font-weight: 600;
        }

        .provider-models {
            font-size: 12px;
            color: var(--color-text-muted);
        }

        .provider-actions {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
            align-items: center;
        }

        .provider-inputs {
            display: flex;
            flex-direction: column;
            gap: 8px;
            min-width: 220px;
        }

        .provider-input {
            display: flex;
            flex-direction: column;
            gap: 6px;
        }

        .provider-label {
            font-size: 11px;
            letter-spacing: 0.04em;
            text-transform: uppercase;
            color: var(--color-text-muted);
        }

        .provider-help {
            font-size: 11px;
            color: var(--color-text-muted);
        }

        .provider-status {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 6px 10px;
            border-radius: 8px;
            font-size: 12px;
            color: var(--color-text-secondary);
            border: 1px solid var(--color-border-subtle);
            background: rgba(255, 255, 255, 0.02);
        }

        .provider-status--error {
            color: var(--color-error);
            border-color: var(--color-error-border);
            background: var(--color-error-surface);
        }

        .provider-status__dot {
            width: 8px;
            height: 8px;
            border-radius: 999px;
            background: currentColor;
            box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.2);
        }

        .provider-status__text {
            color: inherit;
        }

        .input {
            padding: 10px 14px;
            border-radius: 8px;
            border: 1px solid var(--color-border-default);
            background: var(--color-bg-elevated);
            color: var(--color-text-primary);
            font-size: 14px;
            outline: none;
            width: 200px;
        }

        .input::placeholder {
            color: var(--color-text-muted);
        }

        .btn {
            padding: 8px 14px;
            border-radius: 8px;
            border: 1px solid var(--color-border-default);
            background: transparent;
            color: var(--color-text-secondary);
            font-size: 13px;
            cursor: pointer;
            transition: all 120ms ease-out;
        }

        .btn:hover {
            background: var(--color-interactive-hover);
            color: var(--color-text-primary);
        }

        .btn--primary {
            background: linear-gradient(135deg, var(--color-user) 0%, hsl(235, 60%, 55%) 100%);
            border-color: transparent;
            color: white;
        }

        .btn--danger {
            background: transparent;
            border-color: rgba(239, 68, 68, 0.3);
            color: #ef4444;
        }

        .btn--danger:hover {
            background: rgba(239, 68, 68, 0.1);
            border-color: #ef4444;
            color: #ef4444;
        }

        .provider-resources__toggle {
            display: grid;
            grid-template-columns: auto 1fr auto;
            align-items: center;
            gap: 10px;
            width: 100%;
            padding: 10px 12px;
            border-radius: 8px;
            border: 1px solid var(--color-border-subtle);
            background: rgba(255, 255, 255, 0.02);
            color: var(--color-text-secondary);
            font-size: 12px;
            letter-spacing: 0.02em;
            text-transform: uppercase;
            cursor: pointer;
            transition: all 120ms ease-out;
        }

        .provider-resources__toggle:hover {
            background: var(--color-interactive-hover);
            color: var(--color-text-primary);
        }

        .provider-resources__toggle:disabled {
            cursor: not-allowed;
            opacity: 0.6;
        }

        .provider-resources__label {
            font-weight: 600;
        }

        .provider-resources__meta {
            text-transform: none;
            font-size: 12px;
            color: var(--color-text-muted);
        }

        .provider-resources__chevron {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            width: 20px;
            height: 20px;
            border-radius: 6px;
            background: rgba(255, 255, 255, 0.06);
            transition: transform 120ms ease-out;
        }

        .provider-resources__chevron svg {
            width: 12px;
            height: 12px;
        }

        .provider-resources__chevron.open {
            transform: rotate(180deg);
        }

        .provider-resources {
            border-top: 1px solid var(--color-border-subtle);
            padding-top: 10px;
        }

        .provider-resources__list {
            display: flex;
            flex-wrap: wrap;
            gap: 8px;
        }

        .provider-resource {
            padding: 6px 10px;
            border-radius: 999px;
            border: 1px solid var(--color-border-subtle);
            background: rgba(255, 255, 255, 0.04);
            font-size: 12px;
            color: var(--color-text-secondary);
        }
    `,
    ];

    @state()
    private _credentialInputs: Record<string, Record<string, string>> = {};

    @state()
    private _expandedProviders: Record<string, boolean> = {};

    /**
     * update the credential input state for a provider field.
     */
    private _handleCredentialInput(providerName: string, fieldName: string, value: string) {
        const providerInputs = this._credentialInputs[providerName] || {};
        this._credentialInputs = {
            ...this._credentialInputs,
            [providerName]: {
                ...providerInputs,
                [fieldName]: value,
            },
        };
    }

    /**
     * emit a provider connect/configure action with credentials.
     */
    private _handleConfigureProvider(provider: ProviderInfo, action: 'provider-connect' | 'provider-configure') {
        const credentials = this._collectCredentials(provider);
        if (Object.keys(credentials).length === 0) return;

        this.dispatchEvent(new CustomEvent(action, {
            bubbles: true,
            composed: true,
            detail: { name: provider.name, credentials },
        }));

        this._credentialInputs = { ...this._credentialInputs, [provider.name]: {} };
    }

    /**
     * collect credential values from user input and stored non-secret values.
     */
    private _collectCredentials(provider: ProviderInfo): Record<string, string> {
        const fields = provider.credentialFields ?? [];
        const storedValues = provider.credentialValues ?? {};
        const draftValues = this._credentialInputs[provider.name] ?? {};
        const credentials: Record<string, string> = {};

        for (const field of fields) {
            const draftValue = draftValues[field.name];
            if (draftValue && draftValue.trim() !== '') {
                credentials[field.name] = draftValue.trim();
                continue;
            }

            if (!field.secret) {
                const storedValue = storedValues[field.name];
                if (storedValue && storedValue.trim() !== '') {
                    credentials[field.name] = storedValue.trim();
                }
            }
        }

        return credentials;
    }

    /**
     * resolve the displayed input value for a provider field.
     */
    private _resolveCredentialValue(provider: ProviderInfo, field: { name: string; secret?: boolean }): string {
        const draftValue = this._credentialInputs[provider.name]?.[field.name];
        if (draftValue !== undefined) {
            return draftValue;
        }
        if (!field.secret) {
            return provider.credentialValues?.[field.name] ?? '';
        }
        return '';
    }

    /**
     * request a provider disconnect after user confirmation.
     */
    private _handleDisconnectProvider(name: string) {
        if (!confirm('Are you sure you want to disconnect this provider? This will remove stored credentials.')) {
            return;
        }

        this.dispatchEvent(new CustomEvent('provider-disconnect', {
            bubbles: true,
            composed: true,
            detail: { name },
        }));
    }

    /**
     * request a provider resource refresh.
     */
    private _handleRefreshProvider(name: string) {
        this.dispatchEvent(new CustomEvent('provider-refresh', {
            bubbles: true,
            composed: true,
            detail: { name },
        }));
    }

    /**
     * expand or collapse the provider resource list.
     */
    private _toggleProviderResources(name: string) {
        this._expandedProviders = {
            ...this._expandedProviders,
            [name]: !this._expandedProviders[name],
        };
    }

    /**
     * render the provider connections UI.
     */
    render() {
        return html`
            <div class="settings">
                <header class="settings__header">
                    <h1 class="settings__title">Provider Connections</h1>
                    <p class="settings__subtitle">Configure your AI provider credentials to enable chat.</p>
                </header>

                <div class="card">
                    <h2 class="card__title">Available Providers</h2>
                    <div class="provider-list">
                        ${providers.value.map(provider => {
            const resources = provider.resources ?? [];
            const hasResources = resources.length > 0;
            const isExpanded = !!this._expandedProviders[provider.name];
            const status = provider.status;
            const hasIssue = !!status && status.ok === false;
            const resourceMeta = hasResources
                ? `${resources.length} available`
                : (provider.isConnected ? 'No resources found' : 'Connect to load');
            const modelCount = hasResources ? resources.length : provider.models.length;
            const isBusy = !!providerBusy.value[provider.name];
            const fields = provider.credentialFields ?? [];

            return html`
                                <div class="provider-item ${provider.isActive ? 'active' : ''}">
                                    <div class="provider-row">
                                        <div class="provider-info">
                                            <span class="provider-name">${provider.displayName}</span>
                                            <span class="provider-models">${modelCount} models</span>
                                        </div>
                                        <div class="provider-actions">
                                            ${fields.length > 0 ? html`
                                                <div class="provider-inputs">
                                                    ${fields.map((field) => html`
                                                        <label class="provider-input">
                                                            <span class="provider-label">${field.label}${field.required ? ' *' : ''}</span>
                                                            <input
                                                                class="input"
                                                                type=${field.secret ? 'password' : 'text'}
                                                                placeholder=${field.placeholder || field.label}
                                                                .value=${this._resolveCredentialValue(provider, field)}
                                                                @input=${(e: Event) => this._handleCredentialInput(
                                                                    provider.name,
                                                                    field.name,
                                                                    (e.target as HTMLInputElement).value
                                                                )}
                                                            />
                                                            ${field.help ? html`<span class="provider-help">${field.help}</span>` : nothing}
                                                        </label>
                                                    `)}
                                                </div>
                                            ` : nothing}
                                            ${!provider.isConnected ? html`
                                                <button
                                                    class="btn btn--primary"
                                                    @click=${() => this._handleConfigureProvider(provider, 'provider-connect')}
                                                    ?disabled=${isBusy}
                                                >
                                                    ${isBusy ? 'Connecting...' : 'Connect'}
                                                </button>
                                            ` : html`
                                                 <button
                                                    class="btn"
                                                    @click=${() => this._handleConfigureProvider(provider, 'provider-configure')}
                                                    ?disabled=${isBusy}
                                                >
                                                    ${isBusy ? 'Updating...' : 'Update'}
                                                </button>
                                                <button
                                                    class="btn"
                                                    @click=${() => this._handleRefreshProvider(provider.name)}
                                                    ?disabled=${isBusy}
                                                    title="Refresh available models"
                                                >
                                                     ${isBusy ? 'Refreshing...' : 'Refresh'}
                                                </button>
                                                <button
                                                    class="btn btn--danger"
                                                    @click=${() => this._handleDisconnectProvider(provider.name)}
                                                    ?disabled=${isBusy}
                                                    title="Disconnect and remove stored credentials"
                                                >
                                                    Disconnect
                                                </button>
                                            `}
                                        </div>
                                    </div>
                                    ${hasIssue ? html`
                                        <div class="provider-status provider-status--error">
                                            <span class="provider-status__dot" aria-hidden="true"></span>
                                            <span class="provider-status__text">
                                                ${status?.message || 'Provider entitlement check failed.'}
                                            </span>
                                        </div>
                                    ` : nothing}
                                    <button
                                        class="provider-resources__toggle"
                                        ?disabled=${!hasResources}
                                        @click=${() => this._toggleProviderResources(provider.name)}
                                    >
                                        <span class="provider-resources__label">Resources</span>
                                        <span class="provider-resources__meta">${resourceMeta}</span>
                                        <span class="provider-resources__chevron ${isExpanded ? 'open' : ''}">
                                             <svg viewBox="0 0 12 12" aria-hidden="true">
                                                 <path d="M3 4l3 3 3-3" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"></path>
                                             </svg>
                                        </span>
                                    </button>
                                    ${hasResources && isExpanded ? html`
                                        <div class="provider-resources">
                                            <div class="provider-resources__list">
                                                ${resources.map(resource => html`
                                                    <span class="provider-resource">${resource.name || resource.id}</span>
                                                `)}
                                            </div>
                                        </div>
                                    ` : nothing}
                                </div>
                            `;
        })}
                    </div>
                </div>
            </div>
        `;
    }
}
