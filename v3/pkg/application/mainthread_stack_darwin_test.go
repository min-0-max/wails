//go:build darwin && !ios && !server

package application

import (
	"testing"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application/internal/mainthreadharness"
)

const (
	// How many times an entry point runs in one batch.
	stackSampleBatch = 200
	// How long the batches of an entry point run at least, so that the sampler
	// reads the main thread during each entry point many times.
	stackSampleSpan = 200 * time.Millisecond
	// How long a dispatched closure runs Go code, so that a closure which runs
	// on the main thread is sampled there.
	stackSampleWork = 100 * time.Microsecond
	// The time limit of each entry point's run.
	stackSamplePhaseLimit = 60 * time.Second
)

// spinGo runs Go code for d.
func spinGo(d time.Duration) int {
	start := time.Now()
	n := 0
	for time.Since(start) < d {
		n++
	}
	return n
}

// drainEntryChannels consumes what the entry points under test send, which an
// application's event loops consume when it runs, until stop is closed.
func drainEntryChannels(stop <-chan struct{}) {
	for {
		select {
		case <-windowEvents:
		case <-applicationEvents:
		case <-windowMessageBuffer:
		case <-menuItemClicked:
		case <-stop:
			return
		}
	}
}

// sampleMainThread runs batches of drive for stackSampleSpan while the sampler reads the main thread and
// fails the test when a sample has its stack pointer outside the main thread's
// stack, which is a sample of Go code on the main thread.
func sampleMainThread(t *testing.T, name string, drive func()) {
	t.Helper()
	sampler, err := mainthreadharness.StartSampler()
	if err != nil {
		t.Fatal(err)
	}
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		for start := time.Now(); time.Since(start) < stackSampleSpan; {
			drive()
		}
	}()
	select {
	case <-finished:
	case <-time.After(stackSamplePhaseLimit):
		t.Fatalf("%s: the entry point did not return within %s", name, stackSamplePhaseLimit)
	}
	samples := sampler.Stop()
	t.Logf("%s: %d samples, %d outside the main thread stack, %d failed reads",
		name, samples.Total, samples.Outside, samples.Failures)
	if samples.Failures > 0 {
		t.Errorf("%s: %d reads of the main thread failed, the last with kernel result %d",
			name, samples.Failures, samples.LastError)
	}
	if samples.Total == 0 {
		t.Fatalf("%s: the sampler took no sample", name)
	}
	if samples.Outside > 0 {
		t.Errorf("%s: the main thread ran Go code in %d of %d samples",
			name, samples.Outside, samples.Total)
	}
}

// TestMainThreadRunsNoGoCode drives the main thread dispatch and the native
// entry points into Go while a thread samples the main thread as the
// JavaScriptCore collector does when it scans a thread that holds the lock of
// a VM. The collector copies the stack from the stack pointer to the origin of
// the thread's stack; a stack pointer on a goroutine stack, which the Go heap
// holds, makes that size wrap and the allocation of the copy end the process.
// So no sample may find the main thread outside its own stack.
func TestMainThreadRunsNoGoCode(t *testing.T) {
	if !appKitReady {
		t.Skip("AppKit is not available, so the main run loop does not turn")
	}

	stop := make(chan struct{})
	defer close(stop)
	go drainEntryChannels(stop)

	t.Run("InvokeSync", func(t *testing.T) {
		withDispatchApplication(t)
		sampleMainThread(t, "InvokeSync", func() {
			for i := 0; i < stackSampleBatch; i++ {
				InvokeSync(func() { spinGo(stackSampleWork) })
			}
		})
	})

	t.Run("InvokeSyncWithResult", func(t *testing.T) {
		withDispatchApplication(t)
		sampleMainThread(t, "InvokeSyncWithResult", func() {
			for i := 0; i < stackSampleBatch; i++ {
				InvokeSyncWithResult(func() int { return spinGo(stackSampleWork) })
			}
		})
	})

	t.Run("InvokeAsync", func(t *testing.T) {
		withDispatchApplication(t)
		sampleMainThread(t, "InvokeAsync", func() {
			done := make(chan struct{}, stackSampleBatch)
			for i := 0; i < stackSampleBatch; i++ {
				InvokeAsync(func() {
					spinGo(stackSampleWork)
					done <- struct{}{}
				})
			}
			for i := 0; i < stackSampleBatch; i++ {
				<-done
			}
		})
	})

	// AppKit calls a delegate from inside a native call that a step makes,
	// and a modal dialog or a menu runs a nested run loop there.
	t.Run("callback from a forwarded call", func(t *testing.T) {
		withDispatchApplication(t)
		sampleMainThread(t, "callback from a forwarded call", func() {
			InvokeSync(func() {
				for i := 0; i < stackSampleBatch; i++ {
					mainthreadharness.CallFromForwardedCall(mainthreadharness.EntryMenuItemClick, 1)
				}
			})
		})
	})

	entries := []struct {
		entry mainthreadharness.Entry
		value uint
	}{
		{mainthreadharness.EntryWindowEvent, 1},
		// An event identifier that no case of processApplicationEvent reads.
		{mainthreadharness.EntryApplicationEvent, 1},
		{mainthreadharness.EntryMessage, 1},
		{mainthreadharness.EntryMenuItemClick, 1},
		{mainthreadharness.EntryAcceleratorFromKeyPress, 0},
		// Without an application, processWindowKeyEquivalent answers false.
		{mainthreadharness.EntryWindowKeyEquivalent, 1},
		{mainthreadharness.EntryOpenURL, 0},
	}
	for _, tc := range entries {
		t.Run(tc.entry.String(), func(t *testing.T) {
			sampleMainThread(t, tc.entry.String(), func() {
				mainthreadharness.CallOnMain(tc.entry, tc.value, stackSampleBatch)
			})
		})
	}
}

// withDispatchApplication gives InvokeSync and InvokeAsync the application
// whose platform code dispatches to the main thread, for the duration of the
// test.
func withDispatchApplication(t *testing.T) {
	t.Helper()
	previous := globalApplication
	globalApplication = &App{impl: &macosApp{}}
	t.Cleanup(func() { globalApplication = previous })
}
