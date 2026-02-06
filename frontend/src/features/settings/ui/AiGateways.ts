/**
 * render provider connection settings and emit provider actions.
 * frontend/src/features/settings/ui/AiGateways.ts
 */

import { LitElement, html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { providers, providerBusy } from '../state/providerSignals';
import { catalogOverview, catalogBusy } from '../state/catalogSignals';
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
            display: flex;
            flex-direction: column;
            gap: 12px;
        }

        .empty-state {
            padding: 32px 24px;
            text-align: center;
            color: var(--color-text-muted);
            font-size: 14px;
            border: 1px dashed var(--color-border-subtle);
            border-radius: 10px;
            background: rgba(255, 255, 255, 0.02);
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

        .provider-endpoints {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }

        .endpoint-card {
            border: 1px solid var(--color-border-subtle);
            border-radius: 10px;
            padding: 12px;
            background: rgba(15, 18, 26, 0.5);
            display: flex;
            flex-direction: column;
            gap: 10px;
        }

        .endpoint-row {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
            flex-wrap: wrap;
        }

        .endpoint-title {
            font-weight: 600;
            display: flex;
            gap: 8px;
            align-items: center;
        }

        .endpoint-meta {
            font-size: 12px;
            color: var(--color-text-muted);
        }

        .endpoint-actions {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }

        .endpoint-models {
            display: flex;
            flex-direction: column;
            gap: 8px;
            font-size: 12px;
        }

        .model-table {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .model-table__header,
        .model-table__row {
            display: grid;
            grid-template-columns: minmax(220px, 1.4fr) minmax(90px, 0.6fr) minmax(120px, 1fr) minmax(120px, 1fr) minmax(140px, 1fr) minmax(160px, 1fr);
            gap: 12px;
            align-items: center;
        }

        .model-table__header {
            font-size: 11px;
            letter-spacing: 0.08em;
            text-transform: uppercase;
            color: var(--color-text-muted);
            padding: 0 4px;
        }

        .model-table__row {
            padding: 8px 10px;
            border-radius: 8px;
            border: 1px solid var(--color-border-subtle);
            background: rgba(10, 12, 18, 0.6);
        }

        .model-name {
            font-weight: 600;
        }

        .model-context {
            font-size: 11px;
            color: var(--color-text-muted);
            margin-top: 2px;
        }

        .model-cost {
            font-size: 12px;
            color: var(--color-text-secondary);
        }

        .model-pill-group {
            display: flex;
            flex-wrap: wrap;
            gap: 6px;
        }

        .model-table__empty {
            font-size: 12px;
            color: var(--color-text-muted);
        }

        .endpoint-tag {
            padding: 2px 6px;
            border-radius: 999px;
            border: 1px solid var(--color-border-subtle);
            font-size: 11px;
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
     * request an endpoint connectivity test.
     */
    private _handleTestEndpoint(endpointId: string) {
        this.dispatchEvent(new CustomEvent('catalog-endpoint-test', {
            bubbles: true,
            composed: true,
            detail: { endpointId },
        }));
    }

    /**
     * request an endpoint model refresh.
     */
    private _handleRefreshEndpoint(endpointId: string) {
        this.dispatchEvent(new CustomEvent('catalog-endpoint-refresh', {
            bubbles: true,
            composed: true,
            detail: { endpointId },
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
     * format the context window display value.
     */
    private _formatContextWindow(contextWindow?: number): string {
        if (!contextWindow || contextWindow <= 0) {
            return 'Context window: unknown';
        }
        return `Context window: ${contextWindow.toLocaleString()}`;
    }

    /**
     * resolve the display cost tier value.
     */
    private _resolveModelCostTier(costTier?: string): string {
        const value = costTier?.trim() ?? '';
        if (!value || value.trim() === '') {
            return 'unknown';
        }
        return value;
    }

    /**
     * resolve the capability labels for a model entry.
     */
    private _resolveModelCapabilities(model: {
        supportsStreaming: boolean;
        supportsToolCalling: boolean;
        supportsStructuredOutput: boolean;
        supportsVision: boolean;
    }): string[] {
        return [
            model.supportsStreaming ? 'streaming' : '',
            model.supportsToolCalling ? 'tools' : '',
            model.supportsStructuredOutput ? 'structured' : '',
            model.supportsVision ? 'vision' : '',
        ].filter(Boolean);
    }

    /**
     * add a role lookup entry without duplicates.
     */
    private _addRoleLookupEntry(roleLookup: Map<string, string[]>, key: string, roleName: string): void {
        const normalizedKey = key.trim();
        if (normalizedKey === '') {
            return;
        }
        const existing = roleLookup.get(normalizedKey) ?? [];
        if (existing.includes(roleName)) {
            return;
        }
        roleLookup.set(normalizedKey, [...existing, roleName]);
    }

    /**
     * resolve role names for a model catalog entry.
     */
    private _resolveAssignedRoles(modelCatalogEntryId: string, roleNamesByCatalogEntryId: Map<string, string[]>): string[] {
        const names = roleNamesByCatalogEntryId.get(modelCatalogEntryId) ?? [];
        return [...names].sort((a, b) => a.localeCompare(b));
    }

    /**
     * render the provider connections UI.
     */
    render() {
        const overview = catalogOverview.value;
        const endpoints = overview?.endpoints ?? [];
        const endpointGroups = new Map<string, typeof endpoints>();
        for (const endpoint of endpoints) {
            const list = endpointGroups.get(endpoint.providerName) ?? [];
            list.push(endpoint);
            endpointGroups.set(endpoint.providerName, list);
        }
        const roles = overview?.roles ?? [];
        const roleNamesByCatalogEntryId = new Map<string, string[]>();
        for (const role of roles) {
            const assignments = role.assignments ?? [];
            for (const assignment of assignments) {
                this._addRoleLookupEntry(roleNamesByCatalogEntryId, assignment.modelCatalogEntryId, role.name);
            }
        }
        const providerList = providers.value;
        const hasProviders = providerList.length > 0;

        return html`
            ${!hasProviders ? html`
                <div class="empty-state">
                    <p>No providers configured yet.</p>
                    <p>Add an API key to connect to an AI provider and start chatting.</p>
                </div>
            ` : nothing}
            ${hasProviders ? providerList.map(provider => {
            const endpointList = endpointGroups.get(provider.name) ?? [];
            const catalogModels = endpointList.flatMap((endpoint) => endpoint.models ?? []);
            const hasResources = catalogModels.length > 0;
            const isExpanded = !!this._expandedProviders[provider.name];
            const status = provider.status;
            const hasIssue = !!status && status.ok === false;
            const resourceMeta = hasResources
                ? `${catalogModels.length} available`
                : (provider.isConnected ? 'No resources found' : 'Connect to load');
            const modelCount = catalogModels.length;
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
                                    <div class="model-table">
                                        <div class="model-table__header">
                                            <span>Model</span>
                                            <span>Model cost</span>
                                            <span>Input</span>
                                            <span>Output</span>
                                            <span>Capabilities</span>
                                            <span>Role</span>
                                        </div>
                                        ${catalogModels.map(model => {
                    const inputModalities = model.inputModalities ?? [];
                    const outputModalities = model.outputModalities ?? [];
                    const capabilities = this._resolveModelCapabilities(model);
                    const roleNames = this._resolveAssignedRoles(model.id, roleNamesByCatalogEntryId);
                    return html`
                                                    <div class="model-table__row">
                                                        <div>
                                                            <div class="model-name">${model.displayName || model.modelId}</div>
                                                            <div class="model-context">${this._formatContextWindow(model.contextWindow)}</div>
                                                        </div>
                                                        <div class="model-cost">${this._resolveModelCostTier(model.costTier)}</div>
                                                        <div class="model-pill-group">
                                                            ${inputModalities.length === 0 ? html`<span class="model-table__empty">-</span>` : inputModalities.map(modality => html`<span class="endpoint-tag">${modality}</span>`)}
                                                        </div>
                                                        <div class="model-pill-group">
                                                            ${outputModalities.length === 0 ? html`<span class="model-table__empty">-</span>` : outputModalities.map(modality => html`<span class="endpoint-tag">${modality}</span>`)}
                                                        </div>
                                                        <div class="model-pill-group">
                                                            ${capabilities.length === 0 ? html`<span class="model-table__empty">-</span>` : capabilities.map(capability => html`<span class="endpoint-tag">${capability}</span>`)}
                                                        </div>
                                                        <div class="model-pill-group">
                                                            ${roleNames.length === 0 ? html`<span class="model-table__empty">-</span>` : roleNames.map(roleName => html`<span class="endpoint-tag">${roleName}</span>`)}
                                                        </div>
                                                    </div>
                                                `;
                })}
                                    </div>
                                </div>
                            ` : nothing}
                            ${endpointList.length > 0 ? html`
                                <div class="provider-endpoints">
                                    ${endpointList.map(endpoint => {
                const endpointBusy = !!catalogBusy.value[endpoint.id];
                const endpointModels = endpoint.models ?? [];
                return html`
                                                    <div class="endpoint-card">
                                                        <div class="endpoint-row">
                                                            <div>
                                                                <div class="endpoint-title">${endpoint.displayName}</div>
                                                        <div class="endpoint-meta">
                                                            ${endpoint.routeKind}${endpoint.originProvider ? ` · ${endpoint.originProvider}` : ''}${endpoint.originRouteLabel ? ` · ${endpoint.originRouteLabel}` : ''}
                                                        </div>
                                                    </div>
                                                    <div class="endpoint-actions">
                                                        <button
                                                            class="btn"
                                                            @click=${() => this._handleTestEndpoint(endpoint.id)}
                                                            ?disabled=${endpointBusy}
                                                        >
                                                            ${endpointBusy ? 'Testing...' : 'Test'}
                                                        </button>
                                                        <button
                                                            class="btn"
                                                            @click=${() => this._handleRefreshEndpoint(endpoint.id)}
                                                            ?disabled=${endpointBusy}
                                                        >
                                                            ${endpointBusy ? 'Refreshing...' : 'Refresh'}
                                                        </button>
                                                    </div>
                                                        </div>
                                                        <div class="endpoint-models">
                                                            ${endpointModels.length === 0 ? html`
                                                                <span class="provider-models">No models discovered yet.</span>
                                                            ` : html`
                                                                <div class="model-table">
                                                                    <div class="model-table__header">
                                                                        <span>Model</span>
                                                                        <span>Cost tier</span>
                                                                        <span>Input</span>
                                                                        <span>Output</span>
                                                                        <span>Capabilities</span>
                                                                        <span>Role</span>
                                                                    </div>
                                                                    ${endpointModels.map(model => {
                    const inputModalities = model.inputModalities ?? [];
                    const outputModalities = model.outputModalities ?? [];
                    const capabilities = this._resolveModelCapabilities(model);
                    const costTier = this._resolveModelCostTier(model.costTier);
                    const roleNames = this._resolveAssignedRoles(model.id, roleNamesByCatalogEntryId);
                    return html`
                                                                            <div class="model-table__row">
                                                                                <div>
                                                                                    <div class="model-name">${model.displayName || model.modelId}</div>
                                                                                    <div class="model-context">${this._formatContextWindow(model.contextWindow)}</div>
                                                                                </div>
                                                                                <div class="model-cost">${costTier}</div>
                                                                                <div class="model-pill-group">
                                                                                    ${inputModalities.length === 0 ? html`<span class="model-table__empty">—</span>` : inputModalities.map(modality => html`<span class="endpoint-tag">${modality}</span>`)}
                                                                                </div>
                                                                                <div class="model-pill-group">
                                                                                    ${outputModalities.length === 0 ? html`<span class="model-table__empty">—</span>` : outputModalities.map(modality => html`<span class="endpoint-tag">${modality}</span>`)}
                                                                                </div>
                                                                                <div class="model-pill-group">
                                                                                    ${capabilities.length === 0 ? html`<span class="model-table__empty">—</span>` : capabilities.map(capability => html`<span class="endpoint-tag">${capability}</span>`)}
                                                                                </div>
                                                                                <div class="model-pill-group">
                                                                                    ${roleNames.length === 0 ? html`<span class="model-table__empty">-</span>` : roleNames.map(roleName => html`<span class="endpoint-tag">${roleName}</span>`)}
                                                                                </div>
                                                                            </div>
                                                                        `;
                })}
                                                                </div>
                                                            `}
                                                        </div>
                                                    </div>
                                                `;
            })}
                                </div>
                            ` : nothing}
                        </div>
                    `;
        }) : nothing}
        `;
    }
}
