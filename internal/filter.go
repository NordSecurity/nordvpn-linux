package internal

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
