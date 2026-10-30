//go:build darwin && !ios && server

#import "notifications_darwin.h"

// A server build of package application shows no webview and has no worker for
// main-thread steps, so the Go handler runs on the calling thread.
void notificationsRunGo(void (^task)(void)) {
    task();
}
