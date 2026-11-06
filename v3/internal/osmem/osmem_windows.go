//go:build windows

package osmem

import (
	"reflect"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modntdll          = windows.NewLazySystemDLL("ntdll.dll")
	modkernel32       = windows.NewLazySystemDLL("kernel32.dll")
	procRtlMoveMemory = modntdll.NewProc("RtlMoveMemory")
	procLstrlenW      = modkernel32.NewProc("lstrlenW")
)

// Read returns a copy of the value of type T at addr, which Windows owns.
func Read[T any](addr uintptr) T {
	mustBePointerFree(reflect.TypeFor[T]())
	var v T
	if size := unsafe.Sizeof(v); size > 0 {
		procRtlMoveMemory.Call(uintptr(unsafe.Pointer(&v)), addr, size)
	}
	return v
}

// Write copies v to addr, which Windows owns.
func Write[T any](addr uintptr, v T) {
	mustBePointerFree(reflect.TypeFor[T]())
	if size := unsafe.Sizeof(v); size > 0 {
		procRtlMoveMemory.Call(addr, uintptr(unsafe.Pointer(&v)), size)
	}
}

// ReadSlice returns a copy of the n values of type T that start at addr,
// which Windows owns.
func ReadSlice[T any](addr uintptr, n int) []T {
	mustBePointerFree(reflect.TypeFor[T]())
	s := make([]T, n)
	if n > 0 {
		var v T
		procRtlMoveMemory.Call(uintptr(unsafe.Pointer(&s[0])), addr, uintptr(n)*unsafe.Sizeof(v))
	}
	return s
}

// WriteSlice copies the values of s to addr, which Windows owns.
func WriteSlice[T any](addr uintptr, s []T) {
	mustBePointerFree(reflect.TypeFor[T]())
	if len(s) > 0 {
		var v T
		procRtlMoveMemory.Call(addr, uintptr(unsafe.Pointer(&s[0])), uintptr(len(s))*unsafe.Sizeof(v))
	}
}

// UTF16Len returns the number of UTF-16 code units before the NUL that ends
// the string at addr.
func UTF16Len(addr uintptr) int {
	if addr == 0 {
		return 0
	}
	n, _, _ := procLstrlenW.Call(addr)
	return int(n)
}

// UTF16String returns a copy of the NUL-terminated UTF-16 string at addr,
// which Windows owns. A zero addr gives the empty string.
func UTF16String(addr uintptr) string {
	return windows.UTF16ToString(ReadSlice[uint16](addr, UTF16Len(addr)))
}
