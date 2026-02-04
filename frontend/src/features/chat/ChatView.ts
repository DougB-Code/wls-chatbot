/**
 * ChatView renders the chat workspace UI and dispatches chat events.
 * frontend/src/features/chat/ChatView.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement, query } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { SignalWatcher } from '@lit-labs/preact-signals';

import { activeConversation, isStreaming } from '../../store/signals';
import { availableModels, effectiveModelId } from '../../policy/chatSelectors';
import { connectedProviderOptions, isConnected, providerConfig } from '../../policy/providerSelectors';

import './MessageBubble';
import './Composer';
import './EmptyState';
import './ModelSelector';
import './ProviderSelector';


/**
 * compose chat header, message list, and composer UI.
 */
@customElement('wls-chat-view')
export class ChatView extends SignalWatcher(LitElement) {
    // ...


    static styles = css`
        :host {
            display: flex;
            flex-direction: column;
            height: 100%;
            min-height: 0;
            background: var(--color-bg-base, hsl(220, 25%, 8%));
        }

        .header {
            flex-shrink: 0;
            padding: 16px 20px;
            border-bottom: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.06));
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 16px;
        }

        .header-left {
            display: flex;
            flex-direction: column;
            gap: 2px;
        }

        .title {
            margin: 0;
            font-size: 18px;
            font-weight: 600;
        }

        .subtitle {
            margin: 0;
            font-size: 13px;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
        }

        .header-right {
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .status {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 6px 12px;
            border-radius: 999px;
            border: 1px solid var(--color-border-subtle);
            background: var(--color-bg-surface, hsl(220, 20%, 14%));
            font-size: 12px;
            color: var(--color-text-secondary);
        }

        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: var(--color-action, hsl(155, 65%, 48%));
            box-shadow: 0 0 8px var(--color-action);
        }

        .status-dot.offline {
            background: var(--color-text-muted);
            box-shadow: none;
        }

        .status-dot.streaming {
            animation: pulse 1.5s ease-in-out infinite;
        }

        wls-model-selector {
            margin-left: 8px;
        }

        .message-list {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
            display: flex;
            flex-direction: column;
            gap: 16px;
        }

        .message-list--empty {
            align-items: center;
            justify-content: center;
            padding: 0;
            background: #161A22;
        }

        /* Hide scrollbar while allowing scroll */
        .message-list::-webkit-scrollbar {
            display: none;
        }
        .message-list {
            -ms-overflow-style: none;
            scrollbar-width: none;
        }

        @keyframes pulse {
            0%, 100% {
                opacity: 1;
                transform: scale(1);
            }
            50% {
                opacity: 0.6;
                transform: scale(1.1);
            }
        }
    `;

    @query('.message-list')
    private _messageList!: HTMLElement;

    private _lastMessageCount = 0;
    private _allowMessageAnimation = false;
    private _seenMessageIds = new Set<string>();

    /**
     * react to updates and keep the message list scrolled.
     */
    updated(_changedProperties: Map<string, unknown>) {
        super.updated(_changedProperties);
        const conversation = activeConversation.value;
        const messages = conversation?.messages ?? [];
        const messageCount = messages.length;
        const streaming = isStreaming.value;

        if (this._shouldAutoScroll(messageCount, streaming)) {
            this._scrollToBottom();
        }

        this._updateMessageTracking(messages, Boolean(conversation));
    }

    /**
     * scroll the message list to the latest content.
     */
    private _scrollToBottom() {
        requestAnimationFrame(() => {
            if (this._messageList) {
                this._messageList.scrollTop = this._messageList.scrollHeight;
            }
        });
    }

    /**
     * determine whether to auto-scroll based on user position and message changes.
     */
    private _shouldAutoScroll(messageCount: number, streaming: boolean): boolean {
        const initialHydrate = this._lastMessageCount === 0 && messageCount > 0;
        if (initialHydrate) {
            return true;
        }

        if (!this._isNearBottom()) {
            return false;
        }

        return messageCount > this._lastMessageCount || streaming;
    }

    /**
     * check whether the user is already near the bottom of the list.
     */
    private _isNearBottom(): boolean {
        if (!this._messageList) {
            return true;
        }

        const threshold = 48;
        const distance = this._messageList.scrollHeight
            - this._messageList.scrollTop
            - this._messageList.clientHeight;
        return distance <= threshold;
    }

    /**
     * track seen messages and enable animations after initial hydration.
     */
    private _updateMessageTracking(messages: Array<{ id: string }>, hasConversation: boolean): void {
        this._lastMessageCount = messages.length;
        this._seenMessageIds = new Set(messages.map((message) => message.id));

        if (!this._allowMessageAnimation && hasConversation) {
            this._allowMessageAnimation = true;
        }
    }

