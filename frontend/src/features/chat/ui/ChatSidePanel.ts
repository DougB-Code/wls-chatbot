/**
 * ChatSidePanel.ts renders chat-specific side panel context.
 * frontend/src/features/chat/ui/ChatSidePanel.ts
 */

import { LitElement, css, html } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';

import { activeConversation, conversations } from '../state/chatSignals';

/**
 * display high-level chat context in the side panel.
 */
@customElement('wls-chat-side-panel')
export class ChatSidePanel extends SignalWatcher(LitElement) {
    static styles = css`
        :host {
            display: block;
            height: 100%;
        }

        .panel-content {
            display: flex;
            flex-direction: column;
            gap: 20px;
            min-height: 0;
        }

        .title {
            margin: 0;
            font-size: 15px;
            font-weight: 600;
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
        }

        .subtitle {
            margin: 4px 0 0;
            font-size: 12px;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
        }

        .create-chat-button {
            border: 1px solid hsla(215, 65%, 62%, 0.4);
            background: hsla(215, 65%, 62%, 0.18);
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
            border-radius: 10px;
            padding: 10px 12px;
            font-size: 13px;
            font-weight: 600;
            cursor: pointer;
            transition: background 120ms ease-out, border-color 120ms ease-out;
        }

        .create-chat-button:hover {
            background: hsla(215, 65%, 62%, 0.26);
            border-color: hsla(215, 65%, 62%, 0.55);
        }

        .section {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .section-label {
            margin: 0;
            font-size: 11px;
            letter-spacing: 0.04em;
            text-transform: uppercase;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
        }

        .history-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 10px;
        }

        .recycle-toggle {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            font-size: 11px;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
            cursor: pointer;
            user-select: none;
        }

        .recycle-toggle__label {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            letter-spacing: 0.02em;
            transition: color 120ms ease-out, font-weight 120ms ease-out;
        }

        .recycle-toggle--active .recycle-toggle__label {
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
            font-weight: 700;
        }

        .recycle-toggle__switch {
            position: relative;
            width: 34px;
            height: 20px;
            display: inline-flex;
            align-items: center;
        }

        .recycle-toggle__switch input {
            position: absolute;
            inset: 0;
            opacity: 0;
            margin: 0;
            cursor: pointer;
        }

        .recycle-toggle__pill {
            position: absolute;
            inset: 0;
            border-radius: 999px;
            border: 1px solid hsla(0, 0%, 100%, 0.16);
            background: hsla(0, 0%, 100%, 0.08);
            transition: background 120ms ease-out, border-color 120ms ease-out;
        }

        .recycle-toggle__pill::after {
            content: '';
            position: absolute;
            width: 14px;
            height: 14px;
            top: 2px;
            left: 2px;
            border-radius: 50%;
            background: hsla(0, 0%, 100%, 0.9);
            transition: transform 120ms ease-out;
        }

        .recycle-toggle__switch input:checked + .recycle-toggle__pill {
            border-color: hsla(215, 65%, 62%, 0.5);
            background: hsla(215, 65%, 62%, 0.25);
        }

        .recycle-toggle__switch input:checked + .recycle-toggle__pill::after {
            transform: translateX(14px);
        }

        .section-value {
            margin: 0;
            font-size: 13px;
            color: var(--color-text-secondary, hsla(0, 0%, 100%, 0.68));
            line-height: 1.45;
        }

        .history-list {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .history-item {
            display: flex;
            align-items: center;
            gap: 8px;
            border: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.06));
            background: var(--color-bg-elevated, hsl(220, 22%, 11%));
            color: var(--color-text-secondary, hsla(0, 0%, 100%, 0.68));
            border-radius: 10px;
            padding: 10px;
            font-size: 12px;
            transition: background 120ms ease-out, border-color 120ms ease-out;
        }

        .history-item:hover {
            background: hsla(0, 0%, 100%, 0.06);
            border-color: hsla(0, 0%, 100%, 0.18);
        }

        .history-item--active {
            border-color: hsla(215, 65%, 62%, 0.5);
            background: hsla(215, 65%, 62%, 0.15);
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
        }

        .history-select {
            flex: 1;
            min-width: 0;
            border: none;
            background: transparent;
            color: inherit;
            text-align: left;
            padding: 0;
            font: inherit;
            cursor: pointer;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .history-actions {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            flex-shrink: 0;
        }

        .history-delete {
            flex-shrink: 0;
            width: 24px;
            height: 24px;
            border-radius: 6px;
            border: 1px solid transparent;
            background: transparent;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
            cursor: pointer;
            transition: all 120ms ease-out;
            font-size: 13px;
            line-height: 1;
        }

        .history-delete:hover {
            border-color: hsla(0, 70%, 60%, 0.4);
            background: hsla(0, 70%, 60%, 0.12);
            color: hsla(0, 80%, 72%, 0.95);
        }

        .history-action {
            border-radius: 6px;
            border: 1px solid transparent;
            background: transparent;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
            font-size: 11px;
            padding: 4px 7px;
            cursor: pointer;
            transition: all 120ms ease-out;
        }

        .history-action:hover {
            border-color: hsla(0, 0%, 100%, 0.2);
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
        }

        .history-action--danger:hover {
            border-color: hsla(0, 70%, 60%, 0.4);
            background: hsla(0, 70%, 60%, 0.12);
            color: hsla(0, 80%, 72%, 0.95);
        }

        .history-empty {
            margin: 0;
            font-size: 12px;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
        }
    `;

