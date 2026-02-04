/**
 * send log entries to the backend with a console fallback.
 * frontend/src/core/infrastructure/wails/logTransport.ts
 */

import { Log as LogBackend } from '../../../../wailsjs/go/logger/Logger';

/**
 * describe a structured log payload for the backend.
 */
export type LogEntry = {
    level: string;
    message: string;
    fields: Record<string, string>;
};

/**
 * send a log entry to the backend or fallback to console.
 */
export async function sendLog(entry: LogEntry): Promise<void> {
    try {
        await LogBackend(entry);
    } catch {
        // Fallback for dev mode/browser without wails connected if needed
        console.log(`[${entry.level.toUpperCase()}] ${entry.message}`, entry.fields);
    }
}
