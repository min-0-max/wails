//go:build darwin && !ios

#import <AppKit/AppKit.h>

@interface WebviewDrag : NSView <NSDraggingDestination>
@property unsigned int windowId;
@property unsigned long long callbackLifetimeID;
@end

// The pasteboard types that a view or window registers to receive dropped
// files.
NSArray<NSPasteboardType> *wailsFileDropTypes(void);

// Whether the pasteboard holds files.
BOOL wailsPasteboardHasFiles(NSPasteboard *pasteboard);

// The paths of the files that the pasteboard holds, in order.
NSArray<NSString *> *wailsPasteboardFilePaths(NSPasteboard *pasteboard);
