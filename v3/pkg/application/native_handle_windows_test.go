//go:build windows

package application

// On Windows the handle is the HWND as an integer, with no conversion.
var _ = func(w Window) uintptr { return w.NativeWindow() }
