//go:build darwin && !ios && !server

package application

import (
	"testing"
	"unsafe"
)

// TestMacOptionsWindowShowShowsAStandardWindow covers the show of a standard window
// with MacOptions.WindowShow. The window stays off the screen, because the option
// takes the place of the call that puts it there.
func TestMacOptionsWindowShowShowsAStandardWindow(t *testing.T) {
	if !appKitReady {
		t.Fatal("AppKit is not available")
	}
	var shown []unsafe.Pointer
	app := New(Options{Mac: MacOptions{WindowShow: func(window unsafe.Pointer) { shown = append(shown, window) }}})
	app.impl = newPlatformApp(app)
	go func() {
		for range windowEvents {
		}
	}()
	window := NewWindow(WebviewWindowOptions{Width: 400, Height: 300, Hidden: true, HTML: "<html></html>"})
	window.Run()
	contentOf(t, window)
	window.Show()
	native := window.impl.(*macosWebviewWindow).nsWindow
	if len(shown) != 1 || shown[0] != native {
		t.Fatalf("the window show received %v, want the window %v once", shown, native)
	}
	var visible bool
	InvokeSync(func() { visible = window.impl.isVisible() })
	if visible {
		t.Fatal("the window is on the screen, so Show ordered it itself")
	}
}
