//go:build darwin && !ios && !server

#import "dialogs_darwin_delegate.h"

// Override shouldEnableURL
@implementation OpenPanelDelegate
- (BOOL)panel:(id)sender shouldEnableURL:(NSURL *)url {
    if (url == nil) {
        return NO;
    }

    NSFileManager *fileManager = [NSFileManager defaultManager];
    BOOL isDirectory = NO;
    if ([fileManager fileExistsAtPath:url.path isDirectory:&isDirectory] && isDirectory) {
        return YES;
    }

    // If no extensions specified, allow all files
    if (self.allowedExtensions == nil || [self.allowedExtensions count] == 0) {
        return YES;
    }

    NSString *filename = [[url lastPathComponent] lowercaseString];
    if (filename == nil || [filename isEqualToString:@""]) {
        return NO;
    }

    // Check if the extension is in our allowed list (case insensitive)
    for (NSString *allowedExt in self.allowedExtensions) {
        NSString *allowed = [allowedExt lowercaseString];
        if ([allowed length] == 0) {
            continue;
        }
        if ([filename hasSuffix:[@"." stringByAppendingString:allowed]]) {
            return YES;
        }
    }

    return NO;
}

@end

void wailsSetOpenPanelFilter(NSOpenPanel *panel, NSArray<NSString *> *extensions) {
    OpenPanelDelegate *delegate = [[OpenPanelDelegate alloc] init];
    [panel setDelegate:delegate];
    delegate.allowedExtensions = extensions;

    // The panel limits its files to content types only when every extension
    // has one. An extension without one, such as tar.gz, leaves the panel
    // without a type restriction, and the delegate filters by the suffix.
    NSMutableArray<UTType *> *types = [NSMutableArray arrayWithCapacity:[extensions count]];
    for (NSString *extension in extensions) {
        UTType *type = [UTType typeWithFilenameExtension:extension];
        if (type == nil) {
            return;
        }
        [types addObject:type];
    }
    [panel setAllowedContentTypes:types];
}
