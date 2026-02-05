/**
 * ChatHeader.ts renders chat header controls for the app shell.
 * frontend/src/features/chat/ui/ChatHeader.ts
 */

import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';
import { SignalWatcher } from '@lit-labs/preact-signals';

import { activeConversation, isStreaming } from '../state/chatSignals';
import { availableModels, effectiveModelId } from '../application/chatSelectors';
import { connectedProviderOptions, isConnected, providerConfig } from '../../settings/application/providerSelectors';

import './ModelSelector';
import './ProviderSelector';

/**
 * render chat header status and model controls.
 */
@customElement('wls-chat-header')
export class ChatHeader extends SignalWatcher(LitElement) {
    static styles = css`
        :host {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 16px;
            width: 100%;
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
            display: flex;
            align-items: center;
            flex-wrap: wrap;
            gap: 8px;
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

    /**
     * render the chat header with status and model controls.
     */
    render() {
        const conversation = activeConversation.value;
        const streaming = isStreaming.value;
        const provider = providerConfig.value;
        const connected = isConnected.value;
        const statusDotClass = streaming ? 'streaming' : (connected ? '' : 'offline');
        const models = availableModels.value;
        const selectedModel = effectiveModelId.value;
        const providerOptions = connectedProviderOptions.value;
        const isProviderSwitchDisabled = streaming || providerOptions.length < 2;

        return html`
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
                            .models=${models.map((model) => ({ id: model.id, name: model.name }))}
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
        `;
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
        'wls-chat-header': ChatHeader;
    }
}
