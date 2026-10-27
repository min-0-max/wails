package application

import (
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"
)

type retirementProbeWindow struct {
	Window
	id        uint
	activity  []string
	onMessage func()
}

func (w *retirementProbeWindow) ID() uint { return w.id }
func (w *retirementProbeWindow) HandleMessage(string) {
	w.activity = append(w.activity, "message")
	if w.onMessage != nil {
		w.onMessage()
	}
}
func (w *retirementProbeWindow) HandleKeyEvent(string) {
	w.activity = append(w.activity, "key")
}
func (w *retirementProbeWindow) handleDragAndDropMessage([]string, *DropTargetDetails) {
	w.activity = append(w.activity, "drop")
}

func TestWindowFilesDroppedPreservesRetiredSourceLifetime(t *testing.T) {
	app := &App{windows: make(map[uint]Window)}
	app.Window = newWindowManager(app)
	retired := &WebviewWindow{id: 9031}
	app.Window.Add(retired)
	retiredLifetimeID := retired.callbackLifetimeID
	app.Window.Retire(retired.id, retiredLifetimeID)
	replacement := &retirementProbeWindow{id: retired.id}
	app.Window.Add(replacement)
	app.Window.removeLifetime(retired.id, retiredLifetimeID)

	processor := NewMessageProcessor(slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := windowMethodRequest(t, WindowFilesDropped, `{"filenames":["late.txt"],"x":1,"y":2,"elementDetails":{"id":"drop-target","classList":[],"attributes":{}}}`)
	if _, err := processor.processWindowMethod(request, retired); err != nil {
		t.Fatalf("processWindowMethod(FilesDropped) failed: %v", err)
	}
	event := <-windowDragAndDropBuffer
	if event.lifetimeID != retiredLifetimeID {
		t.Fatalf("file-drop producer lifetime = %d, want source lifetime %d", event.lifetimeID, retiredLifetimeID)
	}
	app.handleDragAndDropMessage(event)
	if len(replacement.activity) != 0 {
		t.Fatalf("late file drop reached replacement window: %v", replacement.activity)
	}
}

func TestRetiredWindowDrainsQueuedCallbacksWithoutDispatchOrWarning(t *testing.T) {
	var warnings []string
	app := &App{windows: make(map[uint]Window)}
	app.options.WarningHandler = func(message string) { warnings = append(warnings, message) }
	app.Window = newWindowManager(app)
	window := &retirementProbeWindow{id: 9001}
	app.Window.Add(window)
	lifetimeID := app.windowLifetimes[window.id].id

	windowMessageBuffer <- &windowMessage{windowId: window.id, lifetimeID: lifetimeID, message: "wails:queued"}
	windowKeyEvents <- &windowKeyEvent{windowId: window.id, lifetimeID: lifetimeID, acceleratorString: "ctrl+x"}
	windowDragAndDropBuffer <- &dragAndDropMessage{windowId: window.id, lifetimeID: lifetimeID, filenames: []string{"queued.txt"}}
	app.Window.Retire(window.id, lifetimeID)

	app.handleWindowMessage(<-windowMessageBuffer)
	app.handleWindowKeyEvent(<-windowKeyEvents)
	app.handleDragAndDropMessage(<-windowDragAndDropBuffer)
	app.Window.Remove(window.id)

	if len(window.activity) != 0 {
		t.Fatalf("retired window received queued callbacks %v", window.activity)
	}
	if len(warnings) != 0 {
		t.Fatalf("retired window produced missing-window warnings: %v", warnings)
	}

	app.handleWindowMessage(&windowMessage{windowId: 9002, lifetimeID: ^uint64(0), message: "wails:unknown"})
	app.handleWindowKeyEvent(&windowKeyEvent{windowId: 9002, lifetimeID: ^uint64(0), acceleratorString: "ctrl+x"})
	app.handleDragAndDropMessage(&dragAndDropMessage{windowId: 9002, lifetimeID: ^uint64(0), filenames: []string{"unknown.txt"}})
	want := []string{
		"WebviewWindow #9002 not found",
		"WebviewWindow #9002 not found",
		"WebviewWindow #9002 not found",
	}
	if !reflect.DeepEqual(warnings, want) {
		t.Fatalf("unknown window diagnostics %v, want %v", warnings, want)
	}
}

func TestRetiredQueuedCallbacksDoNotReachAWindowReusingItsID(t *testing.T) {
	app := &App{windows: make(map[uint]Window)}
	app.Window = newWindowManager(app)
	retired := &retirementProbeWindow{id: 9011}
	app.Window.Add(retired)
	retiredLifetimeID := app.windowLifetimes[retired.id].id

	windowMessageBuffer <- &windowMessage{windowId: retired.id, lifetimeID: retiredLifetimeID, message: "wails:queued"}
	windowKeyEvents <- &windowKeyEvent{windowId: retired.id, lifetimeID: retiredLifetimeID, acceleratorString: "ctrl+x"}
	windowDragAndDropBuffer <- &dragAndDropMessage{windowId: retired.id, lifetimeID: retiredLifetimeID, filenames: []string{"queued.txt"}}
	app.Window.Retire(retired.id, retiredLifetimeID)
	app.Window.Remove(retired.id)
	replacement := &retirementProbeWindow{id: retired.id}
	app.Window.Add(replacement)

	app.handleWindowMessage(<-windowMessageBuffer)
	app.handleWindowKeyEvent(<-windowKeyEvents)
	app.handleDragAndDropMessage(<-windowDragAndDropBuffer)
	if len(replacement.activity) != 0 {
		t.Fatalf("queued callbacks from a retired window reached its replacement: %v", replacement.activity)
	}
}

func TestLateCallbacksFromRetiredWindowDoNotReachWindowReusingItsID(t *testing.T) {
	app := &App{windows: make(map[uint]Window)}
	app.Window = newWindowManager(app)
	// WebviewWindow IDs are also the stable source lifetime token passed by
	// native delegates. A manager-only replacement can reuse the key, but gets
	// a distinct manager lifetime.
	retired := &WebviewWindow{id: 9021}
	app.Window.Add(retired)
	retiredLifetimeID := app.windowLifetimes[retired.id].id
	app.Window.Retire(retired.id, retiredLifetimeID)
	replacement := &retirementProbeWindow{id: retired.id}
	app.Window.Add(replacement)
	app.Window.removeLifetime(retired.id, retiredLifetimeID)
	if current, ok := app.Window.GetByID(retired.id); !ok || current != replacement {
		t.Fatal("late cleanup for the retired lifetime removed its replacement")
	}

	// Native callbacks carry the retired window's numeric ID. Delivering them
	// after same-ID replacement models a producer that ran after retirement.
	windowMessageBuffer <- &windowMessage{windowId: retired.id, lifetimeID: retiredLifetimeID, message: "wails:late"}
	windowKeyEvents <- &windowKeyEvent{windowId: retired.id, lifetimeID: retiredLifetimeID, acceleratorString: "ctrl+x"}
	windowDragAndDropBuffer <- &dragAndDropMessage{windowId: retired.id, lifetimeID: retiredLifetimeID, filenames: []string{"late.txt"}}

	app.handleWindowMessage(<-windowMessageBuffer)
	app.handleWindowKeyEvent(<-windowKeyEvents)
	app.handleDragAndDropMessage(<-windowDragAndDropBuffer)
	if len(replacement.activity) != 0 {
		t.Fatalf("late callbacks from a retired window reached its replacement: %v", replacement.activity)
	}
}

func TestAdmittedCallbackRemainsBoundToOldWindowDuringReplacement(t *testing.T) {
	app := &App{windows: make(map[uint]Window)}
	app.Window = newWindowManager(app)
	entered := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})
	retired := &retirementProbeWindow{id: 9041}
	retired.onMessage = func() {
		close(entered)
		<-release
	}
	app.Window.Add(retired)
	lifetimeID := app.windowLifetimes[retired.id].id
	event := &windowMessage{windowId: retired.id, lifetimeID: lifetimeID, message: "wails:queued"}
	go func() {
		app.handleWindowMessage(event)
		close(finished)
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("callback was not admitted")
	}
	app.Window.Retire(retired.id, lifetimeID)
	app.Window.Remove(retired.id)
	replacement := &retirementProbeWindow{id: retired.id}
	app.Window.Add(replacement)
	close(release)
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("admitted callback did not finish")
	}
	if !reflect.DeepEqual(retired.activity, []string{"message"}) {
		t.Fatalf("admitted callback activity = %v, want old window message", retired.activity)
	}
	if len(replacement.activity) != 0 {
		t.Fatalf("admitted callback reached replacement: %v", replacement.activity)
	}
}

