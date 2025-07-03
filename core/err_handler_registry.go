package core

import "sync"

type ErrorHandlingRegistry[ErrorHandler any] struct {
	m    sync.Mutex
	pool map[error][]ErrorHandler
}

// NewErrorHandlerRegistry creates a new ErrorHandlingRegistry with initialized pool.
// It is thread safe
func NewErrorHandlingRegistry[T any]() *ErrorHandlingRegistry[T] {
	return &ErrorHandlingRegistry[T]{
		pool: make(map[error][]T),
	}
}

// Add registers an ErrorHandler for a specific error.
func (e *ErrorHandlingRegistry[ErrorHandler]) Add(err error, handler ErrorHandler) {
	e.m.Lock()
	e.pool[err] = append(e.pool[err], handler)
	e.m.Unlock()
}

// AddMulti registers an ErrorHandler for a list of errors.
func (e *ErrorHandlingRegistry[ErrorHandler]) AddMulti(errs []error, handler ErrorHandler) {
	e.m.Lock()
	for _, err := range errs {
		e.pool[err] = append(e.pool[err], handler)
	}
	e.m.Unlock()
}

// GetCollection returns a deep copy list of the provided error handlers.
func (e *ErrorHandlingRegistry[ErrorHandler]) GetHandlers(err error) []ErrorHandler {
	e.m.Lock()
	defer e.m.Unlock()

	original := e.pool[err]
	copied := make([]ErrorHandler, len(original))
	copy(copied, original)
	return copied
}
