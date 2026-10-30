//go:build darwin && !ios && !server

#ifndef mainthread_darwin_h
#define mainthread_darwin_h

#include <stdbool.h>

// Go code never runs on the main thread. The JavaScriptCore collector scans
// the stack of each thread that holds the lock of a VM from its stack pointer
// to the origin of its stack, and WebKit takes that lock on the main thread of
// the application process. A goroutine stack lies in the Go heap, so a scan of
// the main thread while it runs Go code reads a stack pointer outside the
// thread's stack and the size of the copy wraps.
//
// The work that runs on the main thread in Go runs instead on one worker
// thread, which a goroutine locked to it serves. A step that needs Go code
// hands it to the worker and the main thread waits in C. While it waits, the
// main thread runs only the calls that the worker forwards to it for that
// step, so the step stays one main-thread step and an answer that AppKit or
// WebKit expects from a callback stays synchronous. While the worker waits for
// a forwarded call, it runs the Go work that the main thread hands it, which is
// how a callback reaches Go from a nested run loop of a modal dialog, a menu or
// a drag session.

// wailsRunOnWorker runs task where Go code runs. On the main thread it hands
// the task to the worker and returns when the task has returned, running the
// calls that the worker forwards meanwhile. On any other thread it runs the
// task directly.
void wailsRunOnWorker(void (^task)(void));

// wailsRunOnMain runs call on the main thread for the worker. On the worker it
// forwards the call to the main thread, which waits for the worker, and returns
// when the call has returned, running the Go work that the main thread hands
// over meanwhile. On any other thread it runs the call directly.
void wailsRunOnMain(void (^call)(void));

// wailsRunOnMainSync runs call on the main thread from any thread: directly on
// the main thread, as wailsRunOnMain on the worker, and with dispatch_sync on
// the main queue on any other thread.
void wailsRunOnMainSync(void (^call)(void));

// wailsIsWorkerThread reports whether the calling thread is the worker.
bool wailsIsWorkerThread(void);

// wailsWorkerRun makes the calling thread the worker and serves Go work on it.
// It does not return. The goroutine that calls it must be locked to its thread.
void wailsWorkerRun(void);

// WAILS_ON_MAIN_RESULT evaluates a block that returns a T with wailsRunOnMain
// and yields its result.
#define WAILS_ON_MAIN_RESULT(T, ...) ({ \
	__block T wailsResult; \
	T (^wailsBody)(void) = __VA_ARGS__; \
	wailsRunOnMain(^{ wailsResult = wailsBody(); }); \
	wailsResult; \
})

// WAILS_ON_MAIN_SYNC_RESULT is WAILS_ON_MAIN_RESULT with wailsRunOnMainSync.
#define WAILS_ON_MAIN_SYNC_RESULT(T, ...) ({ \
	__block T wailsResult; \
	T (^wailsBody)(void) = __VA_ARGS__; \
	wailsRunOnMainSync(^{ wailsResult = wailsBody(); }); \
	wailsResult; \
})

#endif
