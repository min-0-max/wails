package application

import (
	"strings"
	"testing"
	"time"
)

// TestWindowEventIsEvaluatedOnlyInItsWindow checks that an event emitted by a
// window is evaluated in that window's page only, while an application event
// is evaluated in every page that listens to its name. Each page evaluates
// the events it receives, so an event delivered to another window's page is
// evaluated there for nothing.
func TestWindowEventIsEvaluatedOnlyInItsWindow(t *testing.T) {
	probe := &threadProbeApp{}
	prev := globalApplication
	app := &App{impl: probe, windows: make(map[uint]Window)}
	app.Event = &EventManager{app: app}
	app.customEventProcessor = NewWailsEventProcessor(app.Event.dispatch)
	app.wailsEventListeners = []WailsEventListener{&EventIPCTransport{app: app}}
	globalApplication = app
	defer func() { globalApplication = prev }()
	probe.onMain.Store(true)

	own := &WebviewWindow{id: 1, options: WebviewWindowOptions{Name: "own"}, impl: &stubWindowImpl{}, runtimeLoaded: true}
	other := &WebviewWindow{id: 2, options: WebviewWindowOptions{Name: "other"}, impl: &stubWindowImpl{}, runtimeLoaded: true}
	app.windows[own.id] = own
	app.windows[other.id] = other
	for _, name := range []string{"sync", "global", "marker"} {
		own.HandleMessage("wails:event:listen:" + name)
	}
	for _, name := range []string{"sync", "global"} {
		other.HandleMessage("wails:event:listen:" + name)
	}

	own.EmitEvent("sync")
	app.Event.Emit("global")
	own.EmitEvent("marker")

	// The event queue is FIFO: the marker reaches the own page after the
	// events emitted before it, so the other page is complete once it arrives.
	deadline := time.Now().Add(5 * time.Second)
	for {
		got := own.impl.(*stubWindowImpl).delivered()
		if len(got) == 3 || time.Now().After(deadline) {
			break
		}
		time.Sleep(time.Millisecond)
	}

	ownGot := own.impl.(*stubWindowImpl).delivered()
	if len(ownGot) != 3 || !strings.Contains(ownGot[0], `"sync"`) || !strings.Contains(ownGot[1], `"global"`) || !strings.Contains(ownGot[2], `"marker"`) {
		t.Fatalf("own page evaluated %v, want \"sync\", \"global\" and \"marker\" in order", ownGot)
	}
	otherGot := other.impl.(*stubWindowImpl).delivered()
	if len(otherGot) != 1 || !strings.Contains(otherGot[0], `"global"`) {
		t.Fatalf("other page evaluated %v, want only the application event \"global\"", otherGot)
	}
}