    /**
     * dispatch a chat-send event with content and attachments.
     */
    private _handleSend(e: CustomEvent<{ content: string; attachments: File[] }>) {
        const { content, attachments } = e.detail;
        this.dispatchEvent(new CustomEvent('chat-send', {
            detail: { content, attachments },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * dispatch a chat-send event for the selected suggestion.
     */
    private _handleSuggestionSelect(e: CustomEvent<{ suggestion: string }>) {
        this.dispatchEvent(new CustomEvent('chat-send', {
            detail: { content: e.detail.suggestion, attachments: [] },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * emit a stop-stream event to halt generation.
     */
    private _handleStopStream() {
        this.dispatchEvent(new CustomEvent('chat-stop-stream', {
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * render the chat UI using current signal state.
     */
    render() {
        // Access signals directly for reactivity
        const conversation = activeConversation.value;
        const streaming = isStreaming.value;

        const messages = conversation?.messages ?? [];
        const hasMessages = messages.length > 0;
        const provider = providerConfig.value;
        const connected = isConnected.value;
        const statusDotClass = streaming ? 'streaming' : (connected ? '' : 'offline');
        const models = availableModels.value;
        const selectedModel = effectiveModelId.value;
        const providerOptions = connectedProviderOptions.value;
        const isProviderSwitchDisabled = streaming || providerOptions.length < 2;

        return html`
            <header class="header">
                <div class="header-left">
                    <h1 class="title">${conversation?.title ?? 'New Chat'}</h1>
                    <p class="subtitle">
                        ${provider ? html`
                            <wls-provider-selector
                                .providers=${providerOptions}
                                .selected=${provider.name}
                                ?disabled=${isProviderSwitchDisabled}
                                @provider-select=${this._handleProviderSelect}
                            ></wls-provider-selector>
                            <wls-model-selector
                                .models=${models.map(m => ({ id: m.id, name: m.name }))}
                                .selected=${selectedModel}
                                ?disabled=${streaming}
                                @model-select=${this._handleModelSelect}
                            ></wls-model-selector>
                        ` : 'No provider connected'}
                    </p>
                </div>
                <div class="header-right">
                    <div class="status">
                        <span class="status-dot ${statusDotClass}"></span>
                        <span>${streaming ? 'Generating...' : (connected ? 'Ready' : 'Offline')}</span>
                    </div>
                </div>
            </header>

            <div class="message-list ${hasMessages ? '' : 'message-list--empty'}">
                ${hasMessages ? repeat(
            messages,
            (msg) => msg.id,
            (msg) => html`
                        <wls-message-bubble
                            .message=${msg}
                            role=${msg.role}
                            ?animate=${this._allowMessageAnimation && !this._seenMessageIds.has(msg.id)}
                            @action-approve=${this._handleActionApprove}
                            @action-reject=${this._handleActionReject}
                        ></wls-message-bubble>
                    `
        ) : html`
                    <wls-empty-state
                        provider=${provider?.displayName ?? ''}
                        model=${selectedModel ?? ''}
                        @suggestion-select=${this._handleSuggestionSelect}
                    ></wls-empty-state>
                `}
            </div>

            <wls-composer
                ?disabled=${!connected}
                ?streaming=${streaming}
                @send=${this._handleSend}
                @stop-stream=${this._handleStopStream}
            ></wls-composer>
        `;
    }

    /**
     * forward action-approve events to the parent shell.
     */
    private _handleActionApprove(e: CustomEvent<{ actionId: string }>) {
        this.dispatchEvent(new CustomEvent('chat-action-approve', {
            detail: e.detail,
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * forward action-reject events to the parent shell.
     */
    private _handleActionReject(e: CustomEvent<{ actionId: string }>) {
        this.dispatchEvent(new CustomEvent('chat-action-reject', {
            detail: e.detail,
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * emit the selected model change to the parent shell.
     */
    private _handleModelSelect(e: CustomEvent<{ model: string }>) {
        const newModel = e.detail.model;

        this.dispatchEvent(new CustomEvent('chat-model-select', {
            detail: { model: newModel },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * emit the selected provider change to the parent shell.
     */
    private _handleProviderSelect(e: CustomEvent<{ provider: string }>) {
        const newProvider = e.detail.provider;

        this.dispatchEvent(new CustomEvent('chat-provider-select', {
            detail: { provider: newProvider },
            bubbles: true,
            composed: true,
        }));
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-chat-view': ChatView;
    }
}
