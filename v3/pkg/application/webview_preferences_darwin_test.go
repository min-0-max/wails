//go:build darwin && !ios && !server

package application

import "testing"

// The accept-first-mouse preference reaches the native window creation only when it is set.
func TestWebviewPreferencesAcceptFirstMouse(t *testing.T) {
	unset := &macosWebviewWindow{parent: &WebviewWindow{}}
	if p := unset.getWebviewPreferences(); bool(p.AcceptFirstMouse) {
		t.Fatal("an unset AcceptFirstMouse must keep the WKWebView default, which refuses the first mouse")
	}

	set := &macosWebviewWindow{parent: &WebviewWindow{options: WebviewWindowOptions{
		Mac: MacWindow{WebviewPreferences: MacWebviewPreferences{AcceptFirstMouse: true}},
	}}}
	if p := set.getWebviewPreferences(); !bool(p.AcceptFirstMouse) {
		t.Fatal("AcceptFirstMouse must reach the native preferences")
	}
}
