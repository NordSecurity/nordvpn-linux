package internal

import (
	"time"

	"golang.org/x/exp/slices"
)

func Find[T comparable](l []T, element T) *T {
	index := slices.Index(l, element)

	if index != -1 {
		return &l[index]
	}

	return nil
}

func Contains[T comparable](l []T, element T) bool {
	e := Find(l, element)
	return e != nil
}

// Retry calls fn up to attempts times, sleeping delay between calls.
// It returns nil as soon as fn returns nil; otherwise it returns the last error.
func Retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		// If this wasn’t the last attempt, wait before retrying
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return err
}
