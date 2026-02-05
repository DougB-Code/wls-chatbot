/**
 * CatalogRoles.ts renders role definitions and model assignments.
 * frontend/src/features/settings/ui/CatalogRoles.ts
 */

import { LitElement, html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { catalogOverview, catalogBusy, catalogError } from '../state/catalogSignals';
import { settingsSharedStyles } from './settingsStyles';
import type { RoleSummary } from '../../../types/catalog';

const modalities = ['text', 'image', 'audio', 'video'];

/**
 * display and manage role assignments for catalog models.
 */
@customElement('wls-catalog-roles')
export class CatalogRoles extends SignalWatcher(LitElement) {
    static styles = [
        settingsSharedStyles,
        css`
        :host {
            display: block;
        }

        .roles-card {
            display: flex;
            flex-direction: column;
            gap: 16px;
        }

        .roles-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            gap: 16px;
            flex-wrap: wrap;
        }

        .roles-table {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }

        .role-item {
            border: 1px solid var(--color-border-subtle);
            border-radius: 10px;
            padding: 14px;
            background: var(--color-bg-elevated);
            display: flex;
            flex-direction: column;
            gap: 12px;
        }

        .role-row {
            display: flex;
            justify-content: space-between;
            gap: 16px;
            flex-wrap: wrap;
        }

        .role-name {
            font-weight: 600;
        }

        .role-meta {
            font-size: 12px;
            color: var(--color-text-muted);
            display: flex;
            flex-wrap: wrap;
            gap: 6px;
        }

        .role-tag {
            padding: 2px 6px;
            border-radius: 999px;
            border: 1px solid var(--color-border-subtle);
            font-size: 11px;
        }

        .role-actions {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }

        .role-form {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: 12px;
            align-items: end;
        }

        .role-form label {
            display: flex;
            flex-direction: column;
            gap: 6px;
            font-size: 12px;
            color: var(--color-text-muted);
        }

        .checkbox-group {
            display: flex;
            flex-wrap: wrap;
            gap: 8px;
        }

        .checkbox-group label {
            flex-direction: row;
            align-items: center;
            gap: 6px;
            font-size: 12px;
            color: var(--color-text-primary);
        }

        .assignment-list {
            display: flex;
            flex-direction: column;
            gap: 6px;
            font-size: 12px;
        }

        .assignment-item {
            display: flex;
            justify-content: space-between;
            gap: 12px;
            align-items: center;
            padding: 6px 8px;
            border-radius: 8px;
            border: 1px solid var(--color-border-subtle);
            background: rgba(12, 15, 22, 0.6);
        }

        .assignment-label {
            color: var(--color-text-secondary);
        }

        .assignment-form {
            display: flex;
            flex-wrap: wrap;
            gap: 8px;
            align-items: center;
        }

        .select {
            padding: 8px 12px;
            border-radius: 8px;
            border: 1px solid var(--color-border-default);
            background: var(--color-bg-elevated);
            color: var(--color-text-primary);
            font-size: 13px;
        }

        .error-text {
            color: var(--color-error);
            font-size: 12px;
        }
        `,
    ];

    @state()
    private _newRole: RoleSummary = this._emptyRole();

    @state()
    private _selectedModels: Record<string, string> = {};

    private _emptyRole(): RoleSummary {
        return {
            id: '',
            name: '',
            requirements: {
                requiredInputModalities: ['text'],
                requiredOutputModalities: ['text'],
                requiresStreaming: false,
                requiresToolCalling: false,
                requiresStructuredOutput: false,
                requiresVision: false,
            },
            constraints: {},
            assignments: [],
        };
    }

    private _handleRoleInput(field: keyof RoleSummary, value: string) {
        this._newRole = { ...this._newRole, [field]: value } as RoleSummary;
    }

    private _toggleRequirement(key: keyof RoleSummary['requirements'], value: boolean) {
        this._newRole = {
            ...this._newRole,
            requirements: {
                ...this._newRole.requirements,
                [key]: value,
            },
        };
    }

    private _toggleModality(listKey: 'requiredInputModalities' | 'requiredOutputModalities', modality: string) {
        const current = new Set(this._newRole.requirements[listKey]);
        if (current.has(modality)) {
            current.delete(modality);
        } else {
            current.add(modality);
        }
        this._newRole = {
            ...this._newRole,
            requirements: {
                ...this._newRole.requirements,
                [listKey]: Array.from(current),
            },
        };
    }

    private _handleConstraintInput(key: keyof RoleSummary['constraints'], value: string) {
        this._newRole = {
            ...this._newRole,
            constraints: {
                ...this._newRole.constraints,
                [key]: value || undefined,
            },
        };
    }

    private _submitRole(): void {
        if (!this._newRole.name.trim()) {
            return;
        }
        this.dispatchEvent(new CustomEvent('catalog-role-save', {
            bubbles: true,
            composed: true,
            detail: { role: this._newRole },
        }));
        this._newRole = this._emptyRole();
    }

    private _deleteRole(roleId: string): void {
        if (!confirm('Delete this role?')) {
            return;
        }
        this.dispatchEvent(new CustomEvent('catalog-role-delete', {
            bubbles: true,
            composed: true,
            detail: { roleId },
        }));
    }

    private _assignRole(roleId: string): void {
        const modelEntryId = this._selectedModels[roleId];
        if (!modelEntryId) {
            return;
        }
        this.dispatchEvent(new CustomEvent('catalog-role-assign', {
            bubbles: true,
            composed: true,
            detail: { roleId, modelEntryId },
        }));
        this._selectedModels = { ...this._selectedModels, [roleId]: '' };
    }

    private _unassignRole(roleId: string, modelEntryId: string): void {
        this.dispatchEvent(new CustomEvent('catalog-role-unassign', {
            bubbles: true,
            composed: true,
            detail: { roleId, modelEntryId },
        }));
    }

    private _availableModels() {
        const overview = catalogOverview.value;
        if (!overview) {
            return [] as { id: string; label: string }[];
        }
        const models = overview.endpoints.flatMap((endpoint) =>
            endpoint.models.map((model) => ({
                id: model.id,
                label: `${endpoint.providerDisplayName} · ${endpoint.displayName} · ${model.modelId}`,
            })),
        );
        models.sort((a, b) => a.label.localeCompare(b.label));
        return models;
    }

    render() {
        const overview = catalogOverview.value;
        const roles = overview?.roles ?? [];
        const availableModels = this._availableModels();
        const error = catalogError.value;

        return html`
            <div class="card roles-card">
                <div class="roles-header">
                    <div>
                        <h2 class="card__title">Roles & Model Assignments</h2>
                        <p class="settings__subtitle">Define role requirements and map compatible models.</p>
                    </div>
                    ${error ? html`<span class="error-text">${error}</span>` : nothing}
                </div>

                <div class="role-form">
                    <label>
                        Role name
                        <input
                            class="input"
                            type="text"
                            .value=${this._newRole.name}
                            @input=${(event: Event) => this._handleRoleInput('name', (event.target as HTMLInputElement).value)}
                        />
                    </label>
                    <label>
                        Max cost tier
                        <input
                            class="input"
                            type="text"
                            placeholder="cheap | standard | expensive"
                            .value=${this._newRole.constraints.maxCostTier ?? ''}
                            @input=${(event: Event) => this._handleConstraintInput('maxCostTier', (event.target as HTMLInputElement).value)}
                        />
                    </label>
                    <label>
                        Max latency tier
                        <input
                            class="input"
                            type="text"
                            placeholder="fast | standard | slow"
                            .value=${this._newRole.constraints.maxLatencyTier ?? ''}
                            @input=${(event: Event) => this._handleConstraintInput('maxLatencyTier', (event.target as HTMLInputElement).value)}
                        />
                    </label>
                    <label>
                        Min reliability tier
                        <input
                            class="input"
                            type="text"
                            placeholder="stable | preview"
                            .value=${this._newRole.constraints.minReliabilityTier ?? ''}
                            @input=${(event: Event) => this._handleConstraintInput('minReliabilityTier', (event.target as HTMLInputElement).value)}
                        />
                    </label>
                    <div>
                        <span class="provider-label">Input modalities</span>
                        <div class="checkbox-group">
                            ${modalities.map((modality) => html`
                                <label>
                                    <input
                                        type="checkbox"
                                        .checked=${this._newRole.requirements.requiredInputModalities.includes(modality)}
                                        @change=${() => this._toggleModality('requiredInputModalities', modality)}
                                    />
                                    ${modality}
                                </label>
                            `)}
                        </div>
                    </div>
                    <div>
                        <span class="provider-label">Output modalities</span>
                        <div class="checkbox-group">
                            ${modalities.map((modality) => html`
                                <label>
                                    <input
                                        type="checkbox"
                                        .checked=${this._newRole.requirements.requiredOutputModalities.includes(modality)}
                                        @change=${() => this._toggleModality('requiredOutputModalities', modality)}
                                    />
                                    ${modality}
                                </label>
                            `)}
                        </div>
                    </div>
                    <div>
                        <span class="provider-label">Required features</span>
                        <div class="checkbox-group">
                            <label>
                                <input
                                    type="checkbox"
                                    .checked=${this._newRole.requirements.requiresStreaming}
                                    @change=${(event: Event) => this._toggleRequirement('requiresStreaming', (event.target as HTMLInputElement).checked)}
                                />
                                streaming
                            </label>
                            <label>
                                <input
                                    type="checkbox"
                                    .checked=${this._newRole.requirements.requiresToolCalling}
                                    @change=${(event: Event) => this._toggleRequirement('requiresToolCalling', (event.target as HTMLInputElement).checked)}
                                />
                                tool calling
                            </label>
                            <label>
                                <input
                                    type="checkbox"
                                    .checked=${this._newRole.requirements.requiresStructuredOutput}
                                    @change=${(event: Event) => this._toggleRequirement('requiresStructuredOutput', (event.target as HTMLInputElement).checked)}
                                />
                                structured output
                            </label>
                            <label>
                                <input
                                    type="checkbox"
                                    .checked=${this._newRole.requirements.requiresVision}
                                    @change=${(event: Event) => this._toggleRequirement('requiresVision', (event.target as HTMLInputElement).checked)}
                                />
                                vision
                            </label>
                        </div>
                    </div>
                    <button class="btn btn--primary" @click=${this._submitRole.bind(this)}>
                        Add role
                    </button>
                </div>

                <div class="roles-table">
                    ${roles.length === 0 ? html`
                        <span class="provider-models">No roles defined yet.</span>
                    ` : roles.map((role) => html`
                        <div class="role-item">
                            <div class="role-row">
                                <div>
                                    <div class="role-name">${role.name}</div>
                                    <div class="role-meta">
                                        ${role.requirements.requiredInputModalities.map((modality) => html`<span class="role-tag">in:${modality}</span>`)}
                                        ${role.requirements.requiredOutputModalities.map((modality) => html`<span class="role-tag">out:${modality}</span>`)}
                                        ${role.requirements.requiresStreaming ? html`<span class="role-tag">streaming</span>` : nothing}
                                        ${role.requirements.requiresToolCalling ? html`<span class="role-tag">tools</span>` : nothing}
                                        ${role.requirements.requiresStructuredOutput ? html`<span class="role-tag">structured</span>` : nothing}
                                        ${role.requirements.requiresVision ? html`<span class="role-tag">vision</span>` : nothing}
                                    </div>
                                </div>
                                <div class="role-actions">
                                    <button class="btn btn--danger" @click=${() => this._deleteRole(role.id)}>
                                        Delete
                                    </button>
                                </div>
                            </div>
                            <div class="assignment-form">
                                <select
                                    class="select"
                                    .value=${this._selectedModels[role.id] ?? ''}
                                    @change=${(event: Event) => {
                                        const value = (event.target as HTMLSelectElement).value;
                                        this._selectedModels = { ...this._selectedModels, [role.id]: value };
                                    }}
                                >
                                    <option value="">Select model to assign</option>
                                    ${availableModels.map((model) => html`
                                        <option value=${model.id}>${model.label}</option>
                                    `)}
                                </select>
                                <button
                                    class="btn"
                                    @click=${() => this._assignRole(role.id)}
                                    ?disabled=${catalogBusy.value[`role-${role.id}`]}
                                >
                                    Assign
                                </button>
                            </div>
                            <div class="assignment-list">
                                ${role.assignments.length === 0 ? html`
                                    <span class="provider-models">No assignments yet.</span>
                                ` : role.assignments.map((assignment) => html`
                                    <div class="assignment-item">
                                        <span class="assignment-label">${assignment.modelLabel}</span>
                                        <button class="btn btn--danger" @click=${() => this._unassignRole(role.id, assignment.modelCatalogEntryId)}>
                                            Remove
                                        </button>
                                    </div>
                                `)}
                            </div>
                        </div>
                    `)}
                </div>
            </div>
        `;
    }
}
