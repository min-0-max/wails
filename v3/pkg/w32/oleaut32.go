//go:build windows

/*
 * Copyright (C) 2019 Tad Vizbaras. All Rights Reserved.
 * Copyright (C) 2010-2012 The W32 Authors. All Rights Reserved.
 */
package w32

import (
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/wailsapp/wails/v3/internal/osmem"
)

var (
	modoleaut32 = syscall.NewLazyDLL("oleaut32")

	procVariantInit        = modoleaut32.NewProc("VariantInit")
	procSysAllocString     = modoleaut32.NewProc("SysAllocString")
	procSysFreeString      = modoleaut32.NewProc("SysFreeString")
	procSysStringLen       = modoleaut32.NewProc("SysStringLen")
	procCreateDispTypeInfo = modoleaut32.NewProc("CreateDispTypeInfo")
	procCreateStdDispatch  = modoleaut32.NewProc("CreateStdDispatch")
)

func VariantInit(v *VARIANT) {
	hr, _, _ := procVariantInit.Call(uintptr(unsafe.Pointer(v)))
	if hr != 0 {
		panic("Invoke VariantInit error.")
	}
	return
}

// SysAllocString returns a BSTR that holds v. A BSTR is the address of a
// string that oleaut32 owns, so it stays a uintptr in Go and is read by copy.
func SysAllocString(v string) uintptr {
	bstr, _, _ := procSysAllocString.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(v))))
	return bstr
}

// SysFreeString frees the BSTR v. The Windows function returns nothing.
func SysFreeString(v uintptr) {
	procSysFreeString.Call(v)
}

func SysStringLen(v uintptr) uint {
	l, _, _ := procSysStringLen.Call(v)
	return uint(l)
}

// bstrToString returns a copy of the BSTR v. A zero v gives the empty string.
func bstrToString(v uintptr) string {
	if v == 0 {
		return ""
	}
	return string(utf16.Decode(osmem.ReadSlice[uint16](v, int(SysStringLen(v)))))
}
