/**
 * render an animated streaming indicator for assistant responses.
 */

import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';

/**
 * display a lightweight animated dot sequence.
 */
@customElement('wls-streaming-indicator')
export class StreamingIndicator extends LitElement {
    static styles = css`
        :host {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            padding: 4px;
        }

        .dot {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: var(--color-assistant, hsl(265, 55%, 62%));
            animation: streamingDot 1.4s ease-in-out infinite;
        }

        .dot:nth-child(2) {
            animation-delay: 0.2s;
        }

        .dot:nth-child(3) {
            animation-delay: 0.4s;
        }

        @keyframes streamingDot {
            0%, 60%, 100% {
                opacity: 0.3;
                transform: scale(0.8);
            }
            30% {
                opacity: 1;
                transform: scale(1);
            }
        }
    `;

    /**
     * render the three-dot streaming animation.
     */
    render() {
        return html`
            <span class="dot"></span>
            <span class="dot"></span>
            <span class="dot"></span>
        `;
    }
}

declare global {
    interface HTMLElementTagNameMap {
        'wls-streaming-indicator': StreamingIndicator;
    }
}
