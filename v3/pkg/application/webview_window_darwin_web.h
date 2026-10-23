//go:build darwin && !ios && !server

#import <WebKit/WebKit.h>

// WKWebView whose acceptsFirstMouse: answers the acceptsFirstMouseClick property, which the window preferences set
// per webview (MacWebviewPreferences.AcceptFirstMouse). The default is NO, the value of WKWebView.
@interface WailsWebView : WKWebView
@property BOOL acceptsFirstMouseClick;
@end