func TestAdmittedWindowReferenceRemainsBoundAfterRetirement(t *testing.T) {
	app := &App{windows: make(map[uint]Window)}
	app.Window = newWindowManager(app)
	retired := &retirementProbeWindow{id: 9042}
	app.Window.Add(retired)
	lifetimeID := app.windowLifetimes[retired.id].id

	// Admission captures the Window value that queued work must use after this
	// point. Retirement may happen before the admitted work invokes the handler.
	admitted, ok, reportMissing := app.admitWindowCallback(retired.id, lifetimeID)
	if !ok || reportMissing || admitted != retired {
		t.Fatalf("admission = (%v, %t, %t), want the original window", admitted, ok, reportMissing)
	}
	app.Window.Retire(retired.id, lifetimeID)
	app.Window.Remove(retired.id)
	replacement := &retirementProbeWindow{id: retired.id}
	app.Window.Add(replacement)

	admitted.HandleMessage("late completion")
	if !reflect.DeepEqual(retired.activity, []string{"message"}) {
		t.Fatalf("admitted work activity = %v, want old window message", retired.activity)
	}
	if len(replacement.activity) != 0 {
		t.Fatalf("admitted work reached same-ID replacement: %v", replacement.activity)
	}
}

func TestCallbackMayRetireItsWindowWithoutHoldingAdmissionLock(t *testing.T) {
	app := &App{windows: make(map[uint]Window)}
	app.Window = newWindowManager(app)
	window := &retirementProbeWindow{id: 9051}
	app.Window.Add(window)
	lifetimeID := app.windowLifetimes[window.id].id
	window.onMessage = func() { app.Window.Retire(window.id, lifetimeID) }
	done := make(chan struct{})
	go func() {
		app.handleWindowMessage(&windowMessage{windowId: window.id, lifetimeID: lifetimeID, message: "wails:close"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("callback deadlocked while retiring its window")
	}
}
