/**
 * render transient toast notifications from toast state.
 */

import { LitElement, html, css, nothing } from 'lit';
import { customElement } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';
import { dismissToast, toasts } from '../store/toastSignals';

/**
 * render the current toast stack from toast state.
 */
@customElement('wls-toast-stack')
export class ToastStack extends SignalWatcher(LitElement) {
    static styles = css`
        :host {
            position: fixed;
            right: var(--space-4);
            bottom: var(--space-4);
            display: flex;
            flex-direction: column;
            gap: var(--space-2);
            z-index: var(--z-toast);
            max-width: 360px;
        }

        .gl-toast {
            display: flex;
            align-items: flex-start;
            gap: var(--space-3);
            padding: var(--space-3) var(--space-4);
            border-radius: var(--radius-lg);
            border: 1px solid var(--color-border-subtle);
            background: var(--color-bg-surface);
            box-shadow: var(--shadow-md);
        }

        .gl-toast--error {
            border-color: var(--color-error-border);
            background: var(--color-error-surface);
        }

        .gl-toast__content {
            display: flex;
            flex-direction: column;
            gap: var(--space-1);
            flex: 1;
            min-width: 0;
        }

        .gl-toast__title {
            font-size: var(--text-sm);
            font-weight: 600;
        }

        .gl-toast__message {
            font-size: var(--text-xs);
            color: var(--color-text-secondary);
            word-break: break-word;
        }

        .gl-toast--error .gl-toast__message {
            color: var(--color-text-primary);
        }

        .gl-toast__close {
            border: 0;
            background: transparent;
            color: var(--color-text-muted);
            font-size: 16px;
            cursor: pointer;
            padding: 0;
            line-height: 1;
        }

        .gl-toast__close:hover {
            color: var(--color-text-primary);
        }
    `;

    /**
     * render the current toast list or nothing if empty.
     */
    render() {
        const toastList = toasts.value;
        if (toastList.length === 0) {
            return nothing;
        }

        return html`
            ${toastList.map(toast => html`
                <div class="gl-toast gl-toast--${toast.type}" role="status" aria-live="polite">
                    <div class="gl-toast__content">
                        ${toast.title ? html`<div class="gl-toast__title">${toast.title}</div>` : nothing}
                        <div class="gl-toast__message">${toast.message}</div>
                    </div>
                    <button
                        class="gl-toast__close"
                        type="button"
                        @click=${() => dismissToast(toast.id)}
                        aria-label="Dismiss notification"
                    >
                        ×
                    </button>
                </div>
            `)}
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-toast-stack': ToastStack;
    }
}
