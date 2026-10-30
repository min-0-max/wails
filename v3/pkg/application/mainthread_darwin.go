//go:build darwin && !ios && !server

package application

/*
#cgo CFLAGS: -mmacosx-version-min=10.13 -x objective-c
#cgo LDFLAGS: -framework Cocoa

#include "Cocoa/Cocoa.h"
#include "mainthread_darwin.h"

extern void dispatchOnMainThreadCallback(unsigned int);

static void dispatchOnMainThread(unsigned int id) {
	// Go code before the run loop starts runs on the main thread, and the
	// callback runs on the worker while the main thread waits for it, as it
	// ran inline before.
	if ([NSThread isMainThread]) {
		wailsRunOnWorker(^{
			dispatchOnMainThreadCallback(id);
		});
		return;
	}
	// Not dispatch_async to the main queue: AppKit drains that queue and does
	// not re-enter, so once a callback shows a modal dialog or a menu - calls
	// which do not return until they are dismissed - everything queued behind it
	// waits that long too. The run loop keeps turning throughout, so schedule
	// the callback on it instead and it is delivered either way.
	CFRunLoopRef mainLoop = CFRunLoopGetMain();
	CFRunLoopPerformBlock(mainLoop, kCFRunLoopCommonModes, ^{
		wailsRunOnWorker(^{
			dispatchOnMainThreadCallback(id);
		});
	});
	CFRunLoopWakeUp(mainLoop);
}
*/
import "C"

import "runtime"

// The worker runs the Go code of every main-thread step: the closures of
// InvokeSync and InvokeAsync and the callbacks of the native code. See
// mainthread_darwin.h for why no Go code runs on the main thread.
func init() {
	go func() {
		runtime.LockOSThread()
		C.wailsWorkerRun()
	}()
}

// isOnMainThread reports whether the caller runs on the worker, which runs the
// Go code of the main-thread steps, so that work for the main thread runs
// inline there.
func (m *macosApp) isOnMainThread() bool {
	return bool(C.wailsIsWorkerThread())
}

func (m *macosApp) dispatchOnMainThread(id uint) {
	C.dispatchOnMainThread(C.uint(id))
}

//export dispatchOnMainThreadCallback
func dispatchOnMainThreadCallback(callbackID C.uint) {
	mainThreadFunctionStoreLock.Lock()
	id := uint(callbackID)
	fn := mainThreadFunctionStore[id]
	if fn == nil {
		mainThreadFunctionStoreLock.Unlock()
		Fatal("dispatchCallback called with invalid id: %v", id)
		return
	}
	delete(mainThreadFunctionStore, id)
	mainThreadFunctionStoreLock.Unlock()
	fn()
}
