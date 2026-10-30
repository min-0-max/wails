//go:build darwin && !ios && !server

package application

/*
#cgo CFLAGS: -mmacosx-version-min=10.13 -x objective-c
#cgo LDFLAGS: -framework Cocoa -mmacosx-version-min=10.13

#include "mainthread_darwin.h"
#import <Cocoa/Cocoa.h>
#import <stdlib.h>

bool setClipboardText(const char* text) {
	return WAILS_ON_MAIN_RESULT(bool, ^bool (void) {
		NSPasteboard *pasteBoard = [NSPasteboard generalPasteboard];
		NSError *error = nil;
		NSString *string = [NSString stringWithUTF8String:text];
		[pasteBoard clearContents];
		return [pasteBoard setString:string forType:NSPasteboardTypeString];
	});
}

const char* getClipboardText() {
	return WAILS_ON_MAIN_RESULT(const char*, ^const char* (void) {
		NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
		NSString *text = [pasteboard stringForType:NSPasteboardTypeString];
		return [text UTF8String];
	});
}

*/
import "C"
import (
	"sync"
	"unsafe"
)

var clipboardLock sync.RWMutex

type macosClipboard struct{}

func (m macosClipboard) setText(text string) bool {
	clipboardLock.Lock()
	defer clipboardLock.Unlock()
	cText := C.CString(text)
	success := C.setClipboardText(cText)
	C.free(unsafe.Pointer(cText))
	return bool(success)
}

func (m macosClipboard) text() (string, bool) {
	clipboardLock.RLock()
	defer clipboardLock.RUnlock()
	clipboardText := C.getClipboardText()
	result := C.GoString(clipboardText)
	return result, true
}

func newClipboardImpl() *macosClipboard {
	return &macosClipboard{}
}
