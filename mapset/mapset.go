// Package mapset provides a generic, map-backed set of comparable elements.
package mapset

// Set is a collection of unique comparable elements. Implementations are not
// safe for concurrent use; callers must provide their own synchronization.
type Set[T comparable] interface {
	// Add inserts element into the set. Adding an element that is already
	// present is a no-op.
	Add(element T)

	// Contains reports whether element is present in the set.
	Contains(element T) bool

	// IsEmpty reports whether the set contains no elements.
	IsEmpty() bool

	// Length returns the number of elements in the set.
	Length() int

	// Equal reports whether the set and other contain exactly the same
	// elements. Two empty sets are equal.
	Equal(other Set[T]) bool

	// IsSubset reports whether every element of the set is also present in
	// other. The empty set is a subset of every set, including itself.
	IsSubset(other Set[T]) bool
}

// set is the default map-backed implementation of Set. The map value type is
// struct{} so that entries consume no storage beyond their keys.
type set[T comparable] map[T]struct{}

// NewSet returns a new Set initialized with the given elements. Duplicate
// elements are stored only once. Calling NewSet with no arguments returns an
// empty set.
func NewSet[T comparable](elements ...T) Set[T] {
	s := make(set[T])
	for _, element := range elements {
		s[element] = struct{}{}
	}
	return &s
}

func (s *set[T]) Add(element T) {
	(*s)[element] = struct{}{}
}

func (s *set[T]) Contains(element T) bool {
	_, ok := (*s)[element]
	return ok
}

func (s *set[T]) IsEmpty() bool {
	return len(*s) == 0
}

func (s *set[T]) Length() int {
	return len(*s)
}

func (s *set[T]) Equal(other Set[T]) bool {
	if s.Length() != other.Length() {
		return false
	}

	for element := range *s {
		if !other.Contains(element) {
			return false
		}
	}

	return true
}

func (s *set[T]) IsSubset(other Set[T]) bool {
	if s.Length() > other.Length() {
		return false
	}

	for element := range *s {
		if !other.Contains(element) {
			return false
		}
	}

	return true
}
