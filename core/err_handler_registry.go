package core

import (
	"errors"
	"sync"
)

type handlerType[T any] func(T)
type ErrorHandlingRegistry[T any] struct {
	m    sync.Mutex
	pool map[error][]handlerType[T]
}

// NewErrorHandlingRegistry creates a new ErrorHandlingRegistry of type func(T).
// It is thread safe
func NewErrorHandlingRegistry[T any]() *ErrorHandlingRegistry[T] {
	return &ErrorHandlingRegistry[T]{
		pool: make(map[error][]handlerType[T]),
	}
}

// Add registers a func(T) for one or more specific errors.
// The method associates the provided handler with each error in the errs slice,
// allowing the handler to be invoked when these errors occur.
//
// Multiple handlers can be registered for the same error, and they will be
// stored in the order they are added. The client is responsible for preventing
// duplicate handler registrations as this method does not check for duplicates.
//
// If the client provides a compound error, it will be decomposed into individual errors
// and processed separately.
//
// Example:
//   Input:    error(err1 + err2 + err3) : handler1
//   Behavior:
//     - err1 : handler1
//     - err2 : handler1
//     - err3 : handler1
//
// Each error is treated independently and matched with the same handler.

func (r *ErrorHandlingRegistry[T]) Add(handler handlerType[T], errs ...error) {
	r.m.Lock()
	defer r.m.Unlock()

	for _, err := range errs {
		r.pool[err] = append(r.pool[err], handler)
	}
}

// GetHandlers returns a deep copy list of the provided error handlers.
func (r *ErrorHandlingRegistry[T]) GetHandlers(err error) []handlerType[T] {
	r.m.Lock()
	defer r.m.Unlock()

	handlers := make([]handlerType[T], 0, 1)
	for e, h := range r.pool {
		if errors.Is(err, e) {
			handlers = append(handlers, h...)
		}
	}

	return handlers
}
