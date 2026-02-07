/**
 * bridge Wails runtime events into the frontend.
 * frontend/src/core/infrastructure/wails/events.ts
 */

import { EventsOn } from '../../../../wailsjs/runtime/runtime';

/**
 * describe a callback used to detach an event subscription.
 */
export type EventUnsubscribe = () => void;

/**
 * subscribe to a named Wails event channel.
 */
export function onEvent<T>(name: string, handler: (payload: T) => void): EventUnsubscribe {
    return EventsOn(name, handler as (payload: unknown) => void);
}
