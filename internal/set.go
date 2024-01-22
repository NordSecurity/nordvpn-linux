package internal

import "fmt"

type Set[E comparable] map[E]struct{}

func NewSet[E comparable](vals ...E) Set[E] {
	s := Set[E]{}
	for _, v := range vals {
		s[v] = struct{}{}
	}
	return s
}

func (s Set[E]) Add(vals ...E) {
	for _, v := range vals {
		s[v] = struct{}{}
	}
}

func (s Set[E]) Remove(value E) {
	delete(s, value)
}

func (s Set[E]) Contains(v E) bool {
	_, ok := s[v]
	return ok
}

// check if the elements of the current Set are all contained in the parameter Set
func (s Set[E]) IsSubset(of Set[E]) bool {
	for k := range s {
		if !of.Contains(k) {
			return false
		}
	}
	return !s.Empty()
}

func (s Set[E]) ToSlice() []E {
	result := make([]E, 0, len(s))
	for v := range s {
		result = append(result, v)
	}
	return result
}

func (s Set[E]) String() string {
	return fmt.Sprintf("%v", s.ToSlice())
}

func (s Set[E]) Equal(other Set[E]) bool {
	if len(s) != len(other) {
		return false
	}

	for i := range other {
		if !s.Contains(i) {
			return false
		}
	}
	return true
}

func (s Set[E]) Empty() bool {
	return len(s) == 0
}
