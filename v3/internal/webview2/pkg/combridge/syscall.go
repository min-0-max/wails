//go:build windows

package combridge

import (
	"unsafe"

	"github.com/wailsapp/wails/v3/internal/osmem"
	"golang.org/x/sys/windows"
)

var (
	modkernel32     = windows.NewLazySystemDLL("kernel32.dll")
	procGlobalAlloc = modkernel32.NewProc("GlobalAlloc")
	procGlobalFree  = modkernel32.NewProc("GlobalFree")

	uintptrSize = unsafe.Sizeof(uintptr(0))
)

// allocUintptrObject copies values into memory that Windows allocates and
// returns its address, which COM clients read as an interface or a vtable.
func allocUintptrObject(values []uintptr) uintptr {
	v := globalAlloc(uintptr(len(values)) * uintptrSize)
	osmem.WriteSlice(v, values)
	return v
}

func globalAlloc(dwBytes uintptr) uintptr {
	ret, _, _ := procGlobalAlloc.Call(uintptr(0), dwBytes)
	if ret == 0 {
		panic("globalAlloc failed")
	}

	return ret
}

func globalFree(data uintptr) {
	ret, _, _ := procGlobalFree.Call(data)
	if ret != 0 {
		panic("globalFree failed")
	}
}
