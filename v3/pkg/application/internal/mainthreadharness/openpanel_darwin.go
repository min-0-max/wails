//go:build darwin && !ios && !server

package mainthreadharness

/*
#cgo CFLAGS: -mmacosx-version-min=12.0 -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework UniformTypeIdentifiers

#import <Cocoa/Cocoa.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>
#include <stdlib.h>
#include <string.h>

// Defined in package application.
extern void wailsRunOnMainSync(void (^call)(void));
extern void wailsSetOpenPanelFilter(NSOpenPanel *panel, NSArray<NSString *> *extensions);

static NSArray<NSString *> *wailsStrings(char **values, int count) {
	NSMutableArray<NSString *> *strings = [NSMutableArray arrayWithCapacity:count];
	for (int i = 0; i < count; i++) {
		[strings addObject:[NSString stringWithUTF8String:values[i]]];
	}
	return strings;
}

// Filters a panel that is never shown by the extensions, and writes the
// identifiers of the content types the panel allows, one per line, whether
// the panel holds a file type restriction, and, for each path, whether the
// panel's delegate enables it.
static char *wailsProbeOpenPanelFilter(char **extensions, int extensionCount, char **paths, int pathCount, int *restricted, int *enabled) {
	__block char *types = NULL;
	wailsRunOnMainSync(^{
		NSOpenPanel *panel = [NSOpenPanel openPanel];
		wailsSetOpenPanelFilter(panel, wailsStrings(extensions, extensionCount));

		NSMutableArray<NSString *> *identifiers = [NSMutableArray array];
		for (UTType *type in panel.allowedContentTypes) {
			[identifiers addObject:type.identifier];
		}
		types = strdup([[identifiers componentsJoinedByString:@"\n"] UTF8String]);

		// allowedFileTypes is the only public view of a restriction made with
		// file types that have no content type: allowedContentTypes reports
		// such a restriction as an empty list, the same as no restriction.
		// nil is the only value that allows every file.
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
		*restricted = panel.allowedFileTypes != nil;
#pragma clang diagnostic pop

		id<NSOpenSavePanelDelegate> delegate = panel.delegate;
		for (int i = 0; i < pathCount; i++) {
			NSURL *url = [NSURL fileURLWithPath:[NSString stringWithUTF8String:paths[i]]];
			enabled[i] = [delegate panel:panel shouldEnableURL:url];
		}
		[panel setDelegate:nil];
		[(id)delegate release];
	});
	return types;
}
*/
import "C"

import (
	"strings"
	"unsafe"
)

// OpenPanelFilter is what an open dialog's panel does with a file filter.
type OpenPanelFilter struct {
	// ContentTypes holds the identifiers of the content types the panel allows.
	ContentTypes []string
	// Restricted reports whether the panel limits its files to file types.
	Restricted bool
	// Enabled reports, for each path, whether the panel's delegate enables it.
	Enabled []bool
}

// ProbeOpenPanelFilter applies the file filter of an open dialog, built from
// the extensions, to a panel that is never shown, and reads it back.
func ProbeOpenPanelFilter(extensions, paths []string) OpenPanelFilter {
	cExtensions := cStrings(extensions)
	defer freeCStrings(cExtensions)
	cPaths := cStrings(paths)
	defer freeCStrings(cPaths)

	var restricted C.int
	enabled := make([]C.int, len(paths)+1)
	types := C.wailsProbeOpenPanelFilter(first(cExtensions), C.int(len(extensions)), first(cPaths), C.int(len(paths)), &restricted, &enabled[0])
	defer C.free(unsafe.Pointer(types))

	filter := OpenPanelFilter{Restricted: restricted != 0}
	if s := C.GoString(types); s != "" {
		filter.ContentTypes = strings.Split(s, "\n")
	}
	for _, e := range enabled[:len(paths)] {
		filter.Enabled = append(filter.Enabled, e != 0)
	}
	return filter
}

func cStrings(values []string) []*C.char {
	result := make([]*C.char, len(values))
	for i, v := range values {
		result[i] = C.CString(v)
	}
	return result
}

func freeCStrings(values []*C.char) {
	for _, v := range values {
		C.free(unsafe.Pointer(v))
	}
}

func first(values []*C.char) **C.char {
	if len(values) == 0 {
		return nil
	}
	return &values[0]
}
