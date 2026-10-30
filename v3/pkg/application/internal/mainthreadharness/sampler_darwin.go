//go:build darwin && !ios && !server

package mainthreadharness

/*
#cgo CFLAGS: -mmacosx-version-min=10.13 -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>
#include <mach/mach.h>
#include <pthread.h>
#include <stdatomic.h>
#include <stdint.h>
#include <stdlib.h>
#include <unistd.h>

// The sampler reads the main thread the way the JavaScriptCore collector does
// in captureStack: it suspends the thread, reads its registers and resumes it.
// A sample counts as outside when the stack pointer is not within the stack
// that the system allocated for the main thread, which is the case whenever
// the main thread runs Go code on a goroutine stack.
typedef struct {
	atomic_int running;
	atomic_long samples;
	atomic_long outside;
	atomic_long failures;
	atomic_int lastError;
	mach_port_t target;
	uintptr_t low;
	uintptr_t high;
	pthread_t thread;
} wails_sampler_t;

static uintptr_t wailsSampleStackPointer(mach_port_t target, kern_return_t *result) {
#if defined(__arm64__)
	arm_thread_state64_t state;
	mach_msg_type_number_t count = ARM_THREAD_STATE64_COUNT;
	*result = thread_get_state(target, ARM_THREAD_STATE64, (thread_state_t)&state, &count);
	return (uintptr_t)arm_thread_state64_get_sp(state);
#elif defined(__x86_64__)
	x86_thread_state64_t state;
	mach_msg_type_number_t count = x86_THREAD_STATE64_COUNT;
	*result = thread_get_state(target, x86_THREAD_STATE64, (thread_state_t)&state, &count);
	return (uintptr_t)state.__rsp;
#else
#error "unsupported architecture"
#endif
}

static void* wailsSamplerMain(void *handle) {
	wails_sampler_t *sampler = (wails_sampler_t *)handle;
	while (atomic_load(&sampler->running)) {
		kern_return_t result = thread_suspend(sampler->target);
		if (result != KERN_SUCCESS) {
			atomic_fetch_add(&sampler->failures, 1);
			atomic_store(&sampler->lastError, result);
		} else {
			uintptr_t sp = wailsSampleStackPointer(sampler->target, &result);
			thread_resume(sampler->target);
			if (result != KERN_SUCCESS) {
				atomic_fetch_add(&sampler->failures, 1);
				atomic_store(&sampler->lastError, result);
			} else {
				atomic_fetch_add(&sampler->samples, 1);
				if (sp < sampler->low || sp > sampler->high) {
					atomic_fetch_add(&sampler->outside, 1);
				}
			}
		}
		usleep(20);
	}
	return NULL;
}

// Image constructors run on the main thread before main, so this records the
// main thread with public interfaces only.
static pthread_t wailsMainThread;

__attribute__((constructor)) static void wailsRecordMainThread(void) {
	wailsMainThread = pthread_self();
}

static void* wailsSamplerStart(int *error) {
	wails_sampler_t *sampler = calloc(1, sizeof(wails_sampler_t));
	pthread_t main = wailsMainThread;
	sampler->target = pthread_mach_thread_np(main);
	sampler->high = (uintptr_t)pthread_get_stackaddr_np(main);
	sampler->low = sampler->high - pthread_get_stacksize_np(main);
	atomic_store(&sampler->running, 1);
	*error = pthread_create(&sampler->thread, NULL, wailsSamplerMain, sampler);
	if (*error != 0) {
		free(sampler);
		return NULL;
	}
	return sampler;
}

static void wailsSamplerStop(void *handle, long *samples, long *outside, long *failures, int *lastError) {
	wails_sampler_t *sampler = (wails_sampler_t *)handle;
	atomic_store(&sampler->running, 0);
	pthread_join(sampler->thread, NULL);
	*samples = atomic_load(&sampler->samples);
	*outside = atomic_load(&sampler->outside);
	*failures = atomic_load(&sampler->failures);
	*lastError = atomic_load(&sampler->lastError);
	free(sampler);
}

// The native code of package application reaches Go through these entry
// points. The harness calls them by name from the main thread, as AppKit and
// WebKit do.
extern void processWindowEvent(unsigned int windowID, unsigned int eventID);
extern void processApplicationEvent(unsigned int eventID, void *data);
extern void processMessage(unsigned int windowID, unsigned long long lifetimeID, const char *message, const char *origin, bool isMainFrame);
extern void processMenuItemClick(unsigned int menuID);
extern char* acceleratorFromKeyPress(unsigned short keyCode, unsigned long modifiers, unsigned int character, int hasCharacter);
extern bool processWindowKeyEquivalent(unsigned int windowID, unsigned long long lifetimeID, const char *accelerator);
extern void HandleOpenURL(char *url);

enum {
	wailsEntryWindowEvent = 0,
	wailsEntryApplicationEvent,
	wailsEntryMessage,
	wailsEntryMenuItemClick,
	wailsEntryAcceleratorFromKeyPress,
	wailsEntryWindowKeyEquivalent,
	wailsEntryOpenURL,
};

static void wailsCallEntry(int entry, unsigned int value) {
	switch (entry) {
	case wailsEntryWindowEvent:
		processWindowEvent(value, value);
		break;
	case wailsEntryApplicationEvent:
		processApplicationEvent(value, NULL);
		break;
	case wailsEntryMessage:
		processMessage(value, 1, "harness message", "wails://wails", true);
		break;
	case wailsEntryMenuItemClick:
		processMenuItemClick(value);
		break;
	case wailsEntryAcceleratorFromKeyPress:
		free(acceleratorFromKeyPress(0, 1 << 20, 'a', 1));
		break;
	case wailsEntryWindowKeyEquivalent:
		processWindowKeyEquivalent(value, 1, "cmd+a");
		break;
	case wailsEntryOpenURL:
		HandleOpenURL("wails://harness");
		break;
	}
}

// Defined in package application: runs a call on the main thread for the
// worker that runs the Go code of main-thread steps.
extern void wailsRunOnMain(void (^call)(void));

// Calls an entry point from a call that a main-thread step forwards to the
// main thread, as AppKit calls a delegate from inside a native call.
static void wailsCallEntryFromForwardedCall(int entry, unsigned int value) {
	wailsRunOnMain(^{
		wailsCallEntry(entry, value);
	});
}

// Calls an entry point count times from the main run loop and returns when
// the last call has returned.
static void wailsCallEntryOnMain(int entry, unsigned int value, int count) {
	dispatch_semaphore_t done = dispatch_semaphore_create(0);
	CFRunLoopRef mainLoop = CFRunLoopGetMain();
	CFRunLoopPerformBlock(mainLoop, kCFRunLoopCommonModes, ^{
		for (int i = 0; i < count; i++) {
			wailsCallEntry(entry, value);
		}
		dispatch_semaphore_signal(done);
	});
	CFRunLoopWakeUp(mainLoop);
	dispatch_semaphore_wait(done, DISPATCH_TIME_FOREVER);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Entry names an entry point of the native code into Go.
type Entry int

const (
	EntryWindowEvent Entry = iota
	EntryApplicationEvent
	EntryMessage
	EntryMenuItemClick
	EntryAcceleratorFromKeyPress
	EntryWindowKeyEquivalent
	EntryOpenURL
)

// String returns the C name of the entry point.
func (e Entry) String() string {
	switch e {
	case EntryWindowEvent:
		return "processWindowEvent"
	case EntryApplicationEvent:
		return "processApplicationEvent"
	case EntryMessage:
		return "processMessage"
	case EntryMenuItemClick:
		return "processMenuItemClick"
	case EntryAcceleratorFromKeyPress:
		return "acceleratorFromKeyPress"
	case EntryWindowKeyEquivalent:
		return "processWindowKeyEquivalent"
	case EntryOpenURL:
		return "HandleOpenURL"
	}
	return fmt.Sprintf("Entry(%d)", int(e))
}

// CallOnMain calls the entry point count times from the main run loop, with
// value as its window, event or menu identifier, and returns when the last
// call has returned.
func CallOnMain(entry Entry, value uint, count int) {
	C.wailsCallEntryOnMain(C.int(entry), C.uint(value), C.int(count))
}

// CallFromForwardedCall calls the entry point from a call that the running
// main-thread step forwards to the main thread. It must be called from a
// main-thread step, such as a closure of InvokeSync.
func CallFromForwardedCall(entry Entry, value uint) {
	C.wailsCallEntryFromForwardedCall(C.int(entry), C.uint(value))
}

// Sampler suspends the main thread repeatedly and reads its stack pointer.
type Sampler struct {
	handle unsafe.Pointer
}

// Samples is the result of a sampler run.
type Samples struct {
	// Total counts the samples taken.
	Total int64
	// Outside counts the samples whose stack pointer was outside the stack of
	// the main thread.
	Outside int64
	// Failures counts the suspensions or register reads that failed, and
	// LastError is the kernel result of the last of them.
	Failures  int64
	LastError int
}

// StartSampler starts a thread that samples the main thread until Stop.
func StartSampler() (*Sampler, error) {
	var errno C.int
	handle := C.wailsSamplerStart(&errno)
	if handle == nil {
		return nil, fmt.Errorf("pthread_create of the sampler thread failed with %d", int(errno))
	}
	return &Sampler{handle: handle}, nil
}

// Stop ends the sampler thread and returns its counts.
func (s *Sampler) Stop() Samples {
	var samples, outside, failures C.long
	var lastError C.int
	C.wailsSamplerStop(s.handle, &samples, &outside, &failures, &lastError)
	return Samples{
		Total:     int64(samples),
		Outside:   int64(outside),
		Failures:  int64(failures),
		LastError: int(lastError),
	}
}
