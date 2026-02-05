/**
 * ShellHeader.ts renders the shared app header and toast override region.
 * frontend/src/shell/ShellHeader.ts
 */

import { LitElement, html, css, nothing } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';

import { dismissToast, toasts, type ToastMessage } from '../app/state/toastSignals';

const toastExitDurationMs = 200;

/**
 * render the shared app header with toast override behavior.
 */
@customElement('wls-shell-header')
export class ShellHeader extends SignalWatcher(LitElement) {
    static styles = css`
        :host {
            display: block;
            width: 100%;
        }

        .header {
            padding: 16px 20px;
            border-bottom: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.06));
            background: var(--color-bg-base, #1b2636);
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 16px;
            min-height: 68px;
        }

        .header--toast {
            justify-content: center;
        }

        .toast {
            width: min(920px, 100%);
            display: flex;
            align-items: flex-start;
            gap: var(--space-3);
            padding: var(--space-3) var(--space-4);
            border-radius: var(--radius-lg);
            border: 1px solid var(--color-border-subtle);
            background: var(--color-bg-surface);
            box-shadow: var(--shadow-md);
            animation: toast-in var(--duration-normal) var(--ease-out) both;
        }

        .toast--error {
            border-color: var(--color-error-border);
            background: var(--color-error-surface);
        }

        .toast--exit {
            animation: toast-out var(--duration-normal) var(--ease-in-out) forwards;
        }

        .toast__content {
            display: flex;
            flex-direction: column;
            gap: var(--space-1);
            flex: 1;
            min-width: 0;
        }

        .toast__title {
            font-size: var(--text-sm);
            font-weight: 600;
        }

        .toast__message {
            font-size: var(--text-xs);
            color: var(--color-text-secondary);
            word-break: break-word;
        }

        .toast--error .toast__message {
            color: var(--color-text-primary);
        }

        .toast__close {
            border: 0;
            background: transparent;
            color: var(--color-text-muted);
            font-size: 16px;
            cursor: pointer;
            padding: 0;
            line-height: 1;
        }

        .toast__close:hover {
            color: var(--color-text-primary);
        }

        @keyframes toast-in {
            from {
                opacity: 0;
                transform: translateY(-6px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        @keyframes toast-out {
            from {
                opacity: 1;
                transform: translateY(0);
            }
            to {
                opacity: 0;
                transform: translateY(-6px);
            }
        }
    `;

    @state()
    private _activeToastId: number | null = null;

    @state()
    private _exitingToastId: number | null = null;

    private _autoDismissTimeoutId: number | null = null;
    private _dismissTimeoutId: number | null = null;
    private _autoDismissDisabled = false;
    private _escapeListenerAttached = false;

    /**
     * handle escape key presses to dismiss the active toast.
     */
    private _handleKeydown = (event: KeyboardEvent) => {
        if (event.key !== 'Escape') {
            return;
        }
        const toast = toasts.value[0];
        if (!toast) {
            return;
        }
        this._requestToastDismiss(toast.id);
    };

    /**
     * clean up timers and listeners when the header disconnects.
     */
    disconnectedCallback() {
        this._clearTimers();
        this._detachEscapeListener();
        super.disconnectedCallback();
    }

    /**
     * react to toast changes and manage dismissal timers.
     */
    updated() {
        super.updated();
        const toast = toasts.value[0] ?? null;
        const nextId = toast?.id ?? null;
        if (nextId !== this._activeToastId) {
            this._setActiveToast(toast);
        }
    }

    /**
     * render the header content or toast override.
     */
    render() {
        const toast = toasts.value[0];
        if (!toast) {
            return html`
                <div class="header">
                    <slot></slot>
                </div>
            `;
        }

        return this._renderToast(toast);
    }

    /**
     * render the toast presentation for the active toast.
     */
    private _renderToast(toast: ToastMessage) {
        const isExiting = this._exitingToastId === toast.id;
        return html`
            <div class="header header--toast">
                <div
                    class="toast ${toast.type === 'error' ? 'toast--error' : ''} ${isExiting ? 'toast--exit' : ''}"
                    role="status"
                    aria-live="polite"
                    @mouseenter=${this._handleToastMouseEnter}
                >
                    <div class="toast__content">
                        ${toast.title ? html`<div class="toast__title">${toast.title}</div>` : nothing}
                        <div class="toast__message">${toast.message}</div>
                    </div>
                    <button
                        class="toast__close"
                        type="button"
                        @click=${() => this._requestToastDismiss(toast.id)}
                        aria-label="Dismiss notification"
                    >
                        ×
                    </button>
                </div>
            </div>
        `;
    }

    /**
     * configure timers and listeners for a newly active toast.
     */
    private _setActiveToast(toast: ToastMessage | null) {
        this._clearTimers();
        this._autoDismissDisabled = false;
        this._exitingToastId = null;
        this._activeToastId = toast?.id ?? null;

        if (!toast) {
            this._detachEscapeListener();
            return;
        }

        this._startAutoDismiss(toast);
        this._attachEscapeListener();
    }

    /**
     * begin the auto-dismiss timer for the current toast.
     */
    private _startAutoDismiss(toast: ToastMessage) {
        const durationMs = toast.durationMs ?? 8000;
        if (this._autoDismissDisabled || durationMs <= 0) {
            return;
        }

        this._autoDismissTimeoutId = window.setTimeout(() => {
            this._autoDismissTimeoutId = null;
            this._requestToastDismiss(toast.id);
        }, durationMs);
    }

    /**
     * stop and clear any active timers.
     */
    private _clearTimers() {
        if (this._autoDismissTimeoutId !== null) {
            window.clearTimeout(this._autoDismissTimeoutId);
            this._autoDismissTimeoutId = null;
        }
        if (this._dismissTimeoutId !== null) {
            window.clearTimeout(this._dismissTimeoutId);
            this._dismissTimeoutId = null;
        }
    }

    /**
     * pause auto-dismiss and require manual dismissal after hover.
     */
    private _handleToastMouseEnter() {
        this._autoDismissDisabled = true;
        this._clearTimers();
    }

    /**
     * start the exit animation and remove the toast from state.
     */
    private _requestToastDismiss(toastId: number) {
        if (this._exitingToastId === toastId) {
            return;
        }

        this._clearTimers();
        this._exitingToastId = toastId;
        this._dismissTimeoutId = window.setTimeout(() => {
            this._dismissTimeoutId = null;
            dismissToast(toastId);
            this._exitingToastId = null;
        }, toastExitDurationMs);
    }

    /**
     * attach a global escape listener while a toast is visible.
     */
    private _attachEscapeListener() {
        if (this._escapeListenerAttached) {
            return;
        }
        window.addEventListener('keydown', this._handleKeydown);
        this._escapeListenerAttached = true;
    }

    /**
     * detach the global escape listener when no toast is active.
     */
    private _detachEscapeListener() {
        if (!this._escapeListenerAttached) {
            return;
        }
        window.removeEventListener('keydown', this._handleKeydown);
        this._escapeListenerAttached = false;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-shell-header': ShellHeader;
    }
}
