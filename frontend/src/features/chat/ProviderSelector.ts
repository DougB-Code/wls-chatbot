/**
 * ProviderSelector.ts renders the active provider pill and a hover menu for switching providers.
 * frontend/src/features/chat/ProviderSelector.ts
 */

import { LitElement, css, html } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * describe a selectable provider option.
 */
export interface ProviderOption {
    name: string;
    displayName: string;
}

/**
 * render provider pill UI and emit provider selection events.
 */
@customElement('wls-provider-selector')
export class ProviderSelector extends LitElement {
    static styles = css`
        :host {
            position: relative;
            display: inline-flex;
            align-items: center;
        }

        .pill {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 4px 10px;
            border-radius: 6px;
            border: 1px solid transparent;
            font-size: 11px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.04em;
            font-family: inherit;
            background: transparent;
            cursor: pointer;
        }

        .pill::after {
            content: '';
            width: 0;
            height: 0;
            border-left: 4px solid transparent;
            border-right: 4px solid transparent;
            border-top: 5px solid currentColor;
            opacity: 0.75;
        }

        .pill.openai {
            background: hsla(160, 70%, 45%, 0.15);
            color: hsl(160, 70%, 45%);
        }

        .pill.groq {
            background: hsla(25, 95%, 55%, 0.15);
            color: hsl(25, 95%, 55%);
        }

        .pill.gemini {
            background: hsla(220, 85%, 60%, 0.15);
            color: hsl(220, 85%, 60%);
        }

        .menu {
            position: absolute;
            top: calc(100% + 6px);
            left: 0;
            min-width: 170px;
            padding: 6px;
            border-radius: 8px;
            border: 1px solid var(--color-border-default, hsla(0, 0%, 100%, 0.15));
            background: var(--color-bg-surface, hsl(220, 20%, 14%));
            box-shadow: 0 10px 22px rgba(0, 0, 0, 0.35);
            opacity: 0;
            visibility: hidden;
            transform: translateY(-4px);
            transition: opacity 120ms ease-out, transform 120ms ease-out, visibility 120ms ease-out;
            z-index: 1100;
        }

        :host(:hover) .menu,
        :host(:focus-within) .menu {
            opacity: 1;
            visibility: visible;
            transform: translateY(0);
        }

        .option {
            width: 100%;
            display: block;
            text-align: left;
            border: 0;
            border-radius: 6px;
            padding: 8px 10px;
            color: var(--color-text-primary, hsl(0, 0%, 100%));
            background: transparent;
            font-size: 12px;
            font-family: inherit;
            cursor: pointer;
        }

        .option:hover,
        .option:focus-visible {
            background: var(--color-interactive-hover, hsla(0, 0%, 100%, 0.08));
            outline: none;
        }

        .option[aria-selected='true'] {
            background: hsl(25, 95%, 55%);
            color: hsl(0, 0%, 100%);
            cursor: default;
        }

        :host([disabled]) .pill {
            opacity: 0.6;
            cursor: not-allowed;
        }

        :host([disabled]) .menu {
            display: none;
        }
    `;

    @property({ type: Array })
    providers: ProviderOption[] = [];

    @property({ type: String })
    selected = '';

    @property({ type: Boolean, reflect: true })
    disabled = false;

    /**
     * emit the selected provider if it is a new value.
     */
    private _handleProviderSelect(providerName: string): void {
        if (this.disabled || providerName === this.selected) {
            return;
        }

        this.dispatchEvent(new CustomEvent('provider-select', {
            detail: { provider: providerName },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * render provider pill and hover menu.
     */
    render() {
        const selectedProvider = this.providers.find((provider) => provider.name === this.selected) ?? null;
        const selectedName = selectedProvider?.displayName ?? 'No Provider';

        return html`
            <button
                class="pill ${this.selected}"
                type="button"
                aria-haspopup="listbox"
                aria-label="Select provider"
                ?disabled=${this.disabled}
            >
                ${selectedName}
            </button>
            <div class="menu" role="listbox" aria-label="Available providers">
                ${this.providers.map((provider) => html`
                    <button
                        class="option"
                        type="button"
                        role="option"
                        aria-selected=${provider.name === this.selected}
                        @click=${() => this._handleProviderSelect(provider.name)}
                    >
                        ${provider.displayName}
                    </button>
                `)}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-provider-selector': ProviderSelector;
    }
}
