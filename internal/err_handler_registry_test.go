package internal_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/stretchr/testify/assert"
)

func Test_NewErrorHandlingRegistry_PanicWhenNotFunc(t *testing.T) {
	nonFuncTypes := []struct {
		name string
		test func()
	}{
		{"int", func() { _ = internal.NewErrorHandlingRegistry[int]() }},
		{"string", func() { _ = internal.NewErrorHandlingRegistry[string]() }},
		{"map[int]int", func() { _ = internal.NewErrorHandlingRegistry[map[int]int]() }},
		{"bool", func() { _ = internal.NewErrorHandlingRegistry[bool]() }},
		{"float64", func() { _ = internal.NewErrorHandlingRegistry[float64]() }},
		{"complex128", func() { _ = internal.NewErrorHandlingRegistry[complex128]() }},
		{"struct{}", func() { _ = internal.NewErrorHandlingRegistry[struct{}]() }},
		{"[]int", func() { _ = internal.NewErrorHandlingRegistry[[]int]() }},
		{"[3]string", func() { _ = internal.NewErrorHandlingRegistry[[3]string]() }},
		{"chan int", func() { _ = internal.NewErrorHandlingRegistry[chan int]() }},
		{"any", func() { _ = internal.NewErrorHandlingRegistry[any]() }},
		{"*int", func() { _ = internal.NewErrorHandlingRegistry[*int]() }},
		{"uintptr", func() { _ = internal.NewErrorHandlingRegistry[uintptr]() }},
		{"func(int)", func() { _ = internal.NewErrorHandlingRegistry[func(int)]() }},
	}

	for _, tc := range nonFuncTypes {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotPanics(t, tc.test)
		})
	}
}

func Test_ValidFunctionTypes(t *testing.T) {
	funcTypes := []struct {
		name string
		test func()
	}{
		{"func(func())", func() { _ = internal.NewErrorHandlingRegistry[func()]() }},
		{"func(error)", func() { _ = internal.NewErrorHandlingRegistry[error]() }},
		{"func(func(string) int)", func() { _ = internal.NewErrorHandlingRegistry[func(string) int]() }},
	}

	for _, tc := range funcTypes {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotPanics(t, tc.test, "T is a function type and should not panic")
		})
	}
}

func Test_AddAndGetHandlers(t *testing.T) {
	registry := internal.NewErrorHandlingRegistry[error]()
	errType := errors.New("test error")

	var calls []string

	handler1 := func(e error) {
		calls = append(calls, "handler1: "+e.Error())
	}
	handler2 := func(e error) {
		calls = append(calls, "handler2: "+e.Error())
	}

	registry.Add(handler1, errType)
	registry.Add(handler2, errType)

	handlers := registry.GetHandlers(errType)

	if len(handlers) != 2 {
		t.Errorf("expected 2 handlers, got %d", len(handlers))
	}

	// Call handlers and verify effect
	handlers[0](errType)
	handlers[1](errType)

	if len(calls) != 2 || calls[0] != "handler1: test error" || calls[1] != "handler2: test error" {
		t.Errorf("handler execution mismatch: %v", calls)
	}

	// Verify copy behavior
	handlers[0] = func(e error) { calls = append(calls, "modified") }
	newHandlers := registry.GetHandlers(errType)
	calls = []string{}
	newHandlers[0](errType)
	if calls[0] != "handler1: test error" {
		t.Errorf("modifying returned slice should not affect internal registry")
	}
}

func Test_Add_Multi(t *testing.T) {
	registry := internal.NewErrorHandlingRegistry[error]()
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	var results []string

	handler := func(e error) {
		results = append(results, "handled: "+e.Error())
	}

	registry.Add(handler, err1, err2)

	for _, err := range []error{err1, err2} {
		handlers := registry.GetHandlers(err)
		if len(handlers) != 1 {
			t.Errorf("expected 1 handler for %v, got %d", err, len(handlers))
			continue
		}
		handlers[0](err)
	}

	if len(results) != 2 || results[0] != "handled: err1" || results[1] != "handled: err2" {
		t.Errorf("unexpected handler effects: %v", results)
	}
}

func Test_EmptyRegistry(t *testing.T) {
	registry := internal.NewErrorHandlingRegistry[error]()
	err := errors.New("unregistered")

	handlers := registry.GetHandlers(err)
	if len(handlers) != 0 {
		t.Errorf("expected no handlers, got %d", len(handlers))
	}
}

func Test_GetHandlers_ReturnsCopy(t *testing.T) {
	registry := internal.NewErrorHandlingRegistry[error]()
	err := errors.New("copy check")

	var log []string

	originalHandler := func(e error) {
		log = append(log, "original handler: "+e.Error())
	}

	registry.Add(originalHandler, err)

	handlers := registry.GetHandlers(err)
	assert.Equal(t, 1, len(handlers))

	// Modify the returned slice
	handlers[0] = func(e error) {
		log = append(log, "tampered handler: "+e.Error())
	}

	// Fetch handlers again and call
	freshHandlers := registry.GetHandlers(err)
	freshHandlers[0](err)

	if len(log) != 1 || log[0] != "original handler: copy check" {
		t.Errorf("modifying returned handlers should not affect the registry; got log: %v", log)
	}
}

func Test_ThreadSafety(t *testing.T) {
	registry := internal.NewErrorHandlingRegistry[int]()
	err1 := errors.New("error1")
	err2 := errors.New("error2")

	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrently add handlers for different errors
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			registry.Add(func(int) {}, err1)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			registry.Add(func(int) {}, err2)
		}
	}()

	wg.Wait()

	// Verify correct counts
	assert.Equal(t, 100, len(registry.GetHandlers(err1)))
	assert.Equal(t, 100, len(registry.GetHandlers(err2)))
}

func Test_CompoundErrors(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errors.New("error2")
	err1err2 := errors.Join(err1, err2)

	reg := internal.NewErrorHandlingRegistry[int]()

	err1HandlerCalled := false
	err2HandlerCalled := false
	err1err2HandlerCalled := false

	reg.Add(func(int) { err1HandlerCalled = true }, err1)
	reg.Add(func(int) { err2HandlerCalled = true }, err2)
	reg.Add(func(int) { err1err2HandlerCalled = true }, err1err2)

	for _, h := range reg.GetHandlers(err1) {
		h(0)
	}
	assert.True(t, err1HandlerCalled)
	assert.False(t, err2HandlerCalled)
	assert.False(t, err1err2HandlerCalled)

	err1HandlerCalled = false
	for _, h := range reg.GetHandlers(err2) {
		h(0)
	}
	assert.False(t, err1HandlerCalled)
	assert.True(t, err2HandlerCalled)
	assert.False(t, err1err2HandlerCalled)

	err2HandlerCalled = false
	for _, h := range reg.GetHandlers(err1err2) {
		h(0)
	}
	assert.True(t, err1HandlerCalled)
	assert.True(t, err2HandlerCalled)
	assert.True(t, err1err2HandlerCalled)
}
