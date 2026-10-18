/*
 _	   __	  _ __
| |	 / /___ _(_) /____
| | /| / / __ `/ / / ___/
| |/ |/ / /_/ / / (__  )
|__/|__/\__,_/_/_/____/
The electron alternative for Go
(c) Lea Anthony 2019-present
*/

// The following utilities have been factored out of ./events.ts
// for testing purposes.

import { invoke } from "./system.js";

export const eventListeners = new Map<string, Listener[]>();

export class Listener {
    eventName: string;
    callback: (data: any) => void;
    maxCallbacks: number;

    constructor(eventName: string, callback: (data: any) => void, maxCallbacks: number) {
        this.eventName = eventName;
        this.callback = callback;
        this.maxCallbacks = maxCallbacks || -1;
    }

    dispatch(data: any): boolean {
        try {
            this.callback(data);
        } catch (err) {
            console.error(err);
        }

        if (this.maxCallbacks === -1) return false;
        this.maxCallbacks -= 1;
        return this.maxCallbacks === 0;
    }
}

/**
 * Reports to the host that this page has a listener for the event name, so
 * that the host evaluates events of that name in this page. The host evaluates
 * no other event here, because dispatchWailsEvent would ignore it.
 */
function announceListen(eventName: string): void {
    invoke("wails:event:listen:" + eventName);
}

/**
 * Reports to the host that this page has no listener left for the event name.
 */
function announceUnlisten(eventName: string): void {
    invoke("wails:event:unlisten:" + eventName);
}

/**
 * Adds a listener; the first listener of an event name is reported to the host.
 */
export function listenerOn(listener: Listener): void {
    const listeners = eventListeners.get(listener.eventName);
    if (listeners) {
        listeners.push(listener);
        return;
    }
    eventListeners.set(listener.eventName, [listener]);
    announceListen(listener.eventName);
}

/**
 * Removes every listener of the event name and reports it to the host.
 */
export function listenersClear(eventName: string): void {
    if (eventListeners.delete(eventName)) {
        announceUnlisten(eventName);
    }
}

export function listenerOff(listener: Listener): void {
    let listeners = eventListeners.get(listener.eventName);
    if (!listeners) {
        return;
    }

    listeners = listeners.filter(l => l !== listener);
    if (listeners.length === 0) {
        listenersClear(listener.eventName);
    } else {
        eventListeners.set(listener.eventName, listeners);
    }
}
