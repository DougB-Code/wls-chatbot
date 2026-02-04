/**
 * MessageBubble renders a single chat message with rich block types.
 * frontend/src/features/chat/MessageBubble.ts
 */

import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import type { Message, Block } from '../../types';

import './StreamingIndicator';

/**
 * display a message and dispatch action approval events.
 */
@customElement('wls-message-bubble')
export class MessageBubble extends LitElement {
    static styles = css`
        :host {
            display: block;
            max-width: 85%;
        }

        :host([animate]) {
            animation: messageSlideIn 200ms cubic-bezier(0.16, 1, 0.3, 1);
        }

        :host([role="user"]) {
            align-self: flex-end;
            margin-left: auto;
        }

        :host([role="assistant"]) {
            align-self: flex-start;
        }

        .message {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .header {
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 11px;
            text-transform: uppercase;
            letter-spacing: 0.12em;
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
        }

        .role-icon {
            width: 16px;
            height: 16px;
        }

        .bubble {
            padding: 12px 16px;
            border-radius: 14px;
            border: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.06));
            background: var(--color-bg-surface, hsl(220, 20%, 14%));
        }

        :host([role="user"]) .bubble {
            background: var(--color-user-surface, hsla(215, 65%, 62%, 0.12));
            border-color: var(--color-user-border, hsla(215, 65%, 62%, 0.35));
        }

        :host([role="assistant"]) .bubble {
            background: var(--color-assistant-surface, hsla(265, 55%, 62%, 0.10));
            border-color: var(--color-assistant-border, hsla(265, 55%, 62%, 0.25));
        }

        .content {
            font-size: 15px;
            line-height: 1.65;
        }

        .content p {
            margin: 0;
        }

        .content p + p {
            margin-top: 12px;
        }

        .code-block {
            margin: 12px 0;
            border-radius: 10px;
            overflow: hidden;
            border: 1px solid var(--color-border-subtle);
            background: var(--color-bg-elevated, hsl(220, 22%, 11%));
        }

        .code-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 8px 12px;
            border-bottom: 1px solid var(--color-border-subtle);
            font-size: 11px;
            color: var(--color-text-secondary);
        }

        .code-lang {
            text-transform: uppercase;
            letter-spacing: 0.04em;
        }

        .code-actions {
            display: flex;
            gap: 4px;
        }

        .code-content {
            padding: 12px;
            overflow-x: auto;
        }

        .code-content pre {
            margin: 0;
            font-family: var(--font-mono, 'JetBrains Mono', monospace);
            font-size: 13px;
            line-height: 1.65;
        }

        .thinking {
            margin: 12px 0;
            padding: 12px 16px;
            border-radius: 10px;
            border: 1px solid var(--color-thinking-border, hsla(45, 85%, 58%, 0.30));
            background: var(--color-thinking-surface, hsla(45, 85%, 58%, 0.10));
        }

        .thinking-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 8px;
            cursor: pointer;
            font-size: 13px;
            color: var(--color-thinking, hsl(45, 85%, 58%));
        }

        .thinking-icon {
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .thinking-content {
            margin-top: 12px;
            padding-left: 16px;
            font-size: 13px;
            color: var(--color-text-secondary);
            border-left: 2px solid var(--color-thinking-border);
        }

        .thinking.collapsed .thinking-content {
            display: none;
        }

        .action {
            margin: 12px 0;
            padding: 12px 16px;
            border-radius: 10px;
            border: 1px solid var(--color-action-border, hsla(155, 65%, 48%, 0.35));
            background: var(--color-action-surface, hsla(155, 65%, 48%, 0.10));
        }

        .action.pending {
            border-color: var(--color-thinking-border);
            background: var(--color-thinking-surface);
        }

        .action.running {
            border-color: var(--color-user-border);
            background: var(--color-user-surface);
        }

        .action.failed {
            border-color: var(--color-error-border, hsla(0, 68%, 55%, 0.30));
            background: var(--color-error-surface, hsla(0, 68%, 55%, 0.10));
        }

        .action-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
        }

        .action-tool {
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 13px;
            font-weight: 600;
            color: var(--color-action, hsl(155, 65%, 48%));
        }

        .action.pending .action-tool {
            color: var(--color-thinking);
        }

        .action.failed .action-tool {
            color: var(--color-error, hsl(0, 68%, 55%));
        }

        .action-controls {
            display: flex;
            gap: 8px;
        }

        .action-btn {
            padding: 6px 12px;
            border: 1px solid var(--color-border-default);
            border-radius: 6px;
            background: transparent;
            color: var(--color-text-secondary);
            font-size: 12px;
            cursor: pointer;
            transition: all 120ms ease-out;
        }

        .action-btn:hover {
            background: var(--color-interactive-hover);
            color: var(--color-text-primary);
        }

        .action-btn.approve {
            border-color: var(--color-action-border);
            color: var(--color-action);
        }

        .action-btn.approve:hover {
            background: var(--color-action-surface);
        }

        .action-description {
            margin-top: 8px;
            font-size: 13px;
            color: var(--color-text-secondary);
        }

        .action-result {
            margin-top: 12px;
            padding: 8px 12px;
            border-radius: 6px;
            background: var(--color-bg-elevated);
            font-family: var(--font-mono);
            font-size: 13px;
        }

        .meta {
            display: flex;
            align-items: center;
            gap: 12px;
            font-size: 11px;
            color: var(--color-text-muted);
        }

        .icon-btn {
            padding: 4px 8px;
            border: none;
            border-radius: 4px;
            background: transparent;
            color: var(--color-text-muted);
            font-size: 11px;
            cursor: pointer;
            transition: all 120ms ease-out;
        }

        .icon-btn:hover {
            background: var(--color-interactive-hover);
            color: var(--color-text-secondary);
        }

        @keyframes messageSlideIn {
            from {
                opacity: 0;
                transform: translateY(12px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
    `;

