import { expect, describe, it, vi, afterEach } from 'vitest';

// The host evaluates an event only in a page that reported a listener for its
// name, so the runtime must report the first listener of each name and the
// removal of the last one.
const sent = vi.hoisted(() => []);
vi.mock('./system', () => ({ invoke: (message) => sent.push(message) }));

const { On, Once, Off, OffAll } = await import('./events');

afterEach(() => {
  OffAll();
  sent.length = 0;
});

describe("listener reports", () => {
  it("reports the first listener of a name and the removal of the last one", () => {
    const first = On('a', () => {});
    const second = On('a', () => {});
    expect(sent).toEqual(["wails:event:listen:a"]);
    first();
    expect(sent).toEqual(["wails:event:listen:a"]);
    second();
    expect(sent).toEqual(["wails:event:listen:a", "wails:event:unlisten:a"]);
  });

  it("reports the removal when a listener expires during dispatch", () => {
    Once('b', () => {});
    window._wails.dispatchWailsEvent({ name: 'b', data: null });
    expect(sent).toEqual(["wails:event:listen:b", "wails:event:unlisten:b"]);
  });

  it("reports the removal for Off and OffAll", () => {
    On('c', () => {});
    On('d', () => {});
    Off('c');
    OffAll();
    expect(sent).toEqual(["wails:event:listen:c", "wails:event:listen:d", "wails:event:unlisten:c", "wails:event:unlisten:d"]);
  });
});
