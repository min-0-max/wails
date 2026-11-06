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
extern void systemTraySetANSILabel(void* nsStatusItem, void* label);

// Gives a status item that never shows in the menu bar an attributed label
// through the tray's label call, and writes the button's title and whether its
// attributed title is the one the tray was given.
static char *wailsProbeStatusItemLabel(const char *text, int *sameAttributes) {
	__block NSStatusItem *statusItem = nil;
	wailsRunOnMainSync(^{
		statusItem = [[[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength] retain];
		statusItem.visible = NO;
	});

	NSDictionary *attributes = @{NSForegroundColorAttributeName: [NSColor systemRedColor]};
	NSAttributedString *want = [[[NSAttributedString alloc] initWithString:[NSString stringWithUTF8String:text] attributes:attributes] autorelease];
	// systemTraySetANSILabel takes ownership of the string it is given.
	NSMutableAttributedString *label = [want mutableCopy];
	systemTraySetANSILabel(statusItem, label);

	// The tray sets the label on the main queue; a call queued after it runs
	// once it has.
	__block char *title = NULL;
	dispatch_sync(dispatch_get_main_queue(), ^{
		NSAttributedString *got = statusItem.button.attributedTitle;
		*sameAttributes = [got isEqualToAttributedString:want];
		title = strdup([[got string] UTF8String]);
		[[NSStatusBar systemStatusBar] removeStatusItem:statusItem];
		[statusItem release];
	});
	return title;
}
*/
import "C"

import "unsafe"

// StatusItemLabel gives a status item that never shows in the menu bar a red
// label through the tray's label call, and returns the title of the item's
// button and whether its attributed title matches the label.
func StatusItemLabel(text string) (title string, sameAttributes bool) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	var same C.int
	result := C.wailsProbeStatusItemLabel(cText, &same)
	defer C.free(unsafe.Pointer(result))
	return C.GoString(result), same != 0
}
