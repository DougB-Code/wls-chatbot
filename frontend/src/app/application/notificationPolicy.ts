/**
 * translate errors into user-facing toast notifications and stored records.
 * frontend/src/app/application/notificationPolicy.ts
 */

import { pushToast } from '../state/toastSignals';
import { createNotification } from '../../features/notifications/application/notificationPolicy';
import { log } from '../../lib/logger';

/**
 * surface an error message via toast state and stored records.
 */
export function notifyError(message: string, title = 'Error'): void {
    pushToast({ type: 'error', title, message });
    void createNotification({ type: 'error', title, message }).catch((err) => {
        const errorMessage = err instanceof Error ? err.message : 'Unknown notification error';
        log.warn().str('error', errorMessage).msg('Failed to persist error notification');
    });
}
