//go:build bench && goexperiment.jsonv2 && go1.27

package application

import (
	"context"
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// CallWithJSONv2 unmarshals the arguments with encoding/json/v2.
func (b *BoundMethod) CallWithJSONv2(ctx context.Context, args []json.RawMessage) (result any, err error) {
	defer handlePanic(handlePanicOptions{skipEnd: 5})

	argCount := len(args)
	if b.needsContext {
		argCount++
	}

	if argCount != len(b.Inputs) {
		return nil, &CallError{
			Kind:    TypeError,
			Message: fmt.Sprintf("%s expects %d arguments, got %d", b.FQN, len(b.Inputs), argCount),
		}
	}

	// Convert inputs to values of appropriate type
	callArgs := make([]reflect.Value, argCount)
	base := 0

	if b.needsContext {
		callArgs[0] = reflect.ValueOf(ctx)
		base++
	}

	// Iterate over given arguments - use JSON v2 for unmarshaling
	for index, arg := range args {
		value := reflect.New(b.Inputs[base+index].ReflectType)
		err = jsonv2.Unmarshal(arg, value.Interface())
		if err != nil {
			return nil, &CallError{
				Kind:    TypeError,
				Message: fmt.Sprintf("could not parse argument #%d: %s", index, err),
				Cause:   json.RawMessage(b.marshalError(err)),
			}
		}
		callArgs[base+index] = value.Elem()
	}

	// Do the call
	var callResults []reflect.Value
	if b.Method.Type().IsVariadic() {
		callResults = b.Method.CallSlice(callArgs)
	} else {
		callResults = b.Method.Call(callArgs)
	}

	var nonErrorOutputs = make([]any, 0, len(callResults))
	var errorOutputs []error

	for _, field := range callResults {
		if field.Type() == errorType {
			if field.IsNil() {
				continue
			}
			if errorOutputs == nil {
				errorOutputs = make([]error, 0, len(callResults)-len(nonErrorOutputs))
				nonErrorOutputs = nil
			}
			errorOutputs = append(errorOutputs, field.Interface().(error))
		} else if nonErrorOutputs != nil {
			nonErrorOutputs = append(nonErrorOutputs, field.Interface())
		}
	}

	if len(errorOutputs) > 0 {
		info := make([]json.RawMessage, len(errorOutputs))
		for i, err := range errorOutputs {
			info[i] = b.marshalError(err)
		}

		cerr := &CallError{
			Kind:    RuntimeError,
			Message: errors.Join(errorOutputs...).Error(),
			Cause:   info,
		}
		if len(info) == 1 {
			cerr.Cause = info[0]
		}
		return nil, cerr
	}

	if len(nonErrorOutputs) == 1 {
		result = nonErrorOutputs[0]
	} else if len(nonErrorOutputs) > 1 {
		result = nonErrorOutputs
	}

	return result, nil
}

func BenchmarkCallJSONv2(b *testing.B) {
	_ = NewTestApp(b, Options{})
	bindings := NewBindings(nil, nil)

	service := &benchService{}
	_ = bindings.Add(NewService(service))

	callOptions := &CallOptions{
		MethodName: "github.com/wailsapp/wails/v3/pkg/application.benchService.StringArg",
	}
	method := bindings.Get(callOptions)
	if method == nil {
		b.Fatal("method not found")
	}

	args := []json.RawMessage{json.RawMessage(`"hello world"`)}
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, _ = method.CallWithJSONv2(ctx, args)
	}
}

func BenchmarkCallJSONv2_MultiArgs(b *testing.B) {
	_ = NewTestApp(b, Options{})
	bindings := NewBindings(nil, nil)

	service := &benchService{}
	_ = bindings.Add(NewService(service))

	callOptions := &CallOptions{
		MethodName: "github.com/wailsapp/wails/v3/pkg/application.benchService.MultipleArgs",
	}
	method := bindings.Get(callOptions)
	if method == nil {
		b.Fatal("method not found")
	}

	args := []json.RawMessage{
		json.RawMessage(`"test"`),
		json.RawMessage(`42`),
		json.RawMessage(`true`),
	}
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, _ = method.CallWithJSONv2(ctx, args)
	}
}
