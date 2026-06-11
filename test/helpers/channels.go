package helpers

import (
	"testing"
	"time"
)

func WaitWithTimeout[T any](t *testing.T, ch <-chan T, d time.Duration) T {
	t.Helper()

	var v T
	select {
	case v = <-ch:
	case <-time.After(d):
	}
	return v
}