    @property({ type: Object })
    message!: Message;

    @property({ type: Boolean, reflect: true })
    animate = false;

    private _thinkingCollapsed = new Set<number>();

    /**
     * toggle collapse state for a thinking block.
     */
    private _toggleThinking(index: number) {
        if (this._thinkingCollapsed.has(index)) {
            this._thinkingCollapsed.delete(index);
        } else {
            this._thinkingCollapsed.add(index);
        }
        this.requestUpdate();
    }

    /**
     * copy code block content to the clipboard.
     */
    private _copyCode(content: string) {
        if (!navigator.clipboard) {
            return;
        }
        void navigator.clipboard.writeText(content).catch(() => {});
    }

    /**
     * select the correct renderer for a block type.
     */
    private _renderBlock(block: Block, index: number) {
        switch (block.type) {
            case 'text':
                return this._renderTextBlock(block);
            case 'code':
                return this._renderCodeBlock(block);
            case 'thinking':
                return this._renderThinkingBlock(block, index);
            case 'action':
                return this._renderActionBlock(block);
            case 'error':
                return this._renderErrorBlock(block);
            default:
                return this._renderTextBlock(block);
        }
    }

    /**
     * render a text block with basic paragraph formatting.
     */
    private _renderTextBlock(block: Block) {
        // Simple markdown-ish rendering
        const paragraphs = block.content.split('\n\n').filter((paragraph) => paragraph.length > 0);
        return html`
            <div class="content">
                ${paragraphs.map((paragraph) => html`<p>${paragraph}</p>`)}
            </div>
        `;
    }

    /**
     * render a code block with header and copy action.
     */
    private _renderCodeBlock(block: Block) {
        return html`
            <div class="code-block">
                <div class="code-header">
                    <span class="code-lang">${block.language || 'code'}</span>
                    <div class="code-actions">
                        <button class="icon-btn" @click=${() => this._copyCode(block.content)}>
                            Copy
                        </button>
                    </div>
                </div>
                <div class="code-content">
                    <pre><code>${block.content}</code></pre>
                </div>
            </div>
        `;
    }

    /**
     * render a collapsible thinking block.
     */
    private _renderThinkingBlock(block: Block, index: number) {
        const isCollapsed = this._thinkingCollapsed.has(index);
        return html`
            <div class="thinking ${isCollapsed ? 'collapsed' : ''}" @click=${() => this._toggleThinking(index)}>
                <div class="thinking-header">
                    <span class="thinking-icon">
                        💭 Thinking...
                    </span>
                    <span>${isCollapsed ? '▶' : '▼'}</span>
                </div>
                <div class="thinking-content">
                    ${block.content}
                </div>
            </div>
        `;
    }

    /**
     * render a tool action block with approval controls.
     */
    private _renderActionBlock(block: Block) {
        if (!block.action) return nothing;
        const action = block.action;
        const statusClass = action.status.toLowerCase();

        return html`
            <div class="action ${statusClass}">
                <div class="action-header">
                    <span class="action-tool">
                        ⚡ ${action.toolName}
                    </span>
                    ${action.status === 'pending' ? html`
                        <div class="action-controls">
                            <button class="action-btn approve" @click=${() => this._approveAction(action.id)}>
                                Allow
                            </button>
                            <button class="action-btn" @click=${() => this._rejectAction(action.id)}>
                                Deny
                            </button>
                        </div>
                    ` : nothing}
                </div>
                <div class="action-description">${action.description}</div>
                ${action.result ? html`
                    <div class="action-result">${action.result}</div>
                ` : nothing}
            </div>
        `;
    }

    /**
     * render an error block with alert styling.
     */
    private _renderErrorBlock(block: Block) {
        return html`
            <div class="action failed">
                <div class="action-header">
                    <span class="action-tool">❌ Error</span>
                </div>
                <div class="action-description">${block.content}</div>
            </div>
        `;
    }

    /**
     * emit an approval event for a tool action.
     */
    private _approveAction(actionId: string) {
        this.dispatchEvent(new CustomEvent('action-approve', {
            detail: { actionId },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * emit a rejection event for a tool action.
     */
    private _rejectAction(actionId: string) {
        this.dispatchEvent(new CustomEvent('action-reject', {
            detail: { actionId },
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * format a timestamp for display in the header.
     */
    private _formatTime(timestamp: number): string {
        return new Date(timestamp).toLocaleTimeString([], {
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    /**
     * render the message bubble with metadata and blocks.
     */
    render() {
        const { message } = this;
        const roleLabel = message.role === 'user' ? 'You' : 'Assistant';

        return html`
            <div class="message">
                <div class="header">
                    <span>${roleLabel}</span>
                    <span>•</span>
                    <span>${this._formatTime(message.timestamp)}</span>
                </div>
                <div class="bubble">
                    ${message.blocks.map((block, i) => this._renderBlock(block, i))}
                    ${message.isStreaming ? html`<wls-streaming-indicator></wls-streaming-indicator>` : nothing}
                </div>
                ${message.metadata ? html`
                    <div class="meta">
                        ${message.metadata.model ? html`<span>${message.metadata.model}</span>` : nothing}
                        ${message.metadata.tokensOut ? html`<span>${message.metadata.tokensOut} tokens</span>` : nothing}
                        ${message.metadata.latencyMs ? html`<span>${message.metadata.latencyMs}ms</span>` : nothing}
                    </div>
                ` : nothing}
            </div>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-message-bubble': MessageBubble;
    }
}
