/**
 * provide the chat input composer with attachments and send controls.
 */

import { LitElement, html, css, nothing } from 'lit';
import { customElement, property, state, query } from 'lit/decorators.js';

/**
 * collect user input and dispatch send/stop events.
 */
@customElement('wls-composer')
export class Composer extends LitElement {
    static styles = css`
        :host {
            display: block;
            flex-shrink: 0;
            padding: 16px 20px;
            border-top: 1px solid var(--color-border-subtle, hsla(0, 0%, 100%, 0.06));
            background: var(--color-bg-elevated, hsl(220, 22%, 11%));
        }

        .container {
            display: flex;
            align-items: flex-end;
            gap: 12px;
            padding: 12px;
            border-radius: 14px;
            border: 1px solid var(--color-border-default, hsla(0, 0%, 100%, 0.10));
            background: var(--color-bg-surface, hsl(220, 20%, 14%));
            transition: border-color 120ms ease-out, box-shadow 120ms ease-out;
        }

        .container:focus-within {
            border-color: var(--color-border-focus, hsla(215, 65%, 62%, 0.6));
            box-shadow: 0 0 0 3px hsla(215, 65%, 62%, 0.15);
        }

        .container.drag-over {
            border-color: var(--color-action, hsl(155, 65%, 48%));
            background: var(--color-action-surface);
        }

        .input-wrapper {
            flex: 1;
            min-height: 0;
        }

        textarea {
            width: 100%;
            min-height: 24px;
            max-height: 200px;
            padding: 8px 4px;
            border: none;
            background: transparent;
            color: var(--color-text-primary, hsla(0, 0%, 100%, 0.92));
            font-family: var(--font-sans, 'Inter', sans-serif);
            font-size: 15px;
            line-height: 1.5;
            resize: none;
            outline: none;
            
            /* Hide scrollbar */
            -ms-overflow-style: none;
            scrollbar-width: none;
        }

        textarea::-webkit-scrollbar {
            display: none;
        }

        textarea::placeholder {
            color: var(--color-text-muted, hsla(0, 0%, 100%, 0.45));
        }

        .actions {
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .action-btn {
            width: 36px;
            height: 36px;
            display: flex;
            align-items: center;
            justify-content: center;
            border: none;
            border-radius: 10px;
            background: transparent;
            color: var(--color-text-secondary, hsla(0, 0%, 100%, 0.68));
            cursor: pointer;
            transition: background 120ms ease-out, color 120ms ease-out;
        }

        .action-btn:hover {
            background: var(--color-interactive-hover, hsla(0, 0%, 100%, 0.08));
            color: var(--color-text-primary);
        }

        .action-btn:disabled {
            opacity: 0.4;
            cursor: not-allowed;
        }

        .action-btn svg {
            width: 20px;
            height: 20px;
        }

        .send-btn {
            width: 40px;
            height: 36px;
            display: flex;
            align-items: center;
            justify-content: center;
            border: none;
            border-radius: 10px;
            background: linear-gradient(135deg, var(--color-user, hsl(215, 65%, 62%)) 0%, hsl(235, 60%, 55%) 100%);
            color: white;
            cursor: pointer;
            transition: opacity 120ms ease-out, transform 120ms ease-out;
        }

        .send-btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }

        .send-btn:not(:disabled):hover {
            transform: scale(1.05);
        }

        .send-btn:not(:disabled):active {
            transform: scale(0.95);
        }

        .send-btn svg {
            width: 18px;
            height: 18px;
        }

        .context {
            display: flex;
            flex-wrap: wrap;
            gap: 8px;
            margin-top: 12px;
        }

        .chip {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            padding: 4px 10px;
            border-radius: 999px;
            background: var(--color-interactive-hover);
            font-size: 11px;
            color: var(--color-text-secondary);
        }

        .chip-remove {
            width: 14px;
            height: 14px;
            display: flex;
            align-items: center;
            justify-content: center;
            border: none;
            border-radius: 50%;
            background: transparent;
            color: inherit;
            cursor: pointer;
            margin-left: 2px;
        }

        .chip-remove:hover {
            background: hsla(0, 0%, 100%, 0.1);
        }

        .drop-overlay {
            position: absolute;
            inset: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: 14px;
            background: var(--color-action-surface);
            border: 2px dashed var(--color-action);
            color: var(--color-action);
            font-weight: 600;
            opacity: 0;
            pointer-events: none;
            transition: opacity 150ms ease-out;
        }

        .container.drag-over .drop-overlay {
            opacity: 1;
        }

        /* Hidden file input */
        input[type="file"] {
            display: none;
        }
    `;

    @property({ type: Boolean })
    disabled = false;

    @property({ type: Boolean })
    streaming = false;

    @state()
    private _value = '';

    @state()
    private _attachments: File[] = [];

    @state()
    private _isDragOver = false;

    @query('textarea')
    private _textarea!: HTMLTextAreaElement;

    @query('input[type="file"]')
    private _fileInput!: HTMLInputElement;

    /**
     * update the draft text and auto-resize the textarea.
     */
    private _handleInput(e: Event) {
        const target = e.target as HTMLTextAreaElement;
        this._value = target.value;
        this._autoResize(target);
    }

