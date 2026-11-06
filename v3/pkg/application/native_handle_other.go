//go:build !windows

package application

import "unsafe"

// NativeHandle is the native window handle that Window.NativeWindow
// returns. Outside Windows it is the address of the platform window object:
// the NSWindow on macOS and iOS, the GtkWindow on Linux.
type NativeHandle = unsafe.Pointer
