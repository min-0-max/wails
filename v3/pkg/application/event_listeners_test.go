package application

import (
	"strings"
	"testing"
)

// The runtime of a page reports the event names it listens to with
// wails:event:listen and wails:event:unlisten, and dispatchWailsEvent ignores
// any other name. These tests check that the window evaluates only events
// with a reported listener once the runtime is ready, as each evaluation of
// an ignored event still allocates in the page.

// TestUnlistenedEventIsNotEvaluated checks that an event for which the page
// reported no listener is not evaluated in the page.
func TestUnlistenedEventIsNotEvaluated(t *testing.T) {
	probe, win, restore := newOrderingProbe(t)
	defer restore()
	impl := win.impl.(*stubWindowImpl)
	probe.onMain.Store(true)

	win.DispatchWailsEvent(&CustomEvent{Name: "unlistened"})
	probe.runPending()

	if got := impl.delivered(); len(got) != 0 {
		t.Fatalf("an event without a listener in the page was evaluated %d times: %v", len(got), got)
	}
}

// TestListenedEventIsEvaluatedUntilUnlisten checks that an event is evaluated
// while the page reports a listener for its name, and not after the page
// reports that the last listener went.
func TestListenedEventIsEvaluatedUntilUnlisten(t *testing.T) {
	probe, win, restore := newOrderingProbe(t)
	defer restore()
	impl := win.impl.(*stubWindowImpl)
	probe.onMain.Store(true)

	win.HandleMessage("wails:event:listen:listened")
	win.DispatchWailsEvent(&CustomEvent{Name: "listened"})
	win.DispatchWailsEvent(&CustomEvent{Name: "unlistened"})
	win.HandleMessage("wails:event:unlisten:listened")
	win.DispatchWailsEvent(&CustomEvent{Name: "listened"})
	probe.runPending()

	got := impl.delivered()
	if len(got) != 1 || !strings.Contains(got[0], `"listened"`) {
		t.Fatalf("evaluated %v, want only the first event \"listened\"", got)
	}
}

// TestEventBeforeRuntimeReadyIsKept checks that an event emitted before the
// runtime of the page is ready is kept for the pending queue, because the page
// has not reported its listeners yet.
func TestEventBeforeRuntimeReadyIsKept(t *testing.T) {
	probe, win, restore := newOrderingProbe(t)
	defer restore()
	probe.onMain.Store(true)
	win.runtimeLoaded = false

	win.DispatchWailsEvent(&CustomEvent{Name: "early"})
	probe.runPending()

	win.pendingJSMutex.Lock()
	pending := append([]string(nil), win.pendingJS...)
	win.pendingJSMutex.Unlock()
	if len(pending) != 1 || !strings.Contains(pending[0], `"early"`) {
		t.Fatalf("pending JS %v, want the event \"early\"", pending)
	}
}
