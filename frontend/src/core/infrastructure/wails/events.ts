/**
 * bridge Wails runtime events into the frontend.
 * frontend/src/core/infrastructure/wails/events.ts
 */

import { EventsOn, EventsOff } from '../../../../wailsjs/runtime/runtime';

/**
 * subscribe to a named Wails event channel.
 */
export function onEvent<T>(name: string, handler: (payload: T) => void): void {
    EventsOn(name, handler as (payload: unknown) => void);
}

/**
 * unsubscribe from a named Wails event channel.
 */
export function offEvent(name: string): void {
    EventsOff(name);
}
