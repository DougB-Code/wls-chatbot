/**
 * send log entries to the backend.
 * frontend/src/core/infrastructure/wails/logTransport.ts
 */

import { Log as LogBackend } from '../../../../wailsjs/go/logger/Logger';

/**
 * describe a structured log payload for the backend.
 */
export interface LogEntry {
    level: string;
    message: string;
    fields: Record<string, string>;
}

/**
 * send a log entry to the backend.
 */
export async function sendLog(entry: LogEntry): Promise<void> {
    try {
        await LogBackend(entry);
    } catch {
        // Logging is best-effort when backend bindings are unavailable.
    }
}
