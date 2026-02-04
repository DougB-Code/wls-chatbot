/**
 * translate errors into user-facing toast notifications.
 */

import { pushToast } from '../store/toastSignals';

/**
 * surface an error message via toast state.
 */
export function notifyError(message: string, title = 'Error'): void {
    pushToast({ type: 'error', title, message });
}
