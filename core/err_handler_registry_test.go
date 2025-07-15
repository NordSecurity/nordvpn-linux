package core_test

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/stretchr/testify/assert"
)

func Test_NewErrorHandlingRegistry_PanicWhenNotFunc(t *testing.T) {
	nonFuncTypes := []struct {
		name string
		test func()
	}{
		{"int", func() { _ = core.NewErrorHandlingRegistry[int]() }},
		{"string", func() { _ = core.NewErrorHandlingRegistry[string]() }},
		{"map[int]int", func() { _ = core.NewErrorHandlingRegistry[map[int]int]() }},
		{"bool", func() { _ = core.NewErrorHandlingRegistry[bool]() }},
		{"float64", func() { _ = core.NewErrorHandlingRegistry[float64]() }},
		{"complex128", func() { _ = core.NewErrorHandlingRegistry[complex128]() }},
		{"struct{}", func() { _ = core.NewErrorHandlingRegistry[struct{}]() }},
		{"[]int", func() { _ = core.NewErrorHandlingRegistry[[]int]() }},
		{"[3]string", func() { _ = core.NewErrorHandlingRegistry[[3]string]() }},
		{"chan int", func() { _ = core.NewErrorHandlingRegistry[chan int]() }},
		{"any", func() { _ = core.NewErrorHandlingRegistry[any]() }},
		{"*int", func() { _ = core.NewErrorHandlingRegistry[*int]() }},
		{"uintptr", func() { _ = core.NewErrorHandlingRegistry[uintptr]() }},
	}

	for _, tc := range nonFuncTypes {
		t.Run(tc.name, func(t *testing.T) {
			assert.Panics(t, tc.test, "T is not a function")
		})
	}
}

func Test_AddAndGetHandlers(t *testing.T) {
	type Handler func(error)

	registry := core.NewErrorHandlingRegistry[Handler]()
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
	type Handler func(error)

	registry := core.NewErrorHandlingRegistry[Handler]()
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
	type Handler func(error)

	registry := core.NewErrorHandlingRegistry[Handler]()
	err := errors.New("unregistered")

	handlers := registry.GetHandlers(err)
	if len(handlers) != 0 {
		t.Errorf("expected no handlers, got %d", len(handlers))
	}
}

func Test_GetHandlers_ReturnsCopy(t *testing.T) {
	type Handler func(error)

	registry := core.NewErrorHandlingRegistry[Handler]()
	err := errors.New("copy check")

	var log []string

	originalHandler := func(e error) {
		log = append(log, "original handler: "+e.Error())
	}

	registry.Add(originalHandler, err)

	handlers := registry.GetHandlers(err)
	assert.Equal(t, len(handlers), 1)

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
