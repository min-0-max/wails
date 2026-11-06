package osmem

import (
	"reflect"
	"testing"
	"unsafe"
)

type rect struct{ Left, Top, Right, Bottom int32 }

type message struct {
	Handle uintptr
	Rect   rect
	Points [2]struct{ X, Y int32 }
}

type withPointer struct {
	Handle uintptr
	Data   unsafe.Pointer
}

type withNestedPointer struct {
	Inner [1]struct{ Text *uint16 }
}

func TestOnlyTypesWithoutPointersCanBeCopied(t *testing.T) {
	for _, c := range []struct {
		value any
		free  bool
	}{
		{uint16(0), true},
		{uintptr(0), true},
		{rect{}, true},
		{message{}, true},
		{[0]*int{}, true},
		{withPointer{}, false},
		{withNestedPointer{}, false},
		{(*int)(nil), false},
		{unsafe.Pointer(nil), false},
		{"", false},
		{[]uint16(nil), false},
		{map[int]int(nil), false},
	} {
		typ := reflect.TypeOf(c.value)
		if got := isPointerFree(typ); got != c.free {
			t.Errorf("isPointerFree(%s) = %v, want %v", typ, got, c.free)
		}
	}
}

func TestATypeWithAPointerIsRefused(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("a type with a pointer field was accepted")
		}
	}()
	mustBePointerFree(reflect.TypeFor[withPointer]())
}
