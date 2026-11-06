package application

import "testing"

// NewTestApp calls New for one test or benchmark and leaves no application
// behind. New keeps the first application of the process and returns it to
// every later call, so an application that one test leaves would answer the
// next test's New with the options of the first. It is exported for the
// external test package.
func NewTestApp(tb testing.TB, options Options) *App {
	tb.Helper()
	if globalApplication != nil {
		tb.Fatal("an earlier test left its application, so New would return it in place of a new one")
	}
	tb.Cleanup(func() { globalApplication = nil })
	return New(options)
}
