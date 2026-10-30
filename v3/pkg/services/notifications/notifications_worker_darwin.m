//go:build darwin && !ios && !server

#import "notifications_darwin.h"

// Defined in package application: runs a task where Go code runs, so that a
// callback on the main thread runs its Go code on the worker of package
// application while the main thread waits in C.
extern void wailsRunOnWorker(void (^task)(void));

void notificationsRunGo(void (^task)(void)) {
    wailsRunOnWorker(task);
}
