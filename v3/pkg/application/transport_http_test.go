package application

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestHTTPTransportCleanupLifecycle(t *testing.T) {
	transport := NewHTTPTransport()
	t.Cleanup(func() { _ = transport.Stop() })

	for cycle := 0; cycle < 100; cycle++ {
		if err := transport.Start(context.Background(), nil); err != nil {
			t.Fatalf("cycle %d: Start: %v", cycle, err)
		}

		transport.cleanupMu.Lock()
		done := transport.cleanupDone
		transport.cleanupMu.Unlock()
		if done == nil {
			t.Fatalf("cycle %d: cleanup goroutine was not started", cycle)
		}
		if err := transport.Start(context.Background(), nil); err != nil {
			t.Fatalf("cycle %d: repeated Start: %v", cycle, err)
		}
		transport.cleanupMu.Lock()
		repeatedDone := transport.cleanupDone
		transport.cleanupMu.Unlock()
		if repeatedDone != done {
			t.Fatalf("cycle %d: repeated Start launched another cleanup goroutine", cycle)
		}

		// Stop is allowed to be called defensively from several shutdown paths.
		// It must close the channel only once and the call that owns the cleanup
		// goroutine must wait for it to finish before returning.
		var stops sync.WaitGroup
		stops.Add(4)
		for range 4 {
			go func() {
				defer stops.Done()
				if err := transport.Stop(); err != nil {
					t.Errorf("cycle %d: Stop: %v", cycle, err)
				}
			}()
		}
		stops.Wait()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatalf("cycle %d: cleanup goroutine did not stop", cycle)
		}
	}
}

// Supply JSON followed by a streamed padding body so the fixture itself does not
// allocate tens of megabytes. Missing object/method keeps accepted requests from
// needing a live application; oversized requests must fail before JSON parsing.
func TestHTTPTransportRuntimeBodyLimit(t *testing.T) {
	for _, size := range []int64{maxRuntimeBodyBytes - 1, maxRuntimeBodyBytes, maxRuntimeBodyBytes + 1} {
		for _, knownLength := range []bool{false, true} {
			t.Run(fmt.Sprintf("size=%d/knownLength=%t", size, knownLength), func(t *testing.T) {
				padding := &runtimeBodyPadding{remaining: size - 2}
				body := io.MultiReader(strings.NewReader("{}"), padding)
				req := httptest.NewRequest(http.MethodPost, "/wails/runtime", body)
				if knownLength {
					req.ContentLength = size
				}
				rec := httptest.NewRecorder()
				transport := NewHTTPTransport()
				transport.handleRuntimeRequest(rec, req)
				want := http.StatusUnprocessableEntity
				if size > maxRuntimeBodyBytes {
					want = http.StatusRequestEntityTooLarge
				}
				if rec.Code != want {
					t.Fatalf("status = %d, want %d: %s", rec.Code, want, rec.Body.String())
				}
			})
		}
	}
}

type runtimeBodyPadding struct{ remaining int64 }

func (p *runtimeBodyPadding) Read(buf []byte) (int, error) {
	if p.remaining == 0 {
		return 0, io.EOF
	}
	n := int(min(int64(len(buf)), p.remaining))
	for i := range buf[:n] {
		buf[i] = ' '
	}
	p.remaining -= int64(n)
	return n, nil
}

func TestHTTPTransportRuntimeBodyStopsReadingAtLimit(t *testing.T) {
	padding := &runtimeBodyPadding{remaining: 2 * maxRuntimeBodyBytes}
	req := httptest.NewRequest(http.MethodPost, "/wails/runtime", padding)
	rec := httptest.NewRecorder()
	NewHTTPTransport().handleRuntimeRequest(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
	if padding.remaining != maxRuntimeBodyBytes-1 {
		t.Fatalf("read beyond limit probe: %d bytes remain", padding.remaining)
	}
}

// An error that the transport returns to the caller reaches the caller only;
// logging it as well writes every failed call twice. A failure that no caller
// receives, such as a failed write of the error response, is still logged.
func TestHTTPTransportErrorIsReturnedNotLogged(t *testing.T) {
	var logged strings.Builder
	logger := slog.New(slog.NewTextHandler(&logged, nil))
	transport := NewHTTPTransport(HTTPTransportWithLogger(logger))
	rec := httptest.NewRecorder()
	transport.handleRuntimeRequest(rec, httptest.NewRequest(http.MethodPost, "/wails/runtime", strings.NewReader("{}")))
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "missing object value") {
		t.Fatalf("response = %d %q", rec.Code, rec.Body.String())
	}
	if logged.Len() != 0 {
		t.Fatalf("the returned error was also logged: %q", logged.String())
	}

	transport.httpError(failingWriter{httptest.NewRecorder()}, fmt.Errorf("returned"))
	if strings.Contains(logged.String(), "returned") || !strings.Contains(logged.String(), "Unable to write error response") {
		t.Fatalf("log after a failed write = %q", logged.String())
	}
}

type failingWriter struct{ *httptest.ResponseRecorder }

func (failingWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("closed") }
