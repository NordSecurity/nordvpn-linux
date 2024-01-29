package internal

import (
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
