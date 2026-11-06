//go:build darwin && !ios && !server

package mainthreadharness

/*
#cgo CFLAGS: -mmacosx-version-min=12.0 -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>
#include <stdlib.h>
#include <string.h>

// Defined in package application.
extern void wailsRunOnMainSync(void (^call)(void));
extern NSArray<NSPasteboardType> *wailsFileDropTypes(void);
extern BOOL wailsPasteboardHasFiles(NSPasteboard *pasteboard);
extern NSArray<NSString *> *wailsPasteboardFilePaths(NSPasteboard *pasteboard);

static char *wailsJoinLines(NSArray<NSString *> *lines) {
	return strdup([[lines componentsJoinedByString:@"\n"] UTF8String]);
}

// Writes the pasteboard types that the file drop view of a window registers,
// one per line.
static char *wailsProbeFileDropTypes(void) {
	__block char *types = NULL;
	wailsRunOnMainSync(^{
		NSView *view = [[NSClassFromString(@"WebviewDrag") alloc] initWithFrame:NSMakeRect(0, 0, 10, 10)];
		types = wailsJoinLines([view registeredDraggedTypes]);
		[view release];
	});
	return types;
}

// Puts one pasteboard item per URL on a new pasteboard, each holding only the
// URL as NSPasteboardTypeFileURL or, for a URL that is not a file URL, as
// NSPasteboardTypeURL. Writes whether the pasteboard holds files and the paths
// that the drop handlers read from it, one per line.
static char *wailsProbePasteboardFiles(char **urls, int count, int *hasFiles) {
	NSPasteboard *pasteboard = [NSPasteboard pasteboardWithUniqueName];
	[pasteboard clearContents];
	NSMutableArray<NSPasteboardItem *> *items = [NSMutableArray arrayWithCapacity:count];
	for (int i = 0; i < count; i++) {
		NSURL *url = [NSURL URLWithString:[NSString stringWithUTF8String:urls[i]]];
		NSPasteboardItem *item = [[[NSPasteboardItem alloc] init] autorelease];
		[item setString:[url absoluteString] forType:[url isFileURL] ? NSPasteboardTypeFileURL : NSPasteboardTypeURL];
		[items addObject:item];
	}
	[pasteboard writeObjects:items];

	*hasFiles = wailsPasteboardHasFiles(pasteboard);
	char *paths = wailsJoinLines(wailsPasteboardFilePaths(pasteboard) ?: @[]);
	[pasteboard releaseGlobally];
	return paths;
}
*/
import "C"

import (
	"strings"
	"unsafe"
)

// FileDropTypes returns the pasteboard types that the file drop view of a
// window registers.
func FileDropTypes() []string {
	types := C.wailsProbeFileDropTypes()
	defer C.free(unsafe.Pointer(types))
	return lines(C.GoString(types))
}

// PasteboardFiles puts one item per URL on a new pasteboard and returns
// whether the drop handlers see files on it and the paths they read.
func PasteboardFiles(urls []string) (hasFiles bool, paths []string) {
	cURLs := cStrings(urls)
	defer freeCStrings(cURLs)
	var has C.int
	result := C.wailsProbePasteboardFiles(first(cURLs), C.int(len(urls)), &has)
	defer C.free(unsafe.Pointer(result))
	return has != 0, lines(C.GoString(result))
}

func lines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
