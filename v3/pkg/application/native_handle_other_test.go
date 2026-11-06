//go:build !windows

package application

import "unsafe"

// Code outside Windows that keeps the handle in an unsafe.Pointer compiles
// unchanged.
var _ = func(w Window) unsafe.Pointer { return w.NativeWindow() }
