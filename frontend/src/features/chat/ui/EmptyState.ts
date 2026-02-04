/**
 * EmptyState renders the chat empty state with onboarding suggestions.
 * frontend/src/features/chat/ui/EmptyState.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import splashUrl from '../../../assets/images/wls-chatbot-logo-splash.png';

/**
 * display the empty state view for new conversations.
 */
@customElement('wls-empty-state')
export class EmptyState extends LitElement {
    static styles = css`
        :host {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            gap: 16px;
            padding: 0;
            text-align: center;
            animation: fadeIn 300ms ease-out;
            width: 100%;
            height: 100%;
            flex: 1;
        }

        .icon {
            width: 360px;
            height: 360px;
            max-width: min(70vw, 480px);
            max-height: min(70vw, 480px);
            opacity: 0.7;
            object-fit: contain;
        }

        .title {
            margin: 0;
            font-size: 20px;
            font-weight: 600;
            color: var(--color-text-secondary, hsla(0, 0%, 100%, 0.68));
        }

        .description {
            margin: 0;
            max-width: 400px;
            font-size: 14px;
            line-height: 1.6;
            color: var(--color-text-muted);
        }

        .suggestions {
            display: flex;
            flex-wrap: wrap;
            justify-content: center;
            gap: 8px;
            margin-top: 16px;
        }

        .suggestion {
            padding: 10px 16px;
            border: 1px solid var(--color-border-default, hsla(0, 0%, 100%, 0.10));
            border-radius: 10px;
            background: var(--color-bg-surface, hsl(220, 20%, 14%));
            color: var(--color-text-secondary);
            font-size: 13px;
            cursor: pointer;
            transition: all 150ms ease-out;
        }

        .suggestion:hover {
            border-color: var(--color-user-border);
            background: var(--color-user-surface);
            color: var(--color-text-primary);
            transform: translateY(-2px);
        }

        @keyframes fadeIn {
            from {
                opacity: 0;
                transform: translateY(10px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
    `;

    @property({ type: String })
    provider = '';

    @property({ type: String })
    model = '';

    private _suggestions = [
        'Help me write a Python script',
        'Explain a complex concept',
        'Review my code for issues',
        'Draft a technical document',
    ];

    /**
     * emit the selected suggestion to the parent view.
     */
    private _handleSuggestionClick(suggestion: string) {
        this.dispatchEvent(new CustomEvent('suggestion-select', {
            detail: { suggestion },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * render the empty state content and suggestion buttons.
     */
    render() {
        const providerText = this.provider ? `Connected to ${this.provider}` : 'Connect a provider to start chatting.';
        return html`
            <img class="icon" src=${splashUrl} alt="WLS ChatBot" />
            <h2 class="title">Start a new conversation</h2>
            <p class="description">
                ${providerText}
            </p>
            <div class="suggestions">
                ${this._suggestions.map((suggestion) => html`
                    <button class="suggestion" @click=${() => this._handleSuggestionClick(suggestion)}>
                        ${suggestion}
                    </button>
                `)}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-empty-state': EmptyState;
    }
}
