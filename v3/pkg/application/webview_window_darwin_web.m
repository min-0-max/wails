//go:build darwin && !ios && !server

#import "webview_window_darwin_web.h"

@implementation WailsWebView

// A click on an inactive window is delivered to the webview only when it accepts the first mouse. The property is
// set when the window is created, so each webview keeps its own choice.
- (BOOL)acceptsFirstMouse:(NSEvent *)event {
    return self.acceptsFirstMouseClick;
}

@end
