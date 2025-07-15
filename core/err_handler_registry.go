package core

import (
	"reflect"
	"slices"
	"sync"
)

type ErrorHandlingRegistry[ErrHandler any] struct {
	m    sync.Mutex
	pool map[error][]ErrHandler
}

// NewErrorHandlingRegistry creates a new ErrorHandlingRegistry with initialized pool.
// It is thread safe
func NewErrorHandlingRegistry[T any]() *ErrorHandlingRegistry[T] {
	var zero T
	if reflect.TypeOf(zero).Kind() != reflect.Func {
		panic("ErrHandler must be a function")
	}

	return &ErrorHandlingRegistry[T]{
		pool: make(map[error][]T),
	}
}

// Add registers an ErrHandler for one or more specific errors.
// The method associates the provided handler with each error in the errs slice,
// allowing the handler to be invoked when these errors occur.
//
// Multiple handlers can be registered for the same error, and they will be
// stored in the order they are added. The client is responsible for preventing
// duplicate handler registrations as this method does not check for duplicates.
func (e *ErrorHandlingRegistry[ErrHandler]) Add(handler ErrHandler, errs ...error) {
	e.m.Lock()
	defer e.m.Unlock()

	for _, err := range errs {
		e.pool[err] = append(e.pool[err], handler)
	}
}

// GetHandlers returns a deep copy list of the provided error handlers.
func (e *ErrorHandlingRegistry[ErrHandler]) GetHandlers(err error) []ErrHandler {
	e.m.Lock()
	defer e.m.Unlock()

	return slices.Clone(e.pool[err])
}
