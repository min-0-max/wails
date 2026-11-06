//go:build windows

package edge

import (
	"errors"
	"io"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v3/internal/osmem"
	"golang.org/x/sys/windows"
)

type _IStreamVtbl struct {
	_IUnknownVtbl
	Read  ComProc
	Write ComProc
}

type IStream struct {
	vtbl *_IStreamVtbl
}

// iUnknownReleaseSlot is the index of IUnknown::Release in every COM vtable.
const iUnknownReleaseSlot = 2

// releaseComObject calls IUnknown::Release on the COM object at obj, which
// Windows owns. The first word of the object is the address of its vtable.
// Both are read by copy, so obj never becomes a Go pointer.
func releaseComObject(obj uintptr) {
	vtbl := osmem.Read[uintptr](obj)
	release := osmem.Read[uintptr](vtbl + iUnknownReleaseSlot*unsafe.Sizeof(uintptr(0)))
	syscall.SyscallN(release, obj)
}

func (i *IStream) AddRef() error {
	_, _, err := i.vtbl.AddRef.Call(uintptr(unsafe.Pointer(i)))
	if err != nil && !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}

	return nil
}

func (i *IStream) Release() error {
	_, _, err := i.vtbl.Release.Call(uintptr(unsafe.Pointer(i)))
	if err != nil && !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}

	return nil
}

func (i *IStream) Read(p []byte) (int, error) {
	bufLen := len(p)
	if bufLen == 0 {
		return 0, nil
	}

	var n int
	hr, _, _ := i.vtbl.Read.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(&p[0])),
		uintptr(bufLen),
		uintptr(unsafe.Pointer(&n)),
	)

	switch windows.Handle(hr) {
	case windows.S_OK:
		// The buffer has been completely filled
		return n, nil
	case windows.S_FALSE:
		// The buffer has been filled with less than len data and the stream is EOF
		return n, io.EOF
	default:
		return 0, syscall.Errno(hr)
	}
}
