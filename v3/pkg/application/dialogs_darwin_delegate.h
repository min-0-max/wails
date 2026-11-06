//go:build darwin && !ios

#ifndef _DIALOGS_DELEGATE_H_
#define _DIALOGS_DELEGATE_H_

#import <Cocoa/Cocoa.h>
#import <UniformTypeIdentifiers/UTType.h>

// OpenPanel delegate to handle file filtering
@interface OpenPanelDelegate : NSObject <NSOpenSavePanelDelegate>
@property (nonatomic, strong) NSArray *allowedExtensions;
@end

// Gives the panel a new OpenPanelDelegate, which the caller releases once the
// panel closes, and filters the panel's files by the extensions.
void wailsSetOpenPanelFilter(NSOpenPanel *panel, NSArray<NSString *> *extensions);

#endif