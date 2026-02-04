/**
 * provide a styled dropdown for selecting provider models.
 * frontend/src/features/chat/ModelSelector.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';

/**
 * describe a selectable model option.
 */
export interface ModelOption {
    id: string;
    name: string;
}

/**
 * render a model selector and emit selection events.
 */
@customElement('wls-model-selector')
export class ModelSelector extends LitElement {
    static styles = css`
        :host {
            position: relative;
            display: inline-block;
        }

        .selector {
            position: relative;
            min-width: 150px;
        }

        .selector-button {
            width: 100%;
            padding: 6px 32px 6px 12px;
            border-radius: 6px;
            border: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.1));
            background: var(--color-bg-elevated, hsl(220, 22%, 11%));
            color: var(--color-text-primary, hsl(0, 0%, 100%));
            font-size: 13px;
            font-family: inherit;
            cursor: pointer;
            transition: all 120ms ease-out;
            text-align: left;
            position: relative;
        }

        .selector-button:hover:not(:disabled) {
            border-color: var(--color-border-default, hsla(0, 0%, 100%, 0.15));
            background: var(--color-bg-surface, hsl(220, 20%, 14%));
        }

        .selector-button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }

        .selector-button::after {
            content: '▼';
            position: absolute;
            right: 10px;
            top: 50%;
            transform: translateY(-50%);
            font-size: 10px;
            opacity: 0.7;
        }

        .selector-button[aria-expanded="true"]::after {
            transform: translateY(-50%) rotate(180deg);
        }

        .dropdown {
            position: absolute;
            top: calc(100% + 4px);
            left: 0;
            right: 0;
            background: hsl(220, 20%, 14%);
            border: 1px solid var(--color-border-default, hsla(0, 0%, 100%, 0.15));
            border-radius: 6px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
            z-index: 1000;
            max-height: 240px;
            overflow-y: auto;
        }

        .dropdown::-webkit-scrollbar {
            width: 8px;
        }

        .dropdown::-webkit-scrollbar-track {
            background: hsl(220, 22%, 11%);
            border-radius: 4px;
        }

        .dropdown::-webkit-scrollbar-thumb {
            background: hsla(0, 0%, 100%, 0.2);
            border-radius: 4px;
        }

        .dropdown::-webkit-scrollbar-thumb:hover {
            background: hsla(0, 0%, 100%, 0.3);
        }

        .option {
            padding: 8px 12px;
            cursor: pointer;
            color: hsl(0, 0%, 100%);
            transition: background 100ms ease-out;
        }

        .option:hover {
            background: hsl(220, 20%, 18%);
        }

        .option[aria-selected="true"] {
            background: hsl(25, 95%, 55%);
            color: hsl(0, 0%, 100%);
        }

        .option[aria-selected="true"]:hover {
            background: hsl(25, 95%, 60%);
        }

        .hidden {
            display: none;
        }
    `;

    @property({ type: Array })
    models: ModelOption[] = [];

    @property({ type: String })
    selected = '';

    @property({ type: Boolean })
    disabled = false;

    @state()
    private _isOpen = false;

    /**
     * toggle the dropdown open state.
     */
    private _handleToggle() {
        if (this.disabled) return;
        this._isOpen = !this._isOpen;
    }

    /**
     * update the selected model and notify listeners.
     */
    private _handleSelect(modelId: string) {
        this.selected = modelId;
        this._isOpen = false;
        this.dispatchEvent(new CustomEvent('model-select', {
            detail: { model: modelId },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * close the dropdown when clicking outside the component.
     */
    private _handleClickOutside = (e: MouseEvent) => {
        const path = typeof e.composedPath === 'function' ? e.composedPath() : null;
        const isInside = path ? path.includes(this) : this.contains(e.target as Node);
        if (!isInside) {
            this._isOpen = false;
        }
    };

    /**
     * attach global click listener when mounted.
     */
    connectedCallback() {
        super.connectedCallback();
        document.addEventListener('click', this._handleClickOutside);
    }

    /**
     * remove global click listener when unmounted.
     */
    disconnectedCallback() {
        super.disconnectedCallback();
        document.removeEventListener('click', this._handleClickOutside);
    }

    /**
     * render the selector button and dropdown options.
     */
    render() {
        const selectedModel = this.models.find(m => m.id === this.selected);
        const displayText = selectedModel?.id || 'Select model';

        return html`
            <div class="selector">
                <button
                    class="selector-button"
                    type="button"
                    ?disabled=${this.disabled}
                    aria-expanded=${this._isOpen}
                    aria-haspopup="listbox"
                    @click=${this._handleToggle}
                >
                    ${displayText}
                </button>
                <div class="dropdown ${this._isOpen ? '' : 'hidden'}" role="listbox">
                    ${this.models.map(model => html`
                        <div
                            class="option"
                            role="option"
                            aria-selected=${model.id === this.selected}
                            @click=${() => this._handleSelect(model.id)}
                        >
                            ${model.id}
                        </div>
                    `)}
                </div>
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-model-selector': ModelSelector;
    }
}
