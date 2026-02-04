/**
 * translate errors into user-facing toast notifications.
 * frontend/src/app/application/notificationPolicy.ts
 */

import { pushToast } from '../state/toastSignals';

/**
 * surface an error message via toast state.
 */
export function notifyError(message: string, title = 'Error'): void {
    pushToast({ type: 'error', title, message });
}
