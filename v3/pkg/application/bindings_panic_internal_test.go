package application

import (
	"context"
	"errors"
	"testing"
)

type panicTestService struct{}

func (panicTestService) Boom() { panic("kaboom") }

// TestBoundMethodPanicInvokesPanicHandler guards the second half of #5037:
// with a custom PanicHandler registered, a panic in a bound method must both
// invoke the handler (telemetry preserved) AND still reject the call with a
// *CallError. The pre-fix code returned (nil, nil) on this path, silently
// resolving the frontend promise with null.
//
// The handler is installed through the options of a new application, which
// the test discards when it ends.
func TestBoundMethodPanicInvokesPanicHandler(t *testing.T) {
	var handled *PanicDetails
	NewTestApp(t, Options{PanicHandler: func(d *PanicDetails) { handled = d }})

	bindings := NewBindings(nil, nil)
	if err := bindings.Add(NewService(&panicTestService{})); err != nil {
		t.Fatalf("bindings.Add() error = %v", err)
	}

	var bound *BoundMethod
	for _, m := range bindings.boundMethods {
		bound = m
	}
	if bound == nil {
		t.Fatal("no bound method registered")
	}

	result, err := bound.Call(context.TODO(), nil)

	// (1) the handler is invoked with the panic details
	if handled == nil {
		t.Error("PanicHandler was not invoked")
	} else if handled.Error == nil {
		t.Error("PanicDetails.Error is nil")
	}

	// (2) the call still rejects with a RuntimeError CallError
	if result != nil {
		t.Errorf("result = %v, expected nil", result)
	}
	var cerr *CallError
	if !errors.As(err, &cerr) {
		t.Fatalf("err = %#v, expected *CallError", err)
	}
	if cerr.Kind != RuntimeError {
		t.Errorf("err.Kind = %q, expected RuntimeError", cerr.Kind)
	}
}