    @state()
    private _showDeleted = false;

    /**
     * emit a request to create a new chat conversation.
     */
    private _handleCreateChat() {
        this.dispatchEvent(new CustomEvent('chat-create', { bubbles: true, composed: true }));
    }

    /**
     * emit a request to activate a saved conversation.
     */
    private _handleSelectConversation(conversationId: string) {
        this.dispatchEvent(new CustomEvent('chat-select', {
            detail: { conversationId },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * emit a request to recycle a saved conversation.
     */
    private _handleDeleteConversation(event: Event, conversationId: string) {
        event.stopPropagation();
        const confirmed = window.confirm('Move this chat to the recycle bin?');
        if (!confirmed) {
            return;
        }
        this.dispatchEvent(new CustomEvent('chat-delete', {
            detail: { conversationId },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * emit a request to restore a recycled conversation.
     */
    private _handleRestoreConversation(event: Event, conversationId: string) {
        event.stopPropagation();
        this.dispatchEvent(new CustomEvent('chat-restore', {
            detail: { conversationId },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * emit a request to permanently delete a recycled conversation.
     */
    private _handlePurgeConversation(event: Event, conversationId: string) {
        event.stopPropagation();
        const confirmed = window.confirm('Delete this chat permanently? This cannot be undone.');
        if (!confirmed) {
            return;
        }
        this.dispatchEvent(new CustomEvent('chat-purge', {
            detail: { conversationId },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * toggle between active history and recycle bin views.
     */
    private _handleToggleDeleted(event: Event) {
        const target = event.target as HTMLInputElement | null;
        this._showDeleted = Boolean(target?.checked);
    }

    /**
     * render chat side panel metadata.
     */
    render() {
        const activeConversationId = activeConversation.value?.id ?? '';
        const sortedConversations = [...conversations.value.values()]
            .filter((conversation) => this._showDeleted ? conversation.isArchived : !conversation.isArchived)
            .sort((left, right) => right.updatedAt - left.updatedAt);

        return html`
            <div class="panel-content">
                <div>
                    <h2 class="title">Chat Panel</h2>
                    <p class="subtitle">Context for the active chat workspace.</p>
                </div>

                <button class="create-chat-button" @click=${() => this._handleCreateChat()}>
                    New Chat
                </button>

                <section class="section">
                    <div class="history-header">
                        <h3 class="section-label">History</h3>
                        <label class="recycle-toggle ${this._showDeleted ? 'recycle-toggle--active' : ''}">
                            <span class="recycle-toggle__label">🗑 Recycle</span>
                            <span class="recycle-toggle__switch">
                                <input
                                    type="checkbox"
                                    .checked=${this._showDeleted}
                                    @change=${(event: Event) => this._handleToggleDeleted(event)}
                                />
                                <span class="recycle-toggle__pill"></span>
                            </span>
                        </label>
                    </div>
                    <div class="history-list">
                        ${sortedConversations.length === 0
        ? html`<p class="history-empty">${this._showDeleted ? 'Recycle bin is empty.' : 'No saved chats yet.'}</p>`
        : sortedConversations.map((conversation) => html`
                                <div
                                    class="history-item ${conversation.id === activeConversationId ? 'history-item--active' : ''}"
                                >
                                    <button
                                        class="history-select"
                                        @click=${() => this._handleSelectConversation(conversation.id)}
                                        title=${conversation.title}
                                    >
                                        ${conversation.title || 'Untitled conversation'}
                                    </button>
                                    ${this._showDeleted
        ? html`
                                            <div class="history-actions">
                                                <button
                                                    class="history-action"
                                                    @click=${(event: Event) => this._handleRestoreConversation(event, conversation.id)}
                                                    title="Restore conversation"
                                                >
                                                    Restore
                                                </button>
                                                <button
                                                    class="history-action history-action--danger"
                                                    @click=${(event: Event) => this._handlePurgeConversation(event, conversation.id)}
                                                    title="Delete permanently"
                                                >
                                                    Delete
                                                </button>
                                            </div>
                                        `
        : html`
                                            <button
                                                class="history-delete"
                                                @click=${(event: Event) => this._handleDeleteConversation(event, conversation.id)}
                                                title="Move to recycle bin"
                                                aria-label="Move to recycle bin"
                                            >
                                                🗑
                                            </button>
                                        `}
                                </div>
                            `)}
                    </div>
                </section>
            </div>
        `;
    }
}
