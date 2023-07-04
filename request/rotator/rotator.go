// Package rotator is responsible for api request transport rotation.
package rotator

import (
	"fmt"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/request"
)

// Rotator handles the rotations for type T. It rotates the active instance of T when needed.
// The selected element is changed among the available elements in the elements slice.
// Thread safe.
type Rotator[T any] struct {
	elements []T
	index    int
	mu       sync.Mutex
}

func NewRotator[T any](elements []T) (*Rotator[T], error) {
	if len(elements) == 0 {
		return nil, fmt.Errorf("cannot create rotator with no elements")
	}
	return &Rotator[T]{
		elements: elements,
	}, nil
}

// Rotate changes the index to the current active element and returns it
func (r *Rotator[T]) Rotate() (T, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	lastElement := r.index+1 >= len(r.elements)
	if lastElement {
		return r.elements[r.index], request.ErrNothingMoreToRotate
	}
	r.index++
	return r.elements[r.index], nil
}

// Restart sets the active element to the first one available in the slice
func (r *Rotator[T]) Restart() T {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.elements) > 0 {
		r.index = 0
	}
	return r.elements[0]
}