    /**
     * grow the textarea to fit content up to the max height.
     */
    private _autoResize(textarea: HTMLTextAreaElement) {
        textarea.style.height = 'auto';
        textarea.style.height = Math.min(textarea.scrollHeight, 200) + 'px';
    }

    /**
     * handle keyboard shortcuts for submit and future commands.
     */
    private _handleKeyDown(e: KeyboardEvent) {
        // Send on Enter (without Shift)
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            this._submit();
        }

        // Command palette on /
        if (e.key === '/' && this._value === '') {
            // Future: open command palette
        }
    }

    /**
     * dispatch a send event with text and attachments.
     */
    private _submit() {
        if (this.disabled || this.streaming) return;
        const content = this._value.trim();
        if (!content && this._attachments.length === 0) return;

        this.dispatchEvent(new CustomEvent('send', {
            detail: {
                content,
                attachments: [...this._attachments],
            },
            bubbles: true,
            composed: true,
        }));

        this._value = '';
        this._attachments = [];
        if (this._textarea) {
            this._textarea.style.height = 'auto';
        }
    }

    /**
     * trigger the hidden file input.
     */
    private _handleAttachClick() {
        this._fileInput?.click();
    }

    /**
     * add selected files to the attachment list.
     */
    private _handleFileSelect(e: Event) {
        const input = e.target as HTMLInputElement;
        if (input.files) {
            this._attachments = [...this._attachments, ...Array.from(input.files)];
        }
        input.value = ''; // Reset for next selection
    }

    /**
     * remove a single attachment by index.
     */
    private _removeAttachment(index: number) {
        this._attachments = this._attachments.filter((_, i) => i !== index);
    }

    /**
     * mark the drop zone as active during drag-over.
     */
    private _handleDragOver(e: DragEvent) {
        e.preventDefault();
        this._isDragOver = true;
    }

    /**
     * clear drag-over visual state.
     */
    private _handleDragLeave() {
        this._isDragOver = false;
    }

    /**
     * accept dropped files into the attachment list.
     */
    private _handleDrop(e: DragEvent) {
        e.preventDefault();
        this._isDragOver = false;
        if (e.dataTransfer?.files) {
            this._attachments = [...this._attachments, ...Array.from(e.dataTransfer.files)];
        }
    }

    /**
     * dispatch a stop-stream event to halt generation.
     */
    private _stopStream() {
        this.dispatchEvent(new CustomEvent('stop-stream', {
            bubbles: true,
            composed: true,
        }));
    }

    /**
     * surface the voice input placeholder action.
     */
    private _handleVoiceClick() {
        alert('Voice input is not available yet.');
    }

    /**
     * render the composer UI and attachment chips.
     */
    render() {
        const canSend = this._value.trim() || this._attachments.length > 0;

        return html`
            <div 
                class="container ${this._isDragOver ? 'drag-over' : ''}"
                @dragover=${this._handleDragOver}
                @dragleave=${this._handleDragLeave}
                @drop=${this._handleDrop}
            >
                <div class="input-wrapper">
                    <textarea
                        placeholder="Ask the agent to do something..."
                        .value=${this._value}
                        @input=${this._handleInput}
                        @keydown=${this._handleKeyDown}
                        ?disabled=${this.disabled}
                        rows="1"
                    ></textarea>
                </div>
                <div class="actions">
                    <button 
                        class="action-btn" 
                        @click=${this._handleAttachClick}
                        ?disabled=${this.disabled}
                        title="Attach file"
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48"/>
                        </svg>
                    </button>
                    <button 
                        class="action-btn" 
                        ?disabled=${this.disabled}
                        title="Voice input (coming soon)"
                        @click=${this._handleVoiceClick}
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M12 1a3 3 0 00-3 3v8a3 3 0 006 0V4a3 3 0 00-3-3z"/>
                            <path d="M19 10v2a7 7 0 01-14 0v-2"/>
                            <line x1="12" y1="19" x2="12" y2="23"/>
                            <line x1="8" y1="23" x2="16" y2="23"/>
                        </svg>
                    </button>
                    ${this.streaming ? html`
                        <button 
                            class="send-btn" 
                            @click=${this._stopStream}
                            title="Stop generation"
                        >
                            <svg viewBox="0 0 24 24" fill="currentColor">
                                <rect x="6" y="6" width="12" height="12" rx="2"/>
                            </svg>
                        </button>
                    ` : html`
                        <button 
                            class="send-btn" 
                            @click=${this._submit}
                            ?disabled=${!canSend || this.disabled}
                            title="Send message"
                        >
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <line x1="22" y1="2" x2="11" y2="13"/>
                                <polygon points="22 2 15 22 11 13 2 9 22 2"/>
                            </svg>
                        </button>
                    `}
                </div>
                <div class="drop-overlay">Drop files here</div>
            </div>

            ${this._attachments.length > 0 ? html`
                <div class="context">
                    ${this._attachments.map((file, i) => html`
                        <span class="chip">
                            📎 ${file.name}
                            <button class="chip-remove" @click=${() => this._removeAttachment(i)}>×</button>
                        </span>
                    `)}
                </div>
            ` : nothing}

            <input type="file" multiple @change=${this._handleFileSelect} />
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-composer': Composer;
    }
}
