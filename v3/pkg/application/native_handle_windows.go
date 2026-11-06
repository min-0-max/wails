//go:build windows

package application

// NativeHandle is the native window handle that Window.NativeWindow
// returns. On Windows it is the HWND, a handle that the system owns and
// that is not an address Go may dereference.
type NativeHandle = uintptr
