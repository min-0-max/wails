//go:build darwin && !ios && !server

package application

import (
	"testing"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application/internal/mainthreadharness"
)

// TestTheContentOfAWindowIsItsRequestedSize covers the frames that a new window
// gives its views. The web view resizes with the content view, so a difference
// between their first frames stays for the life of the window, and the page
// then lays out wider and higher than the part of it that the window shows.
func TestTheContentOfAWindowIsItsRequestedSize(t *testing.T) {
	if !appKitReady {
		t.Fatal("AppKit is not available")
	}
	const width, height = 800, 600
	app := New(Options{})
	app.impl = newPlatformApp(app)
	window := NewWindow(WebviewWindowOptions{Width: width, Height: height, Hidden: true, EnableFileDrop: true, HTML: "<html></html>"})
	window.Run()
	native := window.impl.(*macosWebviewWindow)
	var sizes []mainthreadharness.Size
	for deadline := time.Now().Add(10 * time.Second); len(sizes) < 3 && time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		if native.nsWindow != nil {
			sizes = mainthreadharness.ContentSizes(native.nsWindow)
		}
	}
	if len(sizes) < 3 {
		t.Fatalf("the content of the window has %d views, want the content view, the web view and the drag view", len(sizes))
	}
	for i, size := range sizes {
		if size.Width != width || size.Height != height {
			t.Errorf("view %d of the content is %vx%v, want %dx%d", i, size.Width, size.Height, width, height)
		}
	}
}
