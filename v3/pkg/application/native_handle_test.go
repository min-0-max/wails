package application

import "testing"

func TestNativeWindowWithoutANativeWindowIsTheZeroHandle(t *testing.T) {
	var zero NativeHandle
	if got := (&WebviewWindow{}).NativeWindow(); got != zero {
		t.Fatalf("NativeWindow() = %v for a window without a native window, want the zero handle", got)
	}
}
