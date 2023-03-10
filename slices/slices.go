// Package slices builds upon golang.org/x/exp/slices.
package slices

import "golang.org/x/exp/slices"

// Contains just reexports golang.org/x/exp/slices.Contains to
// avoid package name collisions.
func Contains[E comparable](s []E, v E) bool {
	return slices.Contains(s, v)
}

// ContainsFunc works just like slices.Contains, but instead of an
// element, accepts a function as a second argument.
func ContainsFunc[E any](s []E, f func(E) bool) bool {
	return slices.IndexFunc(s, f) >= 0
}

// Delete just reexports golang.org/x/exp/slices.Delete to
// avoid package name collisions.
func Delete[S ~[]E, E any](s S, i, j int) []E {
	return slices.Delete(s, i, j)
}

// Filter returns a new slice with only the elements, for which
// f returned true.
func Filter[E any](s []E, f func(E) bool) []E {
	var ss []E
	for i, v := range s {
		if f(v) {
			ss = append(ss, s[i])
		}
	}

	return ss
}

// Contains just reexports golang.org/x/exp/slices.IndexFunc to
// avoid package name collisions.
func IndexFunc[E any](s []E, f func(E) bool) int {
	return slices.IndexFunc(s, f)
}
