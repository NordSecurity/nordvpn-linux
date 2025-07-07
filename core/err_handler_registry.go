package core

import (
	"reflect"
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

// Add registers an ErrHandler for a specific error.
func (e *ErrorHandlingRegistry[ErrHandler]) Add(err error, handler ErrHandler) {
	e.m.Lock()
	e.pool[err] = append(e.pool[err], handler)
	e.m.Unlock()
}

// AddMulti registers an ErrHandler for a list of errors.
func (e *ErrorHandlingRegistry[ErrHandler]) AddMulti(errs []error, handler ErrHandler) {
	e.m.Lock()
	for _, err := range errs {
		e.pool[err] = append(e.pool[err], handler)
	}
	e.m.Unlock()
}

// GetHandlers returns a deep copy list of the provided error handlers.
func (e *ErrorHandlingRegistry[ErrHandler]) GetHandlers(err error) []ErrHandler {
	e.m.Lock()
	defer e.m.Unlock()

	original := e.pool[err]
	copied := make([]ErrHandler, len(original))
	for k, v := range e.pool[err] {
		copied[k] = v
	}

	return copied
}
