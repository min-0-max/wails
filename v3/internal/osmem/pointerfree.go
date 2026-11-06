// Package osmem reads and writes memory that the operating system owns.
//
// Go code keeps the address of such memory as a uintptr and never turns it
// into a Go pointer. The contents move between that memory and Go values by
// copy, through an operating system call that takes both addresses as
// integers. The only conversion left on the Go side is a Go pointer to a
// uintptr in the argument list of that call, which the unsafe.Pointer rules
// allow.
//
// A value copied out of OS memory must hold no pointer: the bytes would
// become a Go pointer that Go never made. The copy functions check this for
// every type they are given.
package osmem

import (
	"fmt"
	"reflect"
	"sync"
)

var pointerFree sync.Map // reflect.Type -> bool

// mustBePointerFree panics when values of t can hold a pointer.
func mustBePointerFree(t reflect.Type) {
	if known, ok := pointerFree.Load(t); ok {
		if !known.(bool) {
			panic(fmt.Sprintf("osmem: %s holds a pointer and cannot be copied from or to OS memory", t))
		}
		return
	}
	free := isPointerFree(t)
	pointerFree.Store(t, free)
	if !free {
		panic(fmt.Sprintf("osmem: %s holds a pointer and cannot be copied from or to OS memory", t))
	}
}

// isPointerFree reports whether values of t hold no pointer, so that their
// bytes can come from memory that Go does not manage.
func isPointerFree(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	case reflect.Array:
		return t.Len() == 0 || isPointerFree(t.Elem())
	case reflect.Struct:
		for i := range t.NumField() {
			if !isPointerFree(t.Field(i).Type) {
				return false
			}
		}
		return true
	}
	// Pointers, unsafe.Pointer, strings, slices, maps, channels, functions
	// and interfaces all hold a pointer.
	return false
}
