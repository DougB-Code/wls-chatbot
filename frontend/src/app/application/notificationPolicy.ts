/**
 * translate errors into user-facing toast notifications and stored records.
 * frontend/src/app/application/notificationPolicy.ts
 */

import { pushToast } from '../state/toastSignals';
import { createNotification } from '../../features/notifications/application/notificationPolicy';

/**
 * surface an error message via toast state and stored records.
 */
export function notifyError(message: string, title = 'Error'): void {
    pushToast({ type: 'error', title, message });
    void createNotification({ type: 'error', title, message }).catch(() => undefined);
}
