//go:build darwin && !ios && !server

package application

import (
	"testing"

	"github.com/wailsapp/wails/v3/internal/optional"
)

// The accept-first-mouse preference reaches the native window creation only when it is set.
func TestWebviewPreferencesAcceptFirstMouse(t *testing.T) {
	unset := &macosWebviewWindow{parent: &WebviewWindow{}}
	if p := unset.getWebviewPreferences(); p.AcceptFirstMouse != nil {
		t.Fatal("an unset AcceptFirstMouse must pass NULL")
	}

	for _, want := range []bool{true, false} {
		w := &macosWebviewWindow{parent: &WebviewWindow{options: WebviewWindowOptions{
			Mac: MacWindow{WebviewPreferences: MacWebviewPreferences{AcceptFirstMouse: optional.NewBool(want)}},
		}}}
		p := w.getWebviewPreferences()
		if p.AcceptFirstMouse == nil || bool(*p.AcceptFirstMouse) != want {
			t.Fatalf("AcceptFirstMouse %v must reach the native preferences", want)
		}
	}
}
