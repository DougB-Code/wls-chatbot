/**
 * provide shared settings page styles.
 */

import { css } from 'lit';

export const settingsSharedStyles = css`
    .settings__header {
        margin-bottom: 8px;
    }

    .settings__title {
        margin: 0;
        font-size: 24px;
        font-weight: 600;
    }

    .settings__subtitle {
        margin: 8px 0 0;
        font-size: 14px;
        color: var(--color-text-secondary);
    }

    .card {
        padding: 20px 24px;
        border-radius: 16px;
        border: 1px solid var(--color-border-subtle);
        background: var(--color-bg-surface, hsl(220, 20%, 14%));
    }

    .card__title {
        margin: 0 0 16px;
        font-size: 16px;
        font-weight: 600;
    }
`;
