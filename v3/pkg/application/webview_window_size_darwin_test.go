//go:build darwin && !ios && !server

package application

import (
	"testing"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application/internal/mainthreadharness"
)

// contentOf returns the sizes of the content view of a window and of its web
// view and drag view, when the native window exists.
func contentOf(t *testing.T, window *WebviewWindow) []mainthreadharness.Size {
	t.Helper()
	native := window.impl.(*macosWebviewWindow)
	var sizes []mainthreadharness.Size
	for deadline := time.Now().Add(10 * time.Second); len(sizes) < 3 && time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		if native.nsWindow != nil {
			InvokeSync(func() { sizes = mainthreadharness.ContentSizes(native.nsWindow) })
		}
	}
	if len(sizes) < 3 {
		t.Fatalf("the content of the window has %d views, want the content view, the web view and the drag view", len(sizes))
	}
	return sizes
}

// TestTheContentOfAWindowIsItsRequestedSize covers the frames that each kind of
// new window gives its views, and the frames after the window takes a new size.
// The web view resizes with the content view, so a difference between their
// first frames stays for the life of the window, and the page then lays out
// wider and higher than the part of it that the window shows.
func TestTheContentOfAWindowIsItsRequestedSize(t *testing.T) {
	if !appKitReady {
		t.Fatal("AppKit is not available")
	}
	const width, height = 800, 600
	app := New(Options{})
	app.impl = newPlatformApp(app)
	// A running application reads the events of its windows. Without that a new
	// size waits on the first event that it sends.
	go func() {
		for range windowEvents {
		}
	}()
	for name, options := range map[string]WebviewWindowOptions{
		"titled":       {},
		"frameless":    {Frameless: true},
		"panel":        {Mac: MacWindow{WindowClass: MacWindowClassPanel}},
		"inset bar":    {Mac: MacWindow{TitleBar: MacTitleBarHiddenInset}},
		"translucent":  {Mac: MacWindow{Backdrop: MacBackdropTranslucent}},
		"size limited": {MinWidth: 400, MinHeight: 300, MaxWidth: 1600, MaxHeight: 1200},
	} {
		t.Run(name, func(t *testing.T) {
			options.Width, options.Height, options.Hidden, options.EnableFileDrop, options.HTML = width, height, true, true, "<html></html>"
			window := NewWindow(options)
			window.Run()
			sizes := contentOf(t, window)
			for i, size := range sizes {
				if size.Width != width || size.Height != height {
					t.Errorf("view %d of the content is %vx%v, want %dx%d", i, size.Width, size.Height, width, height)
				}
			}
			native := window.impl.(*macosWebviewWindow).nsWindow
			want := mainthreadharness.Size{Width: 1000, Height: 700}
			var resized []mainthreadharness.Size
			InvokeSync(func() {
				mainthreadharness.SetContentSize(native, want)
				resized = mainthreadharness.ContentSizes(native)
			})
			for i, size := range resized {
				if size != want {
					t.Errorf("after the content took the size %vx%v view %d of the content is %vx%v", want.Width, want.Height, i, size.Width, size.Height)
				}
			}
		})
	}
}
