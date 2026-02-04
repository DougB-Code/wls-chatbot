/**
 * logger.ts forwards structured frontend events to the backend.
 * frontend/src/lib/logger.ts
 */

import type { LogEntry } from '../transport/logTransport';
import { sendLog } from '../transport/logTransport';

/**
 * accumulate a log entry with context fields and emit it.
 */
class LogEvent {
    private level: string;
    private fields: Record<string, string>;

    /**
     * capture base log level and context fields.
     */
    constructor(level: string, baseFields: Record<string, string>) {
        this.level = level;
        this.fields = { ...baseFields };
    }

    /**
     * add a string field to the pending log entry.
     */
    public str(key: string, value: string): LogEvent {
        this.fields[key] = value;
        return this;
    }

    /**
     * emit the log entry to the backend (or console fallback).
     */
    public msg(message: string): void {
        const entry: LogEntry = {
            level: this.level,
            message: message,
            fields: this.fields,
        };

        void sendLog(entry);
    }
}

/**
 * create log events with shared context fields.
 */
export class Logger {
    private context: Record<string, string> = {};

    /**
     * initialize the logger with default context.
     */
    constructor(context: Record<string, string> = {}) {
        this.context = context;
    }

    /**
     * return a logger with merged context fields.
     */
    public with(context: Record<string, string>): Logger {
        return new Logger({ ...this.context, ...context });
    }

    /**
     * create a debug-level log event.
     */
    public debug(): LogEvent {
        return new LogEvent("debug", this.context);
    }

    /**
     * create a trace-level log event.
     */
    public trace(): LogEvent {
        return new LogEvent("trace", this.context);
    }

    /**
     * create an info-level log event.
     */
    public info(): LogEvent {
        return new LogEvent("info", this.context);
    }

    /**
     * create a warn-level log event.
     */
    public warn(): LogEvent {
        return new LogEvent("warn", this.context);
    }

    /**
     * create an error-level log event.
     */
    public error(): LogEvent {
        return new LogEvent("error", this.context);
    }
}

/**
 * provide a shared root logger for the app.
 */
export const log = new Logger();
